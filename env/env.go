package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func Load(target any) error {
	if err := load(target); err != nil {
		return err
	}
	return nil
}

func MustLoad(target any) {
	if err := Load(target); err != nil {
		panic(err)
	}
}

func load(target any) error {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	value := reflect.ValueOf(target)

	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() || value.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("env.Load expects a pointer to a struct")
	}

	return loadStruct(value.Elem())
}

func loadStruct(value reflect.Value) error {
	typ := value.Type()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := typ.Field(i)

		// Unexported field skip
		if !field.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("env")

		if tag == "" {
			if field.Kind() == reflect.Struct && field.Type() != reflect.TypeOf(time.Time{}) {
				if err := loadStruct(field); err != nil {
					return err
				}
			}
			continue
		}

		parts := strings.Split(tag, ",")
		key := parts[0]

		required := false

		for _, option := range parts[1:] {
			if option == "required" {
				required = true
			}
		}

		valueFromEnv, exists := os.LookupEnv(key)

		if !exists {
			defaultValue := fieldType.Tag.Get("default")

			if defaultValue != "" {
				valueFromEnv = defaultValue
			}
		}

		if (!exists || valueFromEnv == "") && required {
			return fmt.Errorf(
				"required environment variable %q is not set",
				key,
			)
		}

		if !exists && valueFromEnv == "" {
			continue
		}

		if err := setField(field, valueFromEnv); err != nil {
			return fmt.Errorf(
				"failed to parse environment variable %q: %w",
				key,
				err,
			)
		}
	}

	return nil
}

func setField(field reflect.Value, value string) error {
	if field.Type() == reflect.TypeOf(time.Duration(0)) {
		result, err := time.ParseDuration(value)
		if err != nil {
			return err
		}

		field.SetInt(int64(result))
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}

		field.SetInt(result)

	case reflect.Bool:
		result, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		field.SetBool(result)

	case reflect.Float32, reflect.Float64:
		result, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}

		field.SetFloat(result)

	default:
		return fmt.Errorf(
			"unsupported field type %s",
			field.Type(),
		)
	}

	return nil
}
