package middleware

import (
	"ValorantAPI/internal/http/response"
	"ValorantAPI/internal/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("access_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Response{
				Error: &response.ErrorResponse{Message: "Authorization cookie is required"},
			})
			c.Abort()
			return
		}

		claims, err := jwt.Parse(tokenStr, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Response{
				Error: &response.ErrorResponse{Message: "Invalid or expired token"},
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}
