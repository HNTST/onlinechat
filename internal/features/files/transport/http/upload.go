package files_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

const maxAvatarSize = 5 * 1024 * 1024

var allowedMIMETypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type UploadAvatarResponse struct {
	Id int64 `json:"file_id"`
}

// UploadAvatar uploads an avatar image for a user
//
//	@Summary		Загрузка аватара
//	@Description	Загрузка файла аватара пользователя в MinIO и сохранение метаданных
//	@Tags			v2 / Файлы
//	@Accept			mpfd
//	@Produce		json
//	@Param			user_id			query		int64	true	"ID пользователя"
//	@Param			new_user_avatar	formData	file	true	"Файл изображения аватара"
//	@Success		201				{object}	UploadAvatarResponse
//	@Failure		400				{object}	object{}
//	@Router			/api/v2/user/avatar [post]
func (h *FileHTTPHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "param is incorrect")
		return
	}

	file, header, err := r.FormFile("new_user_avatar")
	if err != nil {
		log.Warn("не удалось прочитать файл аватарки", slog.String("error", err.Error()))
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "неверный файл аватарки")
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		log.Warn("размер файла превышает лимит", slog.Int64("size", header.Size), slog.Int("max", maxAvatarSize))
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "размер файла превышает 5MB")
		return
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		log.Warn("ошибка чтения первых байт файла", slog.String("error", err.Error()))
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "не удалось прочитать файл")
		return
	}
	buf = buf[:n]

	detectedMIME := http.DetectContentType(buf)
	ext, ok := allowedMIMETypes[detectedMIME]
	if !ok {
		log.Warn("запрещённый MIME-тип", slog.String("detected", detectedMIME), slog.String("client_mime", header.Header.Get("Content-Type")))
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "допустимы только изображения JPEG, PNG или WebP")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Warn("ошибка сброса указателя файла", slog.String("error", err.Error()))
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "ошибка обработки файла")
		return
	}

	fileID, err := h.fileService.UploadAvatar(ctx, int64(userID), file, header.Size, detectedMIME, ext)
	if err != nil {
		log.Warn("ошибка сохранения аватарки", slog.String("error", err.Error()))
		responseHandler.ErrorResponse(core_errors.ErrInvalidArgument, "не удалось сохранить аватарку")
		return
	}

	responseHandler.JSONResponse(UploadAvatarResponse{Id: fileID}, http.StatusCreated)
}
