package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"clock/internal/config"
)

// NewJWTConfig 创建JWT中间件配置
func NewJWTConfig(cfg *config.AuthConfig) middleware.JWTConfig {
	jwtConfig := middleware.DefaultJWTConfig
	jwtConfig.SigningKey = []byte(cfg.JWTSecret)
	jwtConfig.TokenLookup = "header:token:"
	jwtConfig.AuthScheme = ""

	// 跳过不需要认证的路径
	skipPaths := []string{
		"/v1/login",
		"/v1/task/status", // WebSocket
		"webapp",
		"js",
		"css",
	}

	jwtConfig.Skipper = func(c echo.Context) bool {
		uri := c.Request().RequestURI

		for _, path := range skipPaths {
			if strings.Contains(uri, path) {
				return true
			}
		}

		return false
	}

	return jwtConfig
}
