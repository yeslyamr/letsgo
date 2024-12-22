package models

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("pgx", "user=test_web password=pass host=localhost port=5432 database=test_snippetbox")
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}
		_ = db.Close()
	})
	return db
}
