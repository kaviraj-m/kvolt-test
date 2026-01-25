package handler

import (
	"encoding/json"

	"github.com/go-kvolt/kvolt/context"
)

type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var posts = []Post{
	{ID: 1, Title: "First Post"},
	{ID: 2, Title: "KVolt is Awesome"},
}

// GetPosts godoc
// @Summary      Get all posts
// @Description  Get a list of all posts
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {array}   Post
// @Router       /posts [get]
func GetPosts(c *context.Context) error {
	return c.JSON(200, posts)
}

// CreatePost godoc
// @Summary      Create a post
// @Description  Create a new post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param        post  body      Post  true  "Post"
// @Success      201   {object}  Post
// @Router       /posts [post]
func CreatePost(c *context.Context) error {
	var post Post
	if err := json.NewDecoder(c.Request.Body).Decode(&post); err != nil {
		return c.JSON(400, map[string]string{"error": err.Error()})
	}
	post.ID = len(posts) + 1
	posts = append(posts, post)
	return c.JSON(201, post)
}
