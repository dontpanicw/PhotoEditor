package processor

import (
	"os"
	"testing"
)

func TestResizeImage(t *testing.T) {
	// Создаем тестовое изображение (простой JPEG)
	testImage := createTestJPEG(t)

	result, err := ResizeImage(testImage, 100, 100)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}

	if len(result) >= len(testImage) {
		t.Error("Expected resized image to be smaller")
	}
}

func TestGenerateSmartThumbnail(t *testing.T) {
	testImage := createTestJPEG(t)

	result, err := GenerateSmartThumbnail(testImage, 50, 50)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}
}

func TestAddTextWatermark(t *testing.T) {
	testImage := createTestJPEG(t)

	result, err := AddTextWatermark(testImage, "Test Watermark")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}
}

func TestResizeImage_InvalidData(t *testing.T) {
	invalidData := []byte("not an image")

	_, err := ResizeImage(invalidData, 100, 100)

	if err == nil {
		t.Fatal("Expected error for invalid image data, got nil")
	}
}

func TestGenerateSmartThumbnail_InvalidData(t *testing.T) {
	invalidData := []byte("not an image")

	_, err := GenerateSmartThumbnail(invalidData, 50, 50)

	if err == nil {
		t.Fatal("Expected error for invalid image data, got nil")
	}
}

func TestApplyGrayscale(t *testing.T) {
	testImage := createTestJPEG(t)

	result, err := ApplyGrayscale(testImage)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected non-empty result")
	}
}

func TestApplyGrayscale_InvalidData(t *testing.T) {
	invalidData := []byte("not an image")

	_, err := ApplyGrayscale(invalidData)

	if err == nil {
		t.Fatal("Expected error for invalid image data, got nil")
	}
}

// createTestJPEG создает минимальное валидное JPEG изображение для тестов
func createTestJPEG(t *testing.T) []byte {
	// Минимальный валидный JPEG (1x1 пиксель, черный)
	jpeg := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46,
		0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48,
		0x00, 0x48, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
		0x00, 0x03, 0x02, 0x02, 0x02, 0x02, 0x02, 0x03,
		0x02, 0x02, 0x02, 0x03, 0x03, 0x03, 0x03, 0x04,
		0x06, 0x04, 0x04, 0x04, 0x04, 0x04, 0x08, 0x06,
		0x06, 0x05, 0x06, 0x09, 0x08, 0x0A, 0x0A, 0x09,
		0x08, 0x09, 0x09, 0x0A, 0x0C, 0x0F, 0x0C, 0x0A,
		0x0B, 0x0E, 0x0B, 0x09, 0x09, 0x0D, 0x11, 0x0D,
		0x0E, 0x0F, 0x10, 0x10, 0x11, 0x10, 0x0A, 0x0C,
		0x12, 0x13, 0x12, 0x10, 0x13, 0x0F, 0x10, 0x10,
		0x10, 0xFF, 0xC9, 0x00, 0x0B, 0x08, 0x00, 0x01,
		0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xCC,
		0x00, 0x06, 0x00, 0x10, 0x10, 0x05, 0xFF, 0xDA,
		0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00,
		0xD2, 0xCF, 0x20, 0xFF, 0xD9,
	}

	// Проверяем что libvips доступен
	if _, err := os.Stat("/usr/lib/libvips.so"); os.IsNotExist(err) {
		t.Skip("libvips not available, skipping image processing tests")
	}

	return jpeg
}
