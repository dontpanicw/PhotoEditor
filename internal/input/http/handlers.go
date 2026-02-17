package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"github.com/dontpanicw/ImageProcessor/internal/port"
	"github.com/gorilla/mux"
)

type Handler struct {
	usecases port.ImageUsecases
}

func NewHandler(usecases port.ImageUsecases) *Handler {
	return &Handler{
		usecases: usecases,
	}
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to get image file", http.StatusBadRequest)
		return
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Printf("Failed to close Kafka connection: %v", err)
		}
	}()

	// Получаем действия из формы (например: "resize,watermark")
	actionsStr := r.FormValue("actions")
	var actions []string

	if actionsStr != "" {
		// Парсим действия из строки, разделенной запятыми
		for _, action := range splitAndTrim(actionsStr, ",") {
			// Валидируем действие
			if isValidAction(action) {
				actions = append(actions, action)
			}
		}
	}

	// Если действия не указаны, используем resize по умолчанию
	if len(actions) == 0 {
		actions = []string{domain.ResizeAction}
	}

	image := domain.Image{
		FileName: header.Filename,
		FileSize: header.Size,
		Actions:  actions,
		Status:   domain.ImageStatusPending,
	}

	imageID, err := h.usecases.CreateObject(r.Context(), image, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(map[string]string{
		"id":      imageID,
		"status":  domain.ImageStatusPending,
		"message": "Image uploaded successfully",
	})
	if err != nil {
		http.Error(w, "Failed to upload image", http.StatusInternalServerError)
	}
}

// splitAndTrim разбивает строку по разделителю и убирает пробелы
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := []string{}
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	var current string
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// isValidAction проверяет, является ли действие валидным
func isValidAction(action string) bool {
	validActions := []string{
		domain.ResizeAction,
		domain.MiniatureGenerateAction,
		domain.WatermarkAction,
		domain.GrayscaleAction,
	}
	for _, valid := range validActions {
		if action == valid {
			return true
		}
	}
	return false
}

func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]
	if imageID == "" {
		http.Error(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	reader, err := h.usecases.GetObjectByID(r.Context(), imageID)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			log.Printf("Failed to close Kafka connection: %v", err)
		}
	}(reader)

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)

	_, err = io.Copy(w, reader)
	if err != nil {
		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetImageStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]
	if imageID == "" {
		http.Error(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	status, err := h.usecases.GetImageStatus(r.Context(), imageID)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(status)
	if err != nil {
		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
	}
}

func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]
	if imageID == "" {
		http.Error(w, "Image ID is required", http.StatusBadRequest)
		return
	}

	err := h.usecases.RemoveObject(r.Context(), imageID)
	if err != nil {
		http.Error(w, "Failed to delete image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{
		"message": "Image deleted successfully",
	})
	if err != nil {
		http.Error(w, "Failed to serve image", http.StatusInternalServerError)
	}
}
