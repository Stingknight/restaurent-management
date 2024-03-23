package middleware

import (

	"net/http"

	"github.com/Stingknight/restaurentManagement/helpers"
	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc{
	return func (ctx *gin.Context){
		var tokenString string  = ctx.Request.Header.Get("token")

		if tokenString==""{
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims,err := helpers.ValidateToken(tokenString)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	
		ctx.Set("user_id", claims["sub"].(string))
		ctx.Next()
	}	
}