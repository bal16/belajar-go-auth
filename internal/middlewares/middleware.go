package middlewares

import (
	"auth/domain"
)

type customMiddleware struct {
	jwtSer domain.JWTService
}

func New(jwtSer domain.JWTService) *customMiddleware {
	return &customMiddleware{
		jwtSer: jwtSer,
	}
}
