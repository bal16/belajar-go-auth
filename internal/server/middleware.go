package server

import (
	"auth/domain"
	"auth/dto"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type customMiddleware struct {
	jwtSer domain.JWTService
}

func (m customMiddleware) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false, 
				Message: "Unauthorized", 
				Error: "Missing token",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false, 
				Message: "Unauthorized", 
				Error: "Invalid token format",
			})
		}

		tokenString := parts[1]


		userID, err := m.jwtSer.ParseToken(tokenString)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
				Success: false, 
				Message: "Unauthorized", 
				Error: "Invalid or expired token",
			})
		}
		
		c.Set("user_id", userID)

		return next(c)
	}
}

func NewCustomMiddleware(jwtSer domain.JWTService) *customMiddleware {
	return &customMiddleware{
		jwtSer: jwtSer,
	}
}