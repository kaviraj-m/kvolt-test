package handler

import (
	"time"

	"github.com/go-kvolt/kvolt/context"
)

// GetHealth godoc
// @Summary      Get system health
// @Description  Get the health status of the system
// @Tags         misc
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func GetHealth(c *context.Context) error {
	return c.JSON(200, map[string]string{
		"status": "ok",
		"uptime": time.Since(time.Now()).String(), // simplistic uptime for demo
	})
}

// GetVersion godoc
// @Summary      Get API version
// @Description  Get the current version of the API
// @Tags         misc
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /version [get]
func GetVersion(c *context.Context) error {
	return c.JSON(200, map[string]string{
		"version": "1.0.0",
		"build":   "dev",
	})
}
