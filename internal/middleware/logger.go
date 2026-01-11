package middleware

import (
	"time"

	"github.com/labstack/echo/v4"

	"clock/internal/logger"
)

// Logger 返回统一的日志中间件
func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			latency := time.Since(start)
			req := c.Request()
			res := c.Response()

			// 使用统一的 logger 输出
			logger.Infof("[http] %d %s %s %s",
				res.Status,
				req.Method,
				req.RequestURI,
				latency.String(),
			)

			return err
		}
	}
}
