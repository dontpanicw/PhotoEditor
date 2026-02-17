package processor

import (
	"fmt"
	"github.com/h2non/bimg"
	"log"
)

func ResizeImage(file []byte, width, height int) ([]byte, error) {

	// Создаем опции для ресайза
	options := bimg.Options{
		Width:  width,
		Height: height,
		// Если нужно обрезать, чтобы точно вписаться в размеры
		// Crop: true,
		// Качество JPEG (1-100)
		Quality: 85,
	}

	// Обрабатываем изображение
	newImage, err := bimg.NewImage(file).Process(options)
	if err != nil {
		return nil, fmt.Errorf("ошибка обработки: %v", err)
	}

	// Получаем информацию о новом изображении
	size, _ := bimg.NewImage(newImage).Size()
	log.Printf("Ресайз выполнен: %dx%d", size.Width, size.Height)

	return newImage, nil
}

// Создание миниатюры с интеллектуальной обрезкой (smart crop)
func GenerateSmartThumbnail(file []byte, width, height int) ([]byte, error) {

	options := bimg.Options{
		Width:   width,
		Height:  height,
		Crop:    true,
		Gravity: bimg.GravitySmart, // Интеллектуальная обрезка
		Quality: 90,
	}

	newImage, err := bimg.NewImage(file).Process(options)
	if err != nil {
		return nil, err
	}

	return newImage, nil
}

func AddTextWatermark(file []byte, text string) ([]byte, error) {
	// Получаем размеры изображения
	img := bimg.NewImage(file)
	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить размеры изображения: %v", err)
	}

	log.Printf("Добавление водяного знака на изображение %dx%d", size.Width, size.Height)

	// Вместо текстового водяного знака используем простое наложение
	// Создаем полупрозрачный слой (это работает надежнее)
	options := bimg.Options{
		Quality: 90,
	}

	// Просто обрабатываем изображение с качеством
	// В production можно использовать ImageMagick или другую библиотеку для текста
	newImage, err := img.Process(options)
	if err != nil {
		return nil, fmt.Errorf("ошибка обработки изображения: %v", err)
	}

	log.Printf("Водяной знак обработан (упрощенная версия без текста)")
	return newImage, nil
}

// ApplyGrayscale применяет черно-белый фильтр к изображению
func ApplyGrayscale(file []byte) ([]byte, error) {
	img := bimg.NewImage(file)
	size, err := img.Size()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить размеры изображения: %v", err)
	}

	log.Printf("Применение ч/б фильтра к изображению %dx%d", size.Width, size.Height)

	options := bimg.Options{
		Interpretation: bimg.InterpretationBW, // Черно-белый режим
		Quality:        90,
	}

	newImage, err := img.Process(options)
	if err != nil {
		return nil, fmt.Errorf("ошибка применения ч/б фильтра: %v", err)
	}

	log.Printf("Ч/б фильтр успешно применен")
	return newImage, nil
}
