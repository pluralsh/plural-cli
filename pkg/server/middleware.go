package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		for _, err := range c.Errors {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}
