package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	v "github.com/go-playground/validator/v10"

	"auth/domain"
	"auth/internal/config"
	"auth/internal/connection"
	"auth/internal/repositories"
	"auth/internal/services"
)

type Server struct {
	port      int
	authSer   domain.AuthService
	jwtSer    domain.JWTService
	rbacSer   domain.RBACService
	healthSer domain.HealthService
	validator *services.CustomValidator
}

func NewServer() *http.Server {
	cnf := config.Get()

	port, _ := strconv.Atoi(cnf.Server.PORT)

	dbConnection, sqlDB := connection.GetDatabase(cnf.Database)

	customValidator := services.NewCustomValidator(v.New())

	jwtSer := services.NewJWTService(cnf)
	userRepo := repositories.NewUser(dbConnection)
	authSer := services.NewAuthService(userRepo, jwtSer)
	healthRepo := repositories.NewHealthRepository(sqlDB)
	healthSer := services.NewHealthService(healthRepo)
	rbacCacheRepo := repositories.NewInMemoryRBACRepository()
	rbacRepo := repositories.NewRBACRepository(dbConnection)
	rbacSer := services.NewRBACService(rbacCacheRepo, rbacRepo)

	srv := &Server{
		port:      port,
		authSer:   authSer,
		jwtSer:    jwtSer,
		rbacSer:   rbacSer,
		healthSer: healthSer,
		validator: customValidator,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", srv.port),
		Handler:      srv.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
