package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/metadata"
)

const (
	BearerScheme string = "Bearer "
)

type authMiddleware struct {
}

func NewAuthMiddleware() *authMiddleware {
	return &authMiddleware{}
}

func (m *authMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if len(token) > 0 {
			ctx := metadata.Set(c.Request.Context(), "Authorization", BearerScheme+token)
			c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
