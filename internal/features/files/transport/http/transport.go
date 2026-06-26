package files_transport_http

import (
	"context"
	core_http_server "discord/internal/core/transport/server"
	"io"
	"net/http"
)

type FileHTTPHandler struct {
	fileService FileService
}

type FileService interface {
	UploadAvatar(ctx context.Context, userID int64, file io.Reader, origFileName string, size int64, contentType string) (int64, error)
}

func NewFileHTTPHandler(fileService FileService) *FileHTTPHandler {
	return &FileHTTPHandler{
		fileService: fileService,
	}
}

func (h *FileHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/user/avatar",
			Handler: h.UploadAvatar,
		},
	}
}
