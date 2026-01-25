package handler

import (
	"encoding/json"
	"strconv"

	"test/internal/model"

	"github.com/go-kvolt/kvolt/context"
)

// In-memory store for demonstration
var users = []model.User{
	{ID: 1, Name: "John Doe", Email: "john@example.com"},
	{ID: 2, Name: "Jane Doe", Email: "jane@example.com"},
}

// GetUsers godoc
// @Summary      Get all users
// @Description  Get a list of all users
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {array}   model.User
// @Router       /users [get]
func GetUsers(c *context.Context) error {
	return c.JSON(200, users)
}

// GetUser godoc
// @Summary      Get a user
// @Description  Get a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  model.User
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [get]
func GetUser(c *context.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(400, map[string]string{"error": "Invalid ID in path"})
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid ID format"})
	}

	for _, u := range users {
		if u.ID == id {
			return c.JSON(200, u)
		}
	}
	return c.JSON(404, map[string]string{"message": "User not found"})
}

// CreateUser godoc
// @Summary      Create a user
// @Description  Create a new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      model.User  true  "User"
// @Success      201   {object}  model.User
// @Router       /users [post]
func CreateUser(c *context.Context) error {
	var user model.User
	// Use standard json decoder
	if err := json.NewDecoder(c.Request.Body).Decode(&user); err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}
	user.ID = len(users) + 1
	users = append(users, user)
	return c.JSON(201, user)
}

// UpdateUser godoc
// @Summary      Update a user
// @Description  Update an existing user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int         true  "User ID"
// @Param        user  body      model.User  true  "User"
// @Success      200   {object}  model.User
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Router       /users/{id} [put]
func UpdateUser(c *context.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid ID format"})
	}

	var updatedUser model.User
	if err := json.NewDecoder(c.Request.Body).Decode(&updatedUser); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid request body"})
	}

	for i, u := range users {
		if u.ID == id {
			// Update fields
			users[i].Name = updatedUser.Name
			users[i].Email = updatedUser.Email
			// Return updated user (with ID preserved)
			users[i].ID = id // Ensure ID doesn't change
			return c.JSON(200, users[i])
		}
	}
	return c.JSON(404, map[string]string{"message": "User not found"})
}

// DeleteUser godoc
// @Summary      Delete a user
// @Description  Delete a user by ID
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [delete]
func DeleteUser(c *context.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid ID format"})
	}

	for i, u := range users {
		if u.ID == id {
			// Remove from slice
			users = append(users[:i], users[i+1:]...)
			return c.Status(204).String(204, "")
		}
	}
	return c.JSON(404, map[string]string{"message": "User not found"})
}
