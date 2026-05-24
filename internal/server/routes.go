package server

import (
	"auth/internal/handlers"
	"auth/internal/middlewares"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	e := echo.New()
	e.Validator = s.validator

	e.Logger.SetLevel(2)

	file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	e.Logger.SetOutput(io.MultiWriter(os.Stdout, file))

	e.Logger.SetHeader("${time_rfc3339} [${level}] ${short_file}:${line} =>")

	e.Logger.Infof("log level set to: %v", e.Logger.Level())

	cm := middlewares.New(s.jwtSer)

	e.Logger.Info("Registering routes...")
	hh := handlers.NewHealthHandler(s.healthSer)
	e.Logger.Info("HealthHandler registered successfully.")
	ah := handlers.NewAuthHandler(s.authSer)
	e.Logger.Info("AuthHandler registered successfully.")

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogMethod: true,
		LogStatus: true,
		LogError:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			c.Logger().SetHeader("${time_rfc3339} [${level}] REQUEST =>")
			c.Logger().Infof("Method: %s Uri: %s Status: %d Error: %v", v.Method, v.URI, v.Status, v.Error)
			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"https://*", "http://*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	e.GET("/", hh.HelloWorldHandler)
	e.GET("/health", hh.HealthHandler)

	auth := e.Group("/auth")

	auth.POST("/register", ah.Register)
	auth.POST("/login", ah.Login)
	auth.POST("/refresh", ah.RefreshToken)
	auth.POST("/google", ah.GoogleLogin)
	auth.GET("/me", ah.GetMe, cm.AuthMiddleware)
	auth.DELETE("/logout", ah.Logout)

	return e
}
