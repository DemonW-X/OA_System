package handlers

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"oa-system/database"
	"oa-system/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultOperationLogMQURL = "amqp://guest:guest@localhost:5672/"
	defaultOperationLogQueue = "oa.operation_log"
)

type operationLogMessage struct {
	UserID     int    `json:"user_id"`
	Username   string `json:"username"`
	Module     string `json:"module"`
	Action     string `json:"action"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	IP         string `json:"ip"`
	Remark     string `json:"remark"`
	CreatedAt  int64  `json:"created_at"`
}

var (
	operationLogMQURL = getenv("RABBITMQ_URL", defaultOperationLogMQURL)
	operationLogQueue = getenv("OPERATION_LOG_QUEUE", defaultOperationLogQueue)

	publisherConn *amqp.Connection
	publisherCh   *amqp.Channel
	publisherMu   sync.RWMutex

	operationLogMQEnabled bool
)

func InitOperationLogAsync() {
	if !initOperationLogPublisher() {
		log.Printf("[operation_log] rabbitmq unavailable, fallback to sync db logging")
		return
	}

	operationLogMQEnabled = true
	go consumeOperationLogQueue()
}

func enqueueOperationLog(entry models.OperationLog) bool {
	if !operationLogMQEnabled {
		return false
	}

	msg := operationLogMessage{
		UserID:     entry.UserID,
		Username:   entry.Username,
		Module:     entry.Module,
		Action:     entry.Action,
		Method:     entry.Method,
		Path:       entry.Path,
		StatusCode: entry.StatusCode,
		IP:         entry.IP,
		Remark:     entry.Remark,
		CreatedAt:  time.Now().UnixMilli(),
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return false
	}

	if publishToRabbit(body) {
		return true
	}

	log.Printf("[operation_log] publish failed, retry once after reconnect")
	if !initOperationLogPublisher() {
		return false
	}
	return publishToRabbit(body)
}

func initOperationLogPublisher() bool {
	conn, ch, err := openRabbitPublisher()
	if err != nil {
		log.Printf("[operation_log] init publisher failed: %v", err)
		return false
	}

	publisherMu.Lock()
	defer publisherMu.Unlock()

	if publisherCh != nil {
		_ = publisherCh.Close()
	}
	if publisherConn != nil {
		_ = publisherConn.Close()
	}
	publisherConn = conn
	publisherCh = ch
	return true
}

func openRabbitPublisher() (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(operationLogMQURL)
	if err != nil {
		return nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, err
	}

	_, err = ch.QueueDeclare(
		operationLogQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, err
	}

	return conn, ch, nil
}

func publishToRabbit(body []byte) bool {
	publisherMu.RLock()
	ch := publisherCh
	publisherMu.RUnlock()

	if ch == nil {
		return false
	}

	err := ch.Publish(
		"",
		operationLogQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		log.Printf("[operation_log] publish error: %v", err)
		return false
	}
	return true
}

func consumeOperationLogQueue() {
	for {
		conn, ch, deliveries, err := openRabbitConsumer()
		if err != nil {
			log.Printf("[operation_log] init consumer failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for d := range deliveries {
			entry, ok := decodeOperationLogMessage(d.Body)
			if !ok {
				_ = d.Ack(false)
				continue
			}

			if err := database.DB.Create(&entry).Error; err != nil {
				log.Printf("[operation_log] persist failed: %v", err)
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}

		_ = ch.Close()
		_ = conn.Close()
		time.Sleep(time.Second)
	}
}

func openRabbitConsumer() (*amqp.Connection, *amqp.Channel, <-chan amqp.Delivery, error) {
	conn, err := amqp.Dial(operationLogMQURL)
	if err != nil {
		return nil, nil, nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, nil, nil, err
	}

	_, err = ch.QueueDeclare(
		operationLogQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, nil, err
	}

	if err := ch.Qos(50, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, nil, err
	}

	deliveries, err := ch.Consume(
		operationLogQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, nil, nil, err
	}

	return conn, ch, deliveries, nil
}

func decodeOperationLogMessage(body []byte) (models.OperationLog, bool) {
	var msg operationLogMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("[operation_log] decode failed: %v", err)
		return models.OperationLog{}, false
	}

	entry := models.OperationLog{
		UserID:     msg.UserID,
		Username:   msg.Username,
		Module:     msg.Module,
		Action:     msg.Action,
		Method:     msg.Method,
		Path:       msg.Path,
		StatusCode: msg.StatusCode,
		IP:         msg.IP,
		Remark:     msg.Remark,
	}
	if msg.CreatedAt > 0 {
		entry.CreatedAt = time.UnixMilli(msg.CreatedAt)
	} else {
		entry.CreatedAt = time.Now()
	}

	return entry, true
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
