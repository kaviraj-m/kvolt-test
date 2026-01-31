package main

import (
	"log"
	"time"

	"test/internal/handler"

	"github.com/go-kvolt/kvolt"
	"github.com/go-kvolt/kvolt/context"
	"github.com/go-kvolt/kvolt/middleware"
	"github.com/go-kvolt/kvolt/pkg/cache"
	"github.com/go-kvolt/kvolt/pkg/queue"
	"github.com/go-kvolt/kvolt/pkg/scheduler"
	"github.com/go-kvolt/kvolt/pkg/swagger"
	"github.com/golang-jwt/jwt/v5"
)

// @title           KVolt Test API
// @version         1.0
// @description     This is a sample server using KVolt.
// @host            localhost:8080
// @BasePath        /
func main() {
	app := kvolt.New()

	app.Use(middleware.Logger())

	// ---------------------------------------------------------
	// Queue System Integration
	// ---------------------------------------------------------

	// 1. Initialize Queue (1000 buffer, 5 workers)
	q := queue.NewMemoryQueue(1000, 5)

	// 2. Register Handler
	q.Register("test_job", func(job queue.Job) error {
		log.Printf("[Worker] Processing Job: %s | Payload: %v", job.ID, job.Payload)
		// Simulate work
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	// 3. Start Queue
	q.Start()
	defer q.Stop()

	// 4. API to Dispatch
	app.POST("/queue", func(c *context.Context) error {
		type JobRequest struct {
			Message string `json:"message"`
		}
		var req JobRequest
		if err := c.Bind(&req); err != nil {
			return c.Status(400).JSON(400, map[string]string{"error": "Invalid JSON"})
		}

		if err := q.Push("test_job", req.Message); err != nil {
			return c.Status(503).JSON(503, map[string]string{"error": "Queue Full"})
		}

		return c.JSON(202, map[string]string{
			"status": "Job Dispatched",
			"info":   "Check server logs for worker output",
		})
	})
	// ---------------------------------------------------------

	// 1. Load Templates
	app.LoadHTMLGlob("views/*")

	// 2. HTML Route
	app.GET("/html", func(c *context.Context) error {
		return c.RenderHTML(200, "index.html", map[string]string{
			"title": "KVolt Demo",
			"user":  "Kaviraj",
		})
	})

	// 3. Login Endpoint
	app.POST("/login", func(c *context.Context) error {
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.Status(400).JSON(400, map[string]string{"error": "Invalid request"})
		}

		// Mock user validation
		if req.Username != "admin" || req.Password != "password" {
			return c.Status(401).JSON(401, map[string]string{"error": "Invalid credentials"})
		}

		// Create JWT Token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user": "admin",
			"exp":  time.Now().Add(time.Hour * 72).Unix(),
		})

		// Sign and get the complete encoded token as a string using the secret
		tokenString, err := token.SignedString([]byte("secret"))
		if err != nil {
			return c.Status(500).JSON(500, map[string]string{"error": "Could not generate token"})
		}

		return c.JSON(200, map[string]string{
			"token": tokenString,
		})
	})

	// 4. Protected Route with JWT Middleware
	// Configure JWT Middleware
	jwtMiddleware := middleware.JWT(middleware.JWTConfig{
		SigningKey:  "secret",
		TokenLookup: "header:Authorization", // Default, but explicitly showing usage
		AuthScheme:  "Bearer",               // Default
	})

	// Use a group for protected routes
	protectedBase := app.Group("/protected")
	protectedBase.Use(jwtMiddleware)

	// Registers GET /protected/
	protectedBase.GET("/", func(c *context.Context) error {
		claims := c.MustGet("user").(jwt.MapClaims)
		return c.JSON(200, map[string]interface{}{
			"message": "Welcome to the protected area!",
			"user":    claims["user"],
		})
	})

	// Serve Static Files
	app.Static("/static", "./public")

	// User Routes
	app.GET("/users", handler.GetUsers)
	app.GET("/users/:id", handler.GetUser)
	app.PUT("/users/:id", handler.UpdateUser)
	app.DELETE("/users/:id", handler.DeleteUser)
	app.POST("/users", handler.CreateUser)

	// Post Routes
	app.GET("/posts", handler.GetPosts)
	app.POST("/posts", handler.CreatePost)

	// Misc Routes
	app.GET("/health", handler.GetHealth)
	app.GET("/version", handler.GetVersion)

	// Swagger Route
	// Swagger Route
	doc, _ := swagger.ReadDoc() // Read from "swagger" instance by default

	app.GET("/swagger/*any", swagger.Handler(swagger.Config{
		SpecJSON:       doc,
		RoutesProvider: swagger.Adapter(app),
		Title:          "KVolt Auto-Docs Demo",
		Host:           "localhost:8080",
	}))

	// ---------------------------------------------------------
	// Caching System Integration
	// ---------------------------------------------------------

	// Initialize Sharded Cache (Cleanup every 1 minute)
	myCache := cache.NewMemoryStore(1 * time.Minute)

	app.GET("/cache/:key", func(c *context.Context) error {
		key := c.Params.Get("key")

		// 1. Try to get from cache
		if val, err := myCache.Get(key); err == nil {
			return c.JSON(200, map[string]interface{}{
				"source": "cache",
				"data":   val,
			})
		}

		// 2. Simulate heavy work
		data := "Value for " + key + " (Generated at " + time.Now().Format(time.RFC3339) + ")"
		time.Sleep(500 * time.Millisecond)

		// 3. Store in cache for 30 seconds
		myCache.Set(key, data, 30*time.Second)

		return c.JSON(200, map[string]interface{}{
			"source": "generated",
			"data":   data,
		})
	})
	// ---------------------------------------------------------

	// ---------------------------------------------------------
	// Task Scheduler Integration
	// ---------------------------------------------------------
	s := scheduler.New()

	// Run every 10 seconds
	s.Add("@every 10s", func() {
		log.Println("⏰ [Cron] Running scheduled cleanup task...")
	})

	s.Start()
	defer s.Stop()
	// ---------------------------------------------------------

	app.Run(":8080")
}
