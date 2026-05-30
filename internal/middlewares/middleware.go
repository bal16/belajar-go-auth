package middlewares

import (
	"auth/domain"
)

type customMiddleware struct {
	jwtSer  domain.JWTService
	rbacSer domain.RBACService
}

func New(jwtSer domain.JWTService, rbacSer domain.RBACService) *customMiddleware {
	return &customMiddleware{
		jwtSer:  jwtSer,
		rbacSer: rbacSer,
	}
}
