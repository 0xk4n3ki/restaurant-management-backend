package middleware

import (
	"net/http"

	"github.com/0xk4n3ki/restaurant-management-backend/helpers"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientToken := ctx.GetHeader("token")
		if clientToken == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			ctx.Abort()
			return
		}

		claims, msg := helpers.ValidateToken(clientToken)
		if msg != "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			ctx.Abort()
			return
		}

		ctx.Set("email", claims.Email)
		ctx.Set("first_name", claims.First_name)
		ctx.Set("last_name", claims.Last_name)
		ctx.Set("uid", claims.Uid)

		ctx.Next()
	}
}
