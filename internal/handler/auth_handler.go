package handler

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"

	"clock/internal/config"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	cfg *config.AuthConfig
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(cfg *config.AuthConfig) *AuthHandler {
	return &AuthHandler{
		cfg: cfg,
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	User     string `json:"user"`
	Password string `json:"pwd"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
}

// Login 用户登录
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return BadRequest(c, "invalid request body")
	}

	// 验证用户名密码
	if req.User != h.cfg.User || req.Password != h.cfg.Password {
		return Unauthorized(c, "invalid username or password")
	}

	// 生成JWT token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = req.User
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	t, err := token.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return InternalError(c, "failed to generate token")
	}

	return OK(c, LoginResponse{Token: t})
}
