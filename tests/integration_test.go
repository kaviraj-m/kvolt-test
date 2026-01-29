package tests

import (
	"testing"

	"github.com/go-kvolt/kvolt"
	"github.com/go-kvolt/kvolt/context"
	kvtest "github.com/go-kvolt/kvolt/pkg/test"
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
