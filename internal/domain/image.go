package domain

const (
	ResizeAction            = "Resize"
	MiniatureGenerateAction = "Miniature_generate"
	WatermarkAction         = "Watermark"

	ImageStatusPending = "Pending"
	ImageStatusDone    = "Done"
	ImageStatusFailed  = "Failed"
)

type Image struct {
	Id                      string   `json:"id"`
	FileName                string   `json:"filename"`
	FileSize                int64    `json:"file_size"`
	RawImageObjectKey       string   `json:"raw_image_id"`
	ProcessedImageObjectKey string   `json:"processed_image_id,omitempty"`
	Actions                 []string `json:"action"`
	Status                  string   `json:"status,omitempty"`
}

// TaskMessage - структура сообщения для Kafka
type TaskMessage struct {
	ImageID   string   `json:"image_id"`
	Actions   []string `json:"actions"`
	Timestamp int64    `json:"timestamp"`
}
