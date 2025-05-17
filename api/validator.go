package api

import (
	"BankAppGo/db/util"

	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		// checks currency
		return util.IsSupportedCurrency(currency)

	}
	return false
}
