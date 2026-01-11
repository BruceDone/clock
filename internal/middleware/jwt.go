package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"clock/internal/config"
)

// NewJWTConfig 创建JWT中间件配置
func NewJWTConfig(cfg *config.AuthConfig) middleware.JWTConfig {
	jwtConfig := middleware.DefaultJWTConfig
	jwtConfig.SigningKey = []byte(cfg.JWTSecret)
	jwtConfig.TokenLookup = "header:token:,cookie:token,query:token"
	jwtConfig.AuthScheme = ""

	// 跳过不需要认证的路径（注意：JWT 中间件只作用于 /v1 路由组）
	jwtConfig.Skipper = func(c echo.Context) bool {
		path := c.Request().URL.Path
		return path == "/v1/login"
	}

	return jwtConfig
}
