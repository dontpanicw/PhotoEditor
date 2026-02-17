package domain

import "testing"

func TestImageConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"ResizeAction", ResizeAction, "Resize"},
		{"MiniatureGenerateAction", MiniatureGenerateAction, "Miniature_generate"},
		{"WatermarkAction", WatermarkAction, "Watermark"},
		{"ImageStatusPending", ImageStatusPending, "Pending"},
		{"ImageStatusDone", ImageStatusDone, "Done"},
		{"ImageStatusFailed", ImageStatusFailed, "Failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

func TestImageStruct(t *testing.T) {
	image := Image{
		Id:                      "test-id",
		FileName:                "test.jpg",
		FileSize:                1024,
		RawImageObjectKey:       "raw/test.jpg",
		ProcessedImageObjectKey: "processed/test.jpg",
		Actions:                 []string{ResizeAction, WatermarkAction},
		Status:                  ImageStatusPending,
	}

	if image.Id != "test-id" {
		t.Errorf("Expected Id to be 'test-id', got %s", image.Id)
	}

	if len(image.Actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(image.Actions))
	}

	if image.Status != ImageStatusPending {
		t.Errorf("Expected status to be Pending, got %s", image.Status)
	}
}

func TestTaskMessage(t *testing.T) {
	task := TaskMessage{
		ImageID:   "test-id",
		Actions:   []string{ResizeAction},
		Timestamp: 1234567890,
	}

	if task.ImageID != "test-id" {
		t.Errorf("Expected ImageID to be 'test-id', got %s", task.ImageID)
	}

	if len(task.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(task.Actions))
	}

	if task.Timestamp != 1234567890 {
		t.Errorf("Expected timestamp to be 1234567890, got %d", task.Timestamp)
	}
}
