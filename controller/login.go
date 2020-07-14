package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"clock/param"
)

// 登录后回token
func Login(c echo.Context) error {
	var user param.User

	resp := param.ApiResponse{
		Code: 200,
		Msg:  "success",
		Data: nil,
	}

	if err := c.Bind(&user); err != nil {
		resp.Msg = fmt.Sprintf("[login] error to get the user param with: %v", err)
		logrus.Error(resp.Msg)

		return c.JSON(http.StatusBadRequest, resp)
	}

	if param.WebUser != user.UserName || param.WebPwd != user.UserPwd {
		resp.Msg = fmt.Sprintf("[login] invalidate pwd or uid ,please check")
		logrus.Error(resp.Msg + ":uid is " + user.UserName)

		return c.JSON(http.StatusUnauthorized, resp)
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = user.UserName
	claims["exp"] = time.Now().Add(time.Duration(7*24) * time.Hour).Unix() // 过期时间

	t, err := token.SignedString([]byte(param.WebJwt))
	if err != nil {
		return err
	}

	resp.Data = t

	return c.JSON(http.StatusOK, resp)
}
