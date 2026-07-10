package core_http_response

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type HTTPResponseHandler struct {
	log *core_logger.Logger
	rw  http.ResponseWriter
}

func NewHTTPResponseHandler(log *core_logger.Logger, rw http.ResponseWriter) *HTTPResponseHandler {
	return &HTTPResponseHandler{
		log: log,
		rw:  rw,
	}
}

func (h *HTTPResponseHandler) JSONResponse(responseBody any, statusCode int) {
	h.rw.WriteHeader(statusCode)
	if err := json.NewEncoder(h.rw).Encode(responseBody); err != nil {
		h.log.Error("запись HTTP ответа", "error", err)
	}
}

func (h *HTTPResponseHandler) PanicResponse(p any, msg string) {
	statusCode := http.StatusInternalServerError
	err := fmt.Errorf("unexpected panic: %v", p)
	h.log.Error(msg, "error", err)
	h.errorResponse(statusCode, err, msg)
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		logFunc    func(msg string, args ...any)
	)
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrUserNotFound):
		statusCode = http.StatusNotFound
		logFunc = h.log.Debug
	case errors.Is(err, core_errors.ErrConfict):
		statusCode = http.StatusConflict
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		logFunc = h.log.Warn
	default:
		statusCode = http.StatusInternalServerError
		logFunc = h.log.Error
	}
	logFunc(msg, "error", err)
	h.errorResponse(statusCode, err, msg)
}

func (h *HTTPResponseHandler) errorResponse(statusCode int, err error, msg string) {
	errorMsg := err.Error()
	if statusCode == http.StatusInternalServerError {
		errorMsg = "Внутренняя ошибка сервера"
	}
	response := map[string]string{
		"message": msg,
		"error":   errorMsg,
	}
	h.JSONResponse(
		response,
		statusCode,
	)
}
