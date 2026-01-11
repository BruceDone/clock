package router

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"

	"clock/internal/config"
	"clock/internal/handler"
	"clock/internal/logger"
	"clock/internal/middleware"
)

// Handlers 所有处理器
type Handlers struct {
	Task      *handler.TaskHandler
	Container *handler.ContainerHandler
	Relation  *handler.RelationHandler
	Log       *handler.LogHandler
	Auth      *handler.AuthHandler
	System    *handler.SystemHandler
	Message   *handler.MessageHandler
}

// Router 路由器
type Router struct {
	engine   *echo.Echo
	cfg      *config.Config
	handlers *Handlers
	webapp   embed.FS
}

// NewRouter 创建路由器
func NewRouter(cfg *config.Config, handlers *Handlers, webapp embed.FS) *Router {
	return &Router{
		engine:   echo.New(),
		cfg:      cfg,
		handlers: handlers,
		webapp:   webapp,
	}
}

// Setup 设置路由
func (r *Router) Setup() *echo.Echo {
	// 全局中间件
	r.engine.Use(echoMiddleware.CORS())
	r.engine.Use(middleware.Logger())
	r.engine.Use(echoMiddleware.Recover())

	// 注册API路由
	r.registerAPIRoutes()

	// 注册静态资源路由
	r.registerStaticRoutes()

	return r.engine
}

// registerAPIRoutes 注册API路由
func (r *Router) registerAPIRoutes() {
	v1 := r.engine.Group("/v1")

	// JWT中间件
	v1.Use(echoMiddleware.JWTWithConfig(middleware.NewJWTConfig(&r.cfg.Auth)))

	// 任务路由
	task := v1.Group("/task")
	{
		task.GET("", r.handlers.Task.GetTasks)
		task.GET("/:tid", r.handlers.Task.GetTask)
		task.PUT("", r.handlers.Task.PutTask)
		task.GET("/run", r.handlers.Task.RunTask)
		task.DELETE("/:tid", r.handlers.Task.DeleteTask)
		task.GET("/status", r.handlers.Message.GetTaskStatus)
	}

	// 容器路由
	container := v1.Group("/container")
	{
		container.GET("", r.handlers.Container.GetContainers)
		container.GET("/:cid", r.handlers.Container.GetContainer)
		container.PUT("", r.handlers.Container.PutContainer)
		container.GET("/run", r.handlers.Container.RunContainer)
		container.DELETE("/:cid", r.handlers.Container.DeleteContainer)
	}

	// 日志路由
	logGroup := v1.Group("/log")
	{
		logGroup.GET("", r.handlers.Log.GetLogs)
		logGroup.DELETE("", r.handlers.Log.DeleteLogs)
	}

	// 登录路由
	login := v1.Group("/login")
	{
		login.POST("", r.handlers.Auth.Login)
	}

	// 关系路由
	relation := v1.Group("/relation")
	{
		relation.GET("", r.handlers.Relation.GetRelations)
		relation.POST("", r.handlers.Relation.AddRelation)
		relation.DELETE("/:rid", r.handlers.Relation.DeleteRelation)
	}

	// 节点路由
	node := v1.Group("/node")
	{
		node.PUT("", r.handlers.Task.PutNodes)
	}

	// 消息路由
	message := v1.Group("/message")
	{
		message.GET("", r.handlers.Message.GetMessages)
	}

	// 系统监控路由
	system := v1.Group("/system")
	{
		system.GET("/load", r.handlers.System.GetLoadAverage)
		system.GET("/mem", r.handlers.System.GetMemoryUsage)
		system.GET("/cpu", r.handlers.System.GetCPUUsage)
	}
}

// registerStaticRoutes 注册静态资源路由
func (r *Router) registerStaticRoutes() {
	content, err := fs.Sub(r.webapp, "web/dist")
	if err != nil {
		logger.Fatalf("failed to load webapp: %v", err)
	}

	appBytes, err := fs.ReadFile(content, "index.html")
	if err != nil {
		logger.Fatalf("failed to load index.html: %v", err)
	}
	app := string(appBytes)

	fileServer := http.FileServer(http.FS(content))

	r.engine.GET("/*", func(c echo.Context) error {
		path := c.Request().URL.Path

		// API请求不处理
		if strings.HasPrefix(path, "/v1") {
			return echo.NewHTTPError(http.StatusNotFound)
		}

		// 尝试查找静态文件
		filePath := strings.TrimPrefix(path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		// 检查文件是否存在
		if _, err := fs.Stat(content, filePath); err == nil {
			http.StripPrefix("/", fileServer).ServeHTTP(c.Response(), c.Request())
			return nil
		}

		// SPA fallback
		return c.HTML(http.StatusOK, app)
	})
}

// Start 启动服务器
func (r *Router) Start(address string) error {
	return r.engine.Start(address)
}
