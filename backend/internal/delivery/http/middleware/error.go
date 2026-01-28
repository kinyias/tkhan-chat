package middleware

import (
	"net/http"

	"backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ErrorHandler handles panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				utils.ErrorResponse(c, http.StatusInternalServerError, "internal server error", nil)
			}
		}()
		c.Next()
	}
}
