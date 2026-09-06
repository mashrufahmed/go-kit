package validate

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

func Struct(value any) error {
	return v.Struct(value)
}

func IsValidationError(err error) bool {
	var validationErr validator.ValidationErrors

	return errors.As(
		err,
		&validationErr,
	)
}

func Errors(err error) validator.ValidationErrors {
	var validationErr validator.ValidationErrors

	if errors.As(err, &validationErr) {
		return validationErr
	}

	return nil
}

func Details(value any, errs ...error) map[string]string {
	var err error
	if len(errs) > 0 {
		err = errs[0]
	} else {
		err, value = value.(error), nil
	}
	validationErrs := Errors(err)

	if validationErrs == nil {
		return nil
	}

	result := make(map[string]string)

	typ := reflect.TypeOf(value)
	if typ == nil {
		typ = reflect.TypeOf(struct{}{})
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for _, fieldErr := range validationErrs {
		field := jsonPath(typ, fieldErr.StructNamespace())

		switch fieldErr.Tag() {
		case "required":
			result[field] = "is required"

		case "email":
			result[field] = "must be a valid email"

		case "min":
			result[field] = "is too short"

		case "max":
			result[field] = "is too long"

		default:
			result[field] = "is invalid"
		}
	}

	return result
}

func jsonPath(root reflect.Type, namespace string) string {
	parts := strings.Split(namespace, ".")
	if len(parts) == 0 {
		return namespace
	}
	if root.Kind() == reflect.Ptr {
		root = root.Elem()
	}
	if root.Name() == parts[0] {
		parts = parts[1:]
	}
	result := make([]string, 0, len(parts))
	current := root
	for _, part := range parts {
		if current.Kind() == reflect.Ptr {
			current = current.Elem()
		}
		field, ok := current.FieldByName(part)
		if !ok {
			result = append(result, part)
			continue
		}
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" {
			name = part
		}
		if name != "-" {
			result = append(result, name)
		}
		current = field.Type
	}
	return strings.Join(result, ".")
}
