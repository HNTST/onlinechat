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
	h.errorResponse(statusCode, core_errors.ErrInternal.Error(), msg)
}

func (h *HTTPResponseHandler) ErrorResponse(err error, msg string) {
	var (
		statusCode int
		errorMsg   string
		logFunc    func(msg string, args ...any)
	)
	switch {
	case errors.Is(err, core_errors.ErrInvalidArgument):
		statusCode = http.StatusBadRequest
		errorMsg = core_errors.ErrInvalidArgument.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrParamMissing):
		statusCode = http.StatusBadRequest
		errorMsg = core_errors.ErrParamMissing.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrInvalidJSON):
		statusCode = http.StatusBadRequest
		errorMsg = core_errors.ErrInvalidJSON.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrUserNotFound):
		statusCode = http.StatusNotFound
		errorMsg = core_errors.ErrUserNotFound.Error()
		logFunc = h.log.Debug
	case errors.Is(err, core_errors.ErrChatNotExists):
		statusCode = http.StatusNotFound
		errorMsg = core_errors.ErrChatNotExists.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrMessageNotFound):
		statusCode = http.StatusNotFound
		errorMsg = core_errors.ErrMessageNotFound.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrConfict):
		statusCode = http.StatusConflict
		errorMsg = core_errors.ErrConfict.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrUserAlreadyExists):
		statusCode = http.StatusConflict
		errorMsg = core_errors.ErrUserAlreadyExists.Error()
		logFunc = h.log.Warn
	case errors.Is(err, core_errors.ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		errorMsg = core_errors.ErrUnauthorized.Error()
		logFunc = h.log.Warn
	default:
		statusCode = http.StatusInternalServerError
		errorMsg = core_errors.ErrInternal.Error()
		logFunc = h.log.Error
	}
	logFunc(msg, "error", err)
	h.errorResponse(statusCode, errorMsg, msg)
}

func (h *HTTPResponseHandler) errorResponse(statusCode int, errorMsg string, msg string) {
	response := map[string]string{
		"message": msg,
		"error":   errorMsg,
	}
	h.JSONResponse(
		response,
		statusCode,
	)
}
