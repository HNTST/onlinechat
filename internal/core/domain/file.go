package domain

type Files struct {
	Id               int64  `json:"id"`
	FileKey          string `json:"file_key"`
	OriginalFileName string `json:"original_file_name"`
	TypeFile         string `json:"type_file"`
	SizeFile         int64  `json:"size_file"`
}

// // Для постгреса
// type FileRepository interface {
// 	UpdateUserAvatarPath(user_id int64, file Files) (int64, error)
// }
// type FileService interface {
// 	UploadUserAvatar(ctx context.Context, user_id int64, file io.Reader, orig_file_name string, size int64, content_type string) (int64, error)
// }

// // Для minio
// type FileStorage interface {
// 	UploadFile(ctx context.Context, file_name, content_type string, reader io.Reader, size int64) (string, error)
// }
