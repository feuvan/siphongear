package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sunshow/siphongear/internal/templates"
)

func (s *Server) handleListTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, templates.List())
}

func (s *Server) handleGetTemplate(c *gin.Context) {
	t, ok := templates.Get(c.Param("name"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}
	c.JSON(http.StatusOK, t)
}
