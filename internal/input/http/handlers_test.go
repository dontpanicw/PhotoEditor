package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"github.com/gorilla/mux"
)

type mockUsecases struct {
	createObjectFunc    func(ctx context.Context, image domain.Image, r io.Reader, size int64, contentType string) (string, error)
	getObjectByIDFunc   func(ctx context.Context, id string) (io.ReadCloser, error)
	getImageStatusFunc  func(ctx context.Context, id string) (*domain.Image, error)
	removeObjectFunc    func(ctx context.Context, id string) error
}

func (m *mockUsecases) InitMinio() error {
	return nil
}

func (m *mockUsecases) CreateObject(ctx context.Context, image domain.Image, r io.Reader, size int64, contentType string) (string, error) {
	if m.createObjectFunc != nil {
		return m.createObjectFunc(ctx, image, r, size, contentType)
	}
	return "test-id", nil
}

func (m *mockUsecases) GetObjectByID(ctx context.Context, id string) (io.ReadCloser, error) {
	if m.getObjectByIDFunc != nil {
		return m.getObjectByIDFunc(ctx, id)
	}
	return io.NopCloser(strings.NewReader("test image")), nil
}

func (m *mockUsecases) GetImageStatus(ctx context.Context, id string) (*domain.Image, error) {
	if m.getImageStatusFunc != nil {
		return m.getImageStatusFunc(ctx, id)
	}
	return &domain.Image{Id: id, Status: domain.ImageStatusDone}, nil
}

func (m *mockUsecases) RemoveObject(ctx context.Context, id string) error {
	if m.removeObjectFunc != nil {
		return m.removeObjectFunc(ctx, id)
	}
	return nil
}

func TestUploadImage_Success(t *testing.T) {
	usecases := &mockUsecases{}
	handler := NewHandler(usecases)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, _ := writer.CreateFormFile("image", "test.jpg")
	part.Write([]byte("fake image data"))
	writer.WriteField("actions", "Resize")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.UploadImage(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response map[string]string
	json.NewDecoder(w.Body).Decode(&response)

	if response["id"] == "" {
		t.Error("Expected non-empty id in response")
	}

	if response["status"] != domain.ImageStatusPending {
		t.Errorf("Expected status Pending, got %s", response["status"])
	}
}

func TestUploadImage_NoFile(t *testing.T) {
	usecases := &mockUsecases{}
	handler := NewHandler(usecases)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.UploadImage(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetImage_Success(t *testing.T) {
	usecases := &mockUsecases{}
	handler := NewHandler(usecases)

	req := httptest.NewRequest("GET", "/image/test-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	w := httptest.NewRecorder()

	handler.GetImage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("Expected Content-Type image/jpeg, got %s", w.Header().Get("Content-Type"))
	}
}

func TestGetImage_NotFound(t *testing.T) {
	usecases := &mockUsecases{
		getObjectByIDFunc: func(ctx context.Context, id string) (io.ReadCloser, error) {
			return nil, domain.ErrImageNotFound
		},
	}
	handler := NewHandler(usecases)

	req := httptest.NewRequest("GET", "/image/nonexistent", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "nonexistent"})
	w := httptest.NewRecorder()

	handler.GetImage(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetImageStatus_Success(t *testing.T) {
	usecases := &mockUsecases{}
	handler := NewHandler(usecases)

	req := httptest.NewRequest("GET", "/image/test-id/status", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	w := httptest.NewRecorder()

	handler.GetImageStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response domain.Image
	json.NewDecoder(w.Body).Decode(&response)

	if response.Id != "test-id" {
		t.Errorf("Expected id test-id, got %s", response.Id)
	}
}

func TestDeleteImage_Success(t *testing.T) {
	usecases := &mockUsecases{}
	handler := NewHandler(usecases)

	req := httptest.NewRequest("DELETE", "/image/test-id", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "test-id"})
	w := httptest.NewRecorder()

	handler.DeleteImage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestIsValidAction(t *testing.T) {
	tests := []struct {
		action string
		valid  bool
	}{
		{domain.ResizeAction, true},
		{domain.WatermarkAction, true},
		{domain.MiniatureGenerateAction, true},
		{"InvalidAction", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := isValidAction(tt.action)
			if result != tt.valid {
				t.Errorf("isValidAction(%s) = %v, want %v", tt.action, result, tt.valid)
			}
		})
	}
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		sep      string
		expected []string
	}{
		{"a,b,c", ",", []string{"a", "b", "c"}},
		{" a , b , c ", ",", []string{"a", "b", "c"}},
		{"", ",", nil},
		{"single", ",", []string{"single"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitAndTrim(tt.input, tt.sep)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d", len(tt.expected), len(result))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}
