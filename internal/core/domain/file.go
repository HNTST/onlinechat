package domain

type Files struct {
	Id               int64  `json:"id"`
	FileKey          string `json:"file_key"`
	OriginalFileName string `json:"original_file_name"`
	TypeFile         string `json:"type_file"`
	SizeFile         int64  `json:"size_file"`
}
