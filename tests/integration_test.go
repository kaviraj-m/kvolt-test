package tests

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-kvolt/kvolt"
	"github.com/go-kvolt/kvolt/context"
	"github.com/go-kvolt/kvolt/middleware"
	kvtest "github.com/go-kvolt/kvolt/pkg/test"
	"github.com/golang-jwt/jwt/v5"
)

func setupApp() *kvolt.Engine {
	app := kvolt.New()
	app.GET("/health", func(c *context.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	app.POST("/echo", func(c *context.Context) error {
		// In a real app we'd decode properly
		return c.JSON(201, map[string]string{"echo": "received"})
	})
	return app
}

func TestHealth(t *testing.T) {
	app := setupApp()
	test := kvtest.New(t, app)

	test.GET("/health").
		Do().
		ExpectStatus(200).
		ExpectBodyContains("ok").
		ExpectJSON(map[string]string{"status": "ok"})
}

func TestEcho(t *testing.T) {
	app := setupApp()
	test := kvtest.New(t, app)

	test.POST("/echo").
		WithJSON(map[string]string{"msg": "hello"}).
		Do().
		ExpectStatus(201).
		ExpectHeader("Content-Type", "application/json")
}

type UserRequest struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"gte=18"`
}

func TestBind(t *testing.T) {
	app := kvolt.New()
	app.POST("/bind", func(c *context.Context) error {
		var req UserRequest
		if err := c.Bind(&req); err != nil {
			return c.Status(400).JSON(400, map[string]string{"error": err.Error()})
		}
		return c.JSON(200, req)
	})

	test := kvtest.New(t, app)

	// Case 1: Valid
	test.POST("/bind").
		WithJSON(UserRequest{Name: "Alice", Age: 20}).
		Do().
		ExpectStatus(200).
		ExpectBodyContains("Alice")

	// Case 2: Invalid (Age < 18)
	test.POST("/bind").
		WithJSON(UserRequest{Name: "Bob", Age: 10}).
		Do().
		ExpectStatus(400).
		ExpectBodyContains("failed on the 'gte' tag")
}

func TestContextKeys(t *testing.T) {
	app := kvolt.New()

	// Simulate Auth Middleware
	app.Use(func(c *context.Context) error {
		c.Set("user_id", "12345")
		c.Next()
		return nil
	})

	app.GET("/profile", func(c *context.Context) error {
		// Get data from middleware
		userID := c.MustGet("user_id").(string)
		return c.JSON(200, map[string]string{"id": userID})
	})

	test := kvtest.New(t, app)

	test.GET("/profile").
		Do().
		ExpectStatus(200).
		ExpectJSON(map[string]string{"id": "12345"})
}

func TestFileUpload(t *testing.T) {
	app := kvolt.New()
	app.POST("/upload", func(c *context.Context) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).String(400, "Bad Request")
		}

		// Ensure directory exists
		os.Mkdir("uploads", 0755)
		dst := "uploads/" + file.Filename
		if err := c.SaveUploadedFile(file, dst); err != nil {
			return c.Status(500).String(500, "Internal Server Error")
		}
		return c.String(200, "Uploaded: "+file.Filename)
	})

	// Manually create Multipart Request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("hello world"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "Uploaded: test.txt" {
		t.Errorf("Expected 'Uploaded: test.txt', got '%s'", w.Body.String())
	}

	// Clean up
	os.RemoveAll("uploads")
}

func TestHTMLRendering(t *testing.T) {
	app := kvolt.New()

	// Create dummy view
	cwd, _ := os.Getwd()
	t.Logf("Current Working Directory: %s", cwd)

	dir := "views"
	if err := os.Mkdir(dir, 0755); err != nil && !os.IsExist(err) {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	filePath := dir + "/index.html"
	if err := os.WriteFile(filePath, []byte("<h1>Hello {{.user}}</h1>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Force Absolute Path for Glob
	absPattern := filepath.Join(cwd, dir, "*")
	t.Logf("Loading Glob: %s", absPattern)

	app.LoadHTMLGlob(absPattern)

	app.GET("/html", func(c *context.Context) error {
		return c.RenderHTML(200, "index.html", map[string]string{"user": "kaviraj"})
	})

	test := kvtest.New(t, app)
	test.GET("/html").
		Do().
		ExpectStatus(200).
		ExpectBody("<h1>Hello kaviraj</h1>")
}

func TestJWTMiddleware(t *testing.T) {
	app := kvolt.New()
	secret := "super-secret-key"

	// Use JWT Middleware
	app.Use(middleware.JWT(middleware.JWTConfig{
		SigningKey: secret,
	}))

	app.GET("/protected", func(c *context.Context) error {
		claims := c.MustGet("user").(jwt.MapClaims)
		return c.JSON(200, map[string]string{"user": claims["sub"].(string)})
	})

	test := kvtest.New(t, app)

	// Case 1: No Token
	test.GET("/protected").
		Do().
		ExpectStatus(401).
		ExpectJSON(map[string]string{"error": "missing auth header"})

	// Case 2: Invalid Token
	test.GET("/protected").
		WithHeader("Authorization", "Bearer invalid").
		Do().
		ExpectStatus(401)

	// Case 3: Valid Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "kaviraj",
	})
	signed, _ := token.SignedString([]byte(secret))

	test.GET("/protected").
		WithHeader("Authorization", "Bearer "+signed).
		Do().
		ExpectStatus(200).
		ExpectJSON(map[string]string{"user": "kaviraj"})
}
