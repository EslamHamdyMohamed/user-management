package middleware

import (
	"net/http"
	"user-management/pkg/logger"

	"github.com/gin-gonic/gin"
)

func Recovery(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				log.Error().
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Interface("error", err).
					Msg("Recovered from panic")

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_server_error",
					"message": "An unexpected error occurred",
				})

				c.Abort()
			}
		}()

		c.Next()
	}
}
