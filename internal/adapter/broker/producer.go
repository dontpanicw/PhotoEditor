package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dontpanicw/ImageProcessor/config"
	"github.com/dontpanicw/ImageProcessor/internal/domain"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
	config *config.Config
}

func NewProducer(cfg *config.Config) *Producer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.KafkaBrokers...),
		Topic:                  cfg.KafkaTaskTopic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		Async:                  false,
		WriteTimeout:           10 * time.Second,
		ReadTimeout:            10 * time.Second,
		AllowAutoTopicCreation: true, // Разрешаем автосоздание топика
	}

	// Пытаемся создать топик явно
	go func() {
		time.Sleep(2 * time.Second) // Даем Kafka время на инициализацию
		conn, err := kafka.Dial("tcp", cfg.KafkaBrokers[0])
		if err != nil {
			log.Printf("Failed to dial Kafka for topic creation: %v", err)
			return
		}
		defer conn.Close()

		controller, err := conn.Controller()
		if err != nil {
			log.Printf("Failed to get Kafka controller: %v", err)
			return
		}

		controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
		if err != nil {
			log.Printf("Failed to dial Kafka controller: %v", err)
			return
		}
		defer controllerConn.Close()

		topicConfigs := []kafka.TopicConfig{
			{
				Topic:             cfg.KafkaTaskTopic,
				NumPartitions:     3,
				ReplicationFactor: 1,
			},
		}

		err = controllerConn.CreateTopics(topicConfigs...)
		if err != nil {
			log.Printf("Topic creation info: %v (this is OK if topic already exists)", err)
		} else {
			log.Printf("Successfully created topic: %s", cfg.KafkaTaskTopic)
		}
	}()

	return &Producer{
		writer: writer,
		config: cfg,
	}
}

func (p *Producer) SendMessage(ctx context.Context, imageId string, actions []string) error {
	msg := domain.TaskMessage{
		ImageID:   imageId,
		Actions:   actions,
		Timestamp: time.Now().Unix(),
	}

	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	key := []byte(imageId)

	log.Printf("Attempting to send message to Kafka: imageId=%s, actions=%v", imageId, actions)

	// Создаем контекст с таймаутом
	sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Отправляем сообщение
	err = p.writer.WriteMessages(sendCtx, kafka.Message{
		Key:   key,
		Value: value,
	})

	if err != nil {
		log.Printf("ERROR: Failed to send message to Kafka: %v", err)
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Successfully sent message to Kafka for image: %s", imageId)
	return nil
}

func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
