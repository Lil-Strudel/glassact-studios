package app

import (
	"net/http"
)

type ErrorType string

type appError struct {
	RouteNotFound       ErrorType
	RecordNotFound      ErrorType
	ServerError         ErrorType
	AuthenticationError ErrorType
	BadRequest          ErrorType
	MissingRefreshToken ErrorType
	AccountNotFound     ErrorType
}

var AppError = appError{
	RouteNotFound:       ErrorType("route-not-found"),
	RecordNotFound:      ErrorType("record-not-found"),
	ServerError:         ErrorType("server-error"),
	AuthenticationError: ErrorType("authentication-error"),
	BadRequest:          ErrorType("bad-request"),
	MissingRefreshToken: ErrorType("missing-refresh-token"),
	AccountNotFound:     ErrorType("account-not-found"),
}

type ErrorConfig struct {
	Status   int
	Message  string
	Expected bool
}

var ErrorMap = map[ErrorType]ErrorConfig{
	AppError.RouteNotFound: {
		Status:   http.StatusNotFound,
		Message:  "The route you have requested was not found or does not exist.",
		Expected: true,
	},
	AppError.RecordNotFound: {
		Status:   http.StatusNotFound,
		Message:  "The record you have requested was not found or does not exist.",
		Expected: true,
	},
	AppError.ServerError: {
		Status:   http.StatusInternalServerError,
		Message:  "The server encountered an unknown issue while processing your request.",
		Expected: false,
	},
	AppError.AuthenticationError: {
		Status:   http.StatusUnauthorized,
		Message:  `Invalid authentication credentials. Please provide an "Authorization" header in "Bearer $access_token" format.`,
		Expected: true,
	},
	AppError.BadRequest: {
		Status:   http.StatusBadRequest,
		Message:  `The body you have provided is not in the expected format.`,
		Expected: true,
	},
	AppError.MissingRefreshToken: {
		Status:   http.StatusUnauthorized,
		Message:  `No refresh token cookie was found.`,
		Expected: true,
	},
	AppError.AccountNotFound: {
		Status:   http.StatusUnauthorized,
		Message:  `An account was not found for you.`,
		Expected: true,
	},
}
