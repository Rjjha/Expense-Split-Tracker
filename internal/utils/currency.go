package utils

import (
	"strings"

	"expense-split-tracker/pkg/errors"
)

// SupportedCurrencies defines the list of supported currencies
var SupportedCurrencies = map[string]bool{
	"USD": true,
	"EUR": true,
	"GBP": true,
	"JPY": true,
	"CAD": true,
	"AUD": true,
	"CHF": true,
	"CNY": true,
	"INR": true,
}

// ValidateCurrency checks if the currency is supported
func ValidateCurrency(currency string) error {
	currency = strings.ToUpper(currency)
	if !SupportedCurrencies[currency] {
		return errors.NewInvalidValueError("currency", currency)
	}
	return nil
}

// NormalizeCurrency converts currency to uppercase
func NormalizeCurrency(currency string) string {
	return strings.ToUpper(currency)
}

// AreCurrenciesCompatible checks if two currencies are compatible
func AreCurrenciesCompatible(currency1, currency2 string) bool {
	return NormalizeCurrency(currency1) == NormalizeCurrency(currency2)
}
