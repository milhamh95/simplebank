package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/milhamh95/simplebank/pkg/currency"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	cur, ok := fieldLevel.Field().Interface().(string)
	if !ok {
		return false
	}

	return currency.IsSupportedCurrency(cur)
}
