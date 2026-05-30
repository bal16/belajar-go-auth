package middlewares

import (
	"auth/dto"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (m customMiddleware) RBACMiddleware(next echo.HandlerFunc, permission string) echo.HandlerFunc {
	return func(c echo.Context) error {
		userRoles, ok := c.Get("user_roles").([]string)
		if !ok {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Success: false,
				Message: "Forbidden",
				Error:   "User roles not found in context",
			})
		}

		hasPermission, err := m.rbacSer.CheckPermission(c.Request().Context(), userRoles, permission)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
				Success: false,
				Message: "Internal Server Error",
				Error:   "Failed to check permissions",
			})
		}

		if !hasPermission {
			return c.JSON(http.StatusForbidden, dto.ErrorResponse{
				Success: false,
				Message: "Forbidden",
				Error:   "You don't have permission to access this resource",
			})
		}

		return next(c)
	}
}
