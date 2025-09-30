package validation

import (
	"L0/internal/transport/dto"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	UnknownError = errors.New("unknown error")
	validate     = validator.New()
)

func ValidateOrder(order *dto.Order) error {
	const op = "ValidateOrder"

	err := validate.Struct(order)
	if err == nil {
		fmt.Println(err)
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return fmt.Errorf("%s: %w", op, UnknownError)
	}

	var errorMessages []string

	for _, fieldErr := range validationErrors {
		msg := fmt.Sprintf("Field %s isn't valid, validation tag - %s.", fieldErr.Field(), fieldErr.Tag())
		errorMessages = append(errorMessages, msg)
	}

	return fmt.Errorf("Order isn't valid:\n\t-%s",
		strings.Join(errorMessages, "\n\t-"))
}
