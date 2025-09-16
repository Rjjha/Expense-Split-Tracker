package utils

import (
	"regexp"
	"strings"

	"expense-split-tracker/pkg/errors"

	"github.com/shopspring/decimal"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return errors.NewRequiredFieldError("email")
	}
	if !emailRegex.MatchString(email) {
		return errors.NewInvalidValueError("email", email)
	}
	return nil
}

// ValidateName validates name field
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.NewRequiredFieldError("name")
	}
	if len(name) < 2 {
		return errors.NewValidationError("Name must be at least 2 characters long")
	}
	if len(name) > 255 {
		return errors.NewValidationError("Name must be less than 255 characters")
	}
	return nil
}

// ValidateAmount validates monetary amount
func ValidateAmount(amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.NewValidationError("Amount must be greater than zero")
	}
	if amount.GreaterThan(decimal.NewFromFloat(999999999.99)) {
		return errors.NewValidationError("Amount is too large")
	}
	return nil
}

// ValidateDescription validates description field
func ValidateDescription(description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return errors.NewRequiredFieldError("description")
	}
	if len(description) > 1000 {
		return errors.NewValidationError("Description must be less than 1000 characters")
	}
	return nil
}

// ValidatePercentage validates percentage value
func ValidatePercentage(percentage decimal.Decimal) error {
	if percentage.LessThan(decimal.Zero) {
		return errors.NewValidationError("Percentage cannot be negative")
	}
	if percentage.GreaterThan(decimal.NewFromInt(100)) {
		return errors.NewValidationError("Percentage cannot be greater than 100")
	}
	return nil
}

// ValidatePercentageSum validates that percentages sum to 100
func ValidatePercentageSum(percentages []decimal.Decimal) error {
	sum := decimal.Zero
	for _, p := range percentages {
		sum = sum.Add(p)
	}

	if !sum.Equal(decimal.NewFromInt(100)) {
		return errors.NewValidationError("Percentages must sum to 100")
	}
	return nil
}
