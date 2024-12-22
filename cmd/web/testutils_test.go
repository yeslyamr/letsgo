package main

import (
	"bytes"
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/schema"
	"html"
	"io"
	"letsgo/internal/models/mocks"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"
)

func newTestApplication(t *testing.T) *application {
	// HTML TEMPLATING
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	// GORILLA DECODER
	formDecoder := schema.NewDecoder()
	formDecoder.IgnoreUnknownKeys(true)

	// SESSION MANAGER / REDIS STORE
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		errorLog:       log.New(io.Discard, "", 0),
		infoLog:        log.New(io.Discard, "", 0),
		users:          &mocks.UserModel{},
		snippets:       &mocks.SnippetModel{},
		sessionManager: sessionManager,
		templateCache:  templateCache,
		formDecoder:    formDecoder,
	}
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	// cookies for client
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar

	// redirect
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}

// CSRF
var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

func extractCSRFToken(t *testing.T, body string) string {
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}
	return html.UnescapeString(string(matches[1]))
}
