package services

import (
	v "github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *v.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func NewCustomValidator(validator *v.Validate) *CustomValidator {
	return &CustomValidator{validator: validator}
}

