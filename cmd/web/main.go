package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/schema"
	_ "github.com/jackc/pgx/v5/stdlib"
	"html/template"
	"letsgo/internal/models"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       models.SnippetModelInterface
	users          models.UserModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *schema.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	// CLI CONFIG FLAGS
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "user=yeslyamr password= host=localhost port=5432 database=snippetbox", "PostgreSQL dsn")
	flag.Parse()

	// LOGS
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// DB
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	// HTML TEMPLATING
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// GORILLA DECODER
	formDecoder := schema.NewDecoder()
	formDecoder.IgnoreUnknownKeys(true)

	// SESSION MANAGER / REDIS STORE
	pool := &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", "localhost:6379")
		},
	}

	sessionManager := scs.New()
	sessionManager.Store = redisstore.New(pool)
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// APPLICATION
	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
		TLSConfig: &tls.Config{
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		},
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
