package handlers

import (
	"auth/domain"
	"auth/dto"
	"net/http"

	"github.com/labstack/echo/v4"
)

type authHandler struct {
	authService domain.AuthService
}

func (h *authHandler) Login(ctx echo.Context) error {
	req := new(dto.LoginRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Invalid request", Error: "Invalid request"})
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Validation failed", Error: err.Error()})
	}

	stdCtx := ctx.Request().Context()
	token, refreshToken, err := h.authService.Login(stdCtx, *req)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
	}

	return ctx.JSON(
		http.StatusOK,
		dto.BaseResponse{
			Success: true,
			Message: "Login successful",
			Data: map[string]string{
				"access_token":  token,
				"refresh_token": refreshToken,
			},
		})
}

func (h *authHandler) Register(ctx echo.Context) error {
	req := new(dto.RegisterRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Invalid request", Error: "Invalid request"})
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Validation failed", Error: err.Error()})
	}

	stdCtx := ctx.Request().Context()
	err := h.authService.Register(stdCtx, *req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{Success: false, Message: "Internal Server Error", Error: err.Error()})
	}

	return ctx.JSON(http.StatusCreated, dto.BaseResponse{Success: true, Message: "Registrasi Berhasil"})
}

func (h *authHandler) GetMe(ctx echo.Context) error {
	userId := ctx.Get("user_id")

	if userId == nil || userId == 0 {
		return ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Success: false, Message: "Unauthorized"})
	}

	stdCtx := ctx.Request().Context()
	user, err := h.authService.GetMe(stdCtx, userId.(int))

	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
	}

	return ctx.JSON(http.StatusOK, dto.BaseResponse{Success: true, Message: "User retrieved successfully", Data: user})
}

func (h *authHandler) GoogleLogin(ctx echo.Context) error {
	req := new(dto.GoogleLoginRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Invalid request", Error: "Invalid request"})
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Validation failed", Error: err.Error()})
	}

	stdCtx := ctx.Request().Context()
	token, refreshToken, err := h.authService.GoogleLogin(stdCtx, req.Token)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
	}

	return ctx.JSON(
		http.StatusOK,
		dto.BaseResponse{
			Success: true,
			Message: "Login successful",
			Data: map[string]string{
				"access_token":  token,
				"refresh_token": refreshToken,
			},
		})
}

func (h *authHandler) RefreshToken(ctx echo.Context) error {
	req := new(dto.RefreshTokenRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Invalid request", Error: err.Error()})
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Validation failed", Error: err.Error()})
	}

	stdCtx := ctx.Request().Context()
	newAccessToken, err := h.authService.RefreshToken(stdCtx, req.Token)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
	}

	return ctx.JSON(
		http.StatusOK,
		dto.BaseResponse{
			Success: true,
			Message: "Token refreshed successfully",
			Data: map[string]string{
				"access_token":  newAccessToken,
				"refresh_token": req.Token,
			},
		})
}

func (h *authHandler) Logout(ctx echo.Context) error {
	req := new(dto.LogoutRequest)
	if err := ctx.Bind(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Invalid request", Error: "Invalid request"})
	}

	if err := ctx.Validate(req); err != nil {
		return ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{Success: false, Message: "Validation failed", Error: err.Error()})
	}

	stdCtx := ctx.Request().Context()
	err := h.authService.Logout(stdCtx, req.Token)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{Success: false, Message: "Unauthorized", Error: err.Error()})
	}

	return ctx.JSON(http.StatusOK, dto.BaseResponse{Success: true, Message: "Logged out successfully"})
}

func NewAuthHandler(authService domain.AuthService) *authHandler {
	return &authHandler{
		authService,
	}
}
