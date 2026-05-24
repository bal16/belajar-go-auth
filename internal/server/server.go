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
	healthSer domain.HealthService
	validator *services.CustomValidator
}

func NewServer() *http.Server {
	cnf := config.Get()

	port, _ := strconv.Atoi(cnf.Server.PORT)

	_, sqlDB := connection.GetDatabase(cnf.Database)

	customValidator := services.NewCustomValidator(v.New())

	healthRepo := repositories.NewHealthRepository(sqlDB)
	healthSer := services.NewHealthService(healthRepo)

	NewServer := &Server{
		port:      port,
		healthSer: healthSer,
		validator: customValidator,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
