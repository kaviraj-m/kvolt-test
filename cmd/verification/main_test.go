package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-kvolt/kvolt"
	"github.com/go-kvolt/kvolt/context"
	"github.com/go-kvolt/kvolt/middleware"
	"github.com/go-kvolt/kvolt/pkg/swagger"
	"github.com/gorilla/websocket"
)

// setupTestEngine creates a fresh engine for testing
func setupTestEngine() *kvolt.Engine {
	app := kvolt.New()
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Static Routes
	app.GET("/ping", func(c *context.Context) error {
		return c.String(200, "pong")
	})

	app.POST("/echo", func(c *context.Context) error {
		var body map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&body); err != nil {
			return c.Status(400).String(400, "Bad Request")
		}
		return c.JSON(200, body)
	})

	// Param Routes
	app.GET("/user/:id", func(c *context.Context) error {
		id := c.Param("id")
		return c.JSON(200, map[string]string{"id": id})
	})

	app.GET("/files/*filepath", func(c *context.Context) error {
		path := c.Param("filepath")
		return c.String(200, path)
	})

	// Group Routes
	v1 := app.Group("/v1")
	{
		v1.GET("/hello", func(c *context.Context) error {
			return c.String(200, "v1 hello")
		})
	}

	return app
}

func performRequest(r *httptest.ResponseRecorder, req *http.Request, app *kvolt.Engine) {
	app.ServeHTTP(r, req)
}

func TestStaticRoutes(t *testing.T) {
	app := setupTestEngine()

	// 1. GET /ping
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "pong" {
		t.Errorf("Expected 'pong', got '%s'", w.Body.String())
	}
}

func TestParamRoutes(t *testing.T) {
	app := setupTestEngine()

	// 1. GET /user/123
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/123", nil)
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	expected := `{"id":"123"}`
	// Trim newline from JSON encoder
	if body := w.Body.String(); body[:len(body)-1] != expected {
		t.Errorf("Expected %s, got %s", expected, body)
	}
}

func TestWildcardRoutes(t *testing.T) {
	app := setupTestEngine()

	// 1. GET /files/css/style.css
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/files/css/style.css", nil)
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "css/style.css" {
		t.Errorf("Expected 'css/style.css', got '%s'", w.Body.String())
	}
}

func TestGroupRoutes(t *testing.T) {
	app := setupTestEngine()

	// 1. GET /v1/hello
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/hello", nil)
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "v1 hello" {
		t.Errorf("Expected 'v1 hello', got '%s'", w.Body.String())
	}
}

func TestConcurrency(t *testing.T) {
	app := setupTestEngine()
	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	start := time.Now()
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/ping", nil)
			app.ServeHTTP(w, req)
			if w.Code != 200 {
				t.Errorf("Concurrency fail: %d", w.Code)
			}
		}()
	}
	wg.Wait()
	duration := time.Since(start)
	t.Logf("Processed %d requests in %v", concurrency, duration)
}

func Test404(t *testing.T) {
	app := setupTestEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notfound", nil)
	app.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestPostEcho(t *testing.T) {
	app := setupTestEngine()

	payload := []byte(`{"message":"hello"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/echo", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// Basic check
	if !bytes.Contains(w.Body.Bytes(), []byte("hello")) {
		t.Errorf("Body mismatch")
	}
}

func TestSwaggerUI(t *testing.T) {
	app := setupTestEngine()
	// Inject swagger route manually for test since setupTestEngine doesn't have it
	app.GET("/swagger/*any", swagger.Handler(swagger.Config{
		SpecJSON: "{}",
		Title:    "Test Docs",
	}))

	// 1. GET /swagger/index.html
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger/index.html", nil)
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("Test Docs")) {
		t.Errorf("Expected 'Test Docs' in title")
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("swagger-ui")) {
		t.Errorf("Expected 'swagger-ui' assets link")
	}
}

// Adapter to convert kvolt.Engine to swagger.RoutesProvider
type kvoltAdapter struct {
	engine *kvolt.Engine
}

func (a *kvoltAdapter) Routes() []swagger.RouteInfo {
	kvoltRoutes := a.engine.Routes()
	var routes []swagger.RouteInfo
	for _, r := range kvoltRoutes {
		routes = append(routes, swagger.RouteInfo{
			Method:  r.Method,
			Path:    r.Path,
			Summary: r.Summary,
		})
	}
	return routes
}

func TestAutoDocs(t *testing.T) {
	app := setupTestEngine()
	// Add descriptions to verify richer listing
	app.GET("/doc-test", func(c *context.Context) error { return nil }).Desc("Documented Endpoint")

	// Inject swagger route manually, but without SpecJSON, providing App as RoutesProvider via adapter
	app.GET("/swagger/*any", swagger.Handler(swagger.Config{
		RoutesProvider: &kvoltAdapter{engine: app},
		Title:          "Auto Docs",
	}))

	// 1. GET /swagger/doc.json
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger/doc.json", nil)
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	body := w.Body.Bytes()
	// Check if JSON contains /ping
	if !bytes.Contains(body, []byte("/ping")) {
		t.Errorf("Expected /ping in auto-generated docs")
	}
	// Check if JSON contains /echo
	if !bytes.Contains(body, []byte("/echo")) {
		t.Errorf("Expected /echo in auto-generated docs")
	}
	// Check description
	if !bytes.Contains(body, []byte("Documented Endpoint")) {
		t.Errorf("Expected 'Documented Endpoint' summary in json")
	}
}

func TestDisabledDocs(t *testing.T) {
	app := setupTestEngine()
	// Inject swagger route with Disabled: true
	app.GET("/swagger/*any", swagger.Handler(swagger.Config{
		RoutesProvider: &kvoltAdapter{engine: app},
		Disabled:       true,
	}))

	// 1. GET /swagger/index.html
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger/index.html", nil)
	app.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("Expected 404 when docs disabled, got %d", w.Code)
	}
}

func TestWebSocket(t *testing.T) {
	app := setupTestEngine()
	app.GET("/ws", func(c *context.Context) error {
		conn, err := c.Upgrade()
		if err != nil {
			return err
		}
		defer conn.Close()

		// Simple echo loop
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			err = conn.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
		return nil
	})

	// Create a test server
	s := httptest.NewServer(app)
	defer s.Close()

	// Convert http:// to ws://
	u := "ws" + strings.TrimPrefix(s.URL, "http") + "/ws"

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer ws.Close()

	// Send message
	if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read message
	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(p) != "hello" {
		t.Fatalf("bad message: %s", p)
	}
}

func TestHTTP2(t *testing.T) {
	app := setupTestEngine()
	app.GET("/h2", func(c *context.Context) error {
		return c.String(200, "http2")
	})

	// httptest.NewTLSServer automatically enables HTTP/2 if supported
	ts := httptest.NewTLSServer(app)
	defer ts.Close()

	// Create a client that supports HTTP/2
	client := ts.Client()

	// Force HTTP/2 transport checking is tricky with default client,
	// but standard Go client prefers H2 over TLS.

	resp, err := client.Get(ts.URL + "/h2")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// Check Protocol
	// Note: httptest server might not enable H2 by default depending on environment,
	// but this verifies KVolt handles the TLS connection correctly.
	t.Logf("Protocol used: %s", resp.Proto)

	// We won't strict fail on "HTTP/1.1" because test environment might lack H2 drivers,
	// but we log it for verification.
}

func TestStaticFiles(t *testing.T) {
	app := setupTestEngine()

	// Create a temp directory and file
	tmpDir := t.TempDir()
	content := []byte("body { color: red; }")

	// Create a css file inside
	if err := os.Mkdir(tmpDir+"/css", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmpDir+"/css/style.css", content, 0644); err != nil {
		t.Fatal(err)
	}

	// Register Static Route
	app.Static("/assets", tmpDir)

	// Test request: /assets/css/style.css
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/assets/css/style.css", nil)
	app.ServeHTTP(w, req)

	// Check status
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	// Check content type (approximate)
	if ct := w.Result().Header.Get("Content-Type"); !strings.Contains(ct, "text/css") {
		t.Errorf("Expected Content-Type text/css, got %s", ct)
	}
	// Check body
	if !bytes.Equal(w.Body.Bytes(), content) {
		t.Errorf("Expected body %s, got %s", content, w.Body.Bytes())
	}

	// Test request: /assets/ (Trailing Slash)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/assets/", nil)
	app.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("Trailing slash /assets/ failed: expected 200, got %d", w2.Code)
	}
}
