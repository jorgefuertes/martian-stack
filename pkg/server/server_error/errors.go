package server_error

import "net/http"

var (
	ErrNotFound          = Error{Code: http.StatusNotFound, Msg: "Resource not found"}
	ErrSessionNotStarted = Error{Code: http.StatusInternalServerError, Msg: "Session not started"}
)
