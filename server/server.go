package server

import (
	"clock/param"
	"net/http"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"clock/controller"
)

func addApi(e *echo.Echo) {
	// 增加cors 中间件
	e.Use(middleware.CORS())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "[${status}] - ${method} - ${uri} - ${query} - ${form}\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}))
	e.Use(middleware.Recover())

	// 使用jwt token验证
	v1 := e.Group("/v1")
	{
		v1.Use(middleware.JWTWithConfig(createJWTConfig()))
		t := v1.Group("/task")
		{
			t.GET("", controller.GetTasks)
			t.GET("/:tid", controller.GetTask)
			t.PUT("", controller.PutTask)
			t.GET("/run", controller.RunTask)
			t.DELETE("/:tid", controller.DeleteTask)
			t.GET("/status", controller.GetTaskStatus)
		}

		l := v1.Group("/log")
		{
			l.GET("", controller.GetLogs)
			l.DELETE("", controller.DeleteLogs)
		}

		u := v1.Group("/login")
		{
			u.POST("", controller.Login)
		}

		r := v1.Group("/relation")
		{
			r.GET("", controller.GetRelations)
			r.POST("", controller.AddRelation)
			r.DELETE("/:rid", controller.DeleteRelation)
		}

		n := v1.Group("/node")
		{
			n.PUT("", controller.PutNodes)
		}

		c := v1.Group("/container")
		{
			c.GET("", controller.GetContainers)
			c.GET("/:cid", controller.GetContainer)
			c.PUT("", controller.PutContainer)
			c.GET("/run", controller.RunContainer)
			c.DELETE("/:cid", controller.DeleteContainer)
		}

		// 消息中心
		m := v1.Group("/message")
		{
			m.GET("", controller.GetMessages)
		}

		s := v1.Group("/system")
		{
			s.GET("/load", controller.GetLoadAverage)
			s.GET("/mem", controller.GetMemoryUsage)
			s.GET("/cpu", controller.GetCpuUsage)
		}
	}

}

func createJWTConfig() middleware.JWTConfig {
	d := middleware.DefaultJWTConfig

	d.SigningKey = []byte(param.WebJwt)
	d.TokenLookup = "header:token:duckduckgo "
	d.AuthScheme = "duckduckgo"

	filterUri := []string{"webapp", "js", "css"}

	d.Skipper = func(c echo.Context) bool {
		uri := c.Request().RequestURI
		if strings.Contains(uri, "/v1/login") {
			return true
		}

		for _, v := range filterUri {
			if strings.Contains(uri, v) {
				return true
			}
		}

		if strings.Contains(uri, "/v1/task/status") {
			return true
		}

		return false
	}

	return d
}

func addApp(e *echo.Echo) {
	webapp := packr.New("webapp", "../webapp")

	app, err := webapp.FindString("index.html")
	if err != nil {
		log.Fatalf("not find the index.html of app : %v", err)
	}

	static := packr.New("static", "../webapp")

	e.GET("/*", echo.WrapHandler(http.StripPrefix("/", http.FileServer(static))))
	e.GET("/app", func(c echo.Context) error {
		return c.HTML(200, app)
	})

}

func CreateEngine() (*echo.Echo, error) {
	e := echo.New()

	addApi(e)
	addApp(e)

	return e, nil
}
