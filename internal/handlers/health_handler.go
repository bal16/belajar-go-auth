package handlers

import (
	"auth/domain"
	"auth/dto"
	"net/http"

	"github.com/labstack/echo/v4"
)

type healthHandler struct {
	service domain.HealthService
}

func NewHealthHandler(service domain.HealthService) *healthHandler {
	return &healthHandler{
		service: service,
	}
}

func (h *healthHandler) HelloWorldHandler(c echo.Context) error {
	resp := dto.BaseResponse{
		Success: true,
		Message: "Hello World",
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *healthHandler) HealthHandler(c echo.Context) error {
	// c.Logger().Info("Health check endpoint hit")
	resp := dto.BaseResponse{
		Success: true,
		Message: "Health Check OK",
		Data:    h.service.Health(c.Request().Context()),
	}
	return c.JSON(http.StatusOK, resp)
}
