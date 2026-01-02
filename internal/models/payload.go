package models

type SecretPayload struct {
	Text     string `json:"text,omitempty"`
	FileData []byte `json:"file_data,omitempty"`
	Filename string `json:"filename,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	IsFile   bool   `json:"is_file"`
}
