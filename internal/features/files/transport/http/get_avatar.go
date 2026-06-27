package files_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"io"
	"net/http"
	"strings"
)

var mimeTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

func contentTypeByExt(filename string) string {
	idx := strings.LastIndex(filename, ".")
	if idx == -1 {
		return "image/jpeg"
	}
	ext := strings.ToLower(filename[idx:])
	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "image/jpeg"
}

// GetAvatar отдает файл аватарки для отображения на UI
//
//	@Summary		Получить изображение аватарки
//	@Description	Возвращает бинарный поток файла изображения из хранилища MinIO для прямой вставки в HTML-тег <img />
//	@Tags			v2 / Файлы
//	@Produce		image/jpeg
//	@Produce		image/png
//	@Param			filename	query	string	true	"Имя файла аватарки в хранилище"
//	@Success		200			{file}	binary	"Бинарный поток изображения"
//	@Failure		404			{object}	object{}	"Файл не найден"
//	@Router			/api/v2/user/avatar [get]
func (h *FileHTTPHandler) GetAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "filename is required")
		return
	}

	stream, err := h.fileService.GetAvatarStream(ctx, filename)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrUserNotFound, "file not found")
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", contentTypeByExt(filename))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, stream)
}
