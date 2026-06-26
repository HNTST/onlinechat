package files_transport_http

import (
	core_errors "discord/internal/core/errors"
	core_logger "discord/internal/core/logger"
	core_http_response "discord/internal/core/transport/response"
	"net/http"
	"strconv"
)

type UploadAvatarResponse struct {
	Id int64 `json:"file_id"`
}

func (h *FileHTTPHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := core_logger.FromContext(ctx)
	responseHandler := core_http_response.NewHTTPResponseHandler(log, w)

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
		responseHandler.ErrorResponse(err, "failed to get avatar")
		return
	}
	defer file.Close()

	fileID, err := h.fileService.UploadAvatar(ctx, int64(userID), file, header.Filename, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		responseHandler.ErrorResponse(err, "failed to save user avatar")
		return
	}

	responseHandler.JSONResponse(UploadAvatarResponse{Id: fileID}, http.StatusCreated)
}
