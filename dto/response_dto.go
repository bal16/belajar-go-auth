package dto

type BaseResponse struct {
	Success bool   `json:"success" default:"true"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success" default:"false"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}