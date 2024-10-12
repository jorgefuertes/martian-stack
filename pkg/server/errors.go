package server

import "net/http"

var (
	ErrNotFound          = HttpError{Code: http.StatusNotFound, Msg: "Resource not found"}
	ErrSessionNotStarted = HttpError{Code: http.StatusInternalServerError, Msg: "Session not started"}
)
