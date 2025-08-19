package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

func (app *Application) WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	js, err := json.Marshal(data)
	if err != nil {
		app.WriteError(w, r, app.Err.ServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}

func (app *Application) WriteError(w http.ResponseWriter, r *http.Request, errorType ErrorType, err error) {
	if err != nil {
		app.Log.Error(err.Error(), "method", r.Method, "uri", r.URL.RequestURI())
	}

	errorConfig := ErrorMap[errorType]

	resBody := map[string]any{
		"error-type": errorType,
		"message":    errorConfig.Message,
	}

	if err != nil {
		resBody["error"] = err.Error()
	}

	app.WriteJSON(w, r, errorConfig.Status, resBody)
}

func (app *Application) ReadJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	err = app.Validate.Struct(dst)
	if err != nil {
		var messages []string

		for _, err := range err.(validator.ValidationErrors) {
			var message string
			fieldName := strings.ToLower(err.Field())

			switch err.Tag() {
			case "required":
				message = fmt.Sprintf(`"%s" is required`, fieldName)
			case "email":
				message = fmt.Sprintf(`"%s" must be a valid email address`, fieldName)
			case "min":
				if err.Kind().String() == "string" {
					message = fmt.Sprintf(`"%s" must be at least %s characters long`, fieldName, err.Param())
				} else {
					message = fmt.Sprintf(`"%s" must be at least %s`, fieldName, err.Param())
				}
			case "max":
				if err.Kind().String() == "string" {
					message = fmt.Sprintf(`"%s" must be at most %s characters long`, fieldName, err.Param())
				} else {
					message = fmt.Sprintf(`"%s" must be at most %s`, fieldName, err.Param())
				}
			default:
				if err.Param() != "" {
					message = fmt.Sprintf(`"%s" failed validation '%s' with parameter '%s' (current value: '%v')`,
						fieldName, err.Tag(), err.Param(), err.Value())
				} else {
					message = fmt.Sprintf(`"%s" failed validation '%s' (current value: '%v')`,
						fieldName, err.Tag(), err.Value())
				}
			}

			messages = append(messages, message)
		}

		return errors.New("There were issues with your body: " + strings.Join(messages, ", "))
	}

	return nil
}
