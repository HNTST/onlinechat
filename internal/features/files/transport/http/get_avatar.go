package files_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	core_http_utils "discord/internal/core/transport/utils"
	"fmt"
	"io"
	"net/http"
)

// GetAvatarByID отдает файл аватарки по его ID из базы данных
//
//	@Summary		Получить изображение аватарки по ID
//	@Description	Находит метаданные файла в таблице files по ID, запрашивает объект из MinIO и стримит бинарный поток изображения для тега <img />
//	@Tags			v2 / Файлы
//	@Produce		image/jpeg
//	@Produce		image/png
//	@Param			id	query		int64	true	"Идентификатор аватарки (avatar_id)"
//	@Success		200	{file}		binary "Бинарный поток изображения"
//	@Failure		404	{object}	object "Файл или аватарка не найдены"
//	@Router			/api/v2/user/avatar [get]
func (h *FileHTTPHandler) GetAvatarByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

	id, err := getAvatarIDQueryParam(r)
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to get 'id' query param")
		return
	}

	stream, mimeType, err := h.fileService.GetAvatarByID(ctx, int64(*id))
	if err != nil {
		responseHandler.ErrorResponse(core_errors.ErrUserNotFound, fmt.Sprintf("файл или аватарка с id=%d не найдены", *id))
		return
	}
	defer stream.Close()

	w.Header().Set("Content-Type", mimeType)
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, stream); err != nil {
		log.Debug("аватарка отправлена не полностью — клиент оборвал соединение", "error", err)
	}
}

func getAvatarIDQueryParam(r *http.Request) (*int, error) {
	id, err := core_http_utils.GetIntQueryParam(r, "id")
	if err != nil {
		return nil, fmt.Errorf("get 'id' query param: %w", err)
	}
	if id == nil {
		return nil, fmt.Errorf("missing required 'id' query param: %w", core_errors.ErrInvalidArgument)
	}
	return id, nil
}
