package main

import (
	"log"

	"test/internal/handler"

	"github.com/go-kvolt/kvolt"
	"github.com/go-kvolt/kvolt/context"
	"github.com/go-kvolt/kvolt/middleware"
	"github.com/go-kvolt/kvolt/pkg/swagger"
)

// @title           KVolt Test API
// @version         1.0
// @description     This is a sample server using KVolt.
// @host            localhost:8080
// @BasePath        /
func main() {
	app := kvolt.New()

	app.Use(middleware.Logger())

	app.GET("/", func(c *context.Context) error {
		return c.JSON(200, map[string]string{"msg": "Hello KVolt"})
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

	log.Println("Server starting on :8080")
	app.Run(":8080")
}
