package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cinema-booking/config"
	"cinema-booking/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	QueueBookingEvents = "booking_events"
	QueueAuditLogs     = "audit_logs"
	QueueNotifications = "notifications"
)

type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(cfg *config.Config) (*Publisher, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	// Declare queues
	queues := []string{QueueBookingEvents, QueueAuditLogs, QueueNotifications}
	for _, q := range queues {
		if _, err := ch.QueueDeclare(q, true, false, false, false, nil); err != nil {
			return nil, fmt.Errorf("declare queue %s: %w", q, err)
		}
	}

	return &Publisher{conn: conn, channel: ch}, nil
}

func (p *Publisher) Publish(queue string, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.channel.PublishWithContext(
		context.Background(),
		"", queue, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
}

func (p *Publisher) PublishBookingConfirmed(booking *models.Booking) error {
	return p.Publish(QueueBookingEvents, Event{
		Type:      "BOOKING_CONFIRMED",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"booking_id":  booking.ID.Hex(),
			"user_id":     booking.UserID.Hex(),
			"showtime_id": booking.ShowtimeID.Hex(),
			"seat_code":   booking.SeatCode,
			"total_price": booking.TotalPrice,
		},
	})
}

func (p *Publisher) PublishBookingTimeout(booking *models.Booking) error {
	return p.Publish(QueueBookingEvents, Event{
		Type:      "BOOKING_TIMEOUT",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"booking_id":  booking.ID.Hex(),
			"user_id":     booking.UserID.Hex(),
			"seat_code":   booking.SeatCode,
			"showtime_id": booking.ShowtimeID.Hex(),
		},
	})
}

func (p *Publisher) PublishAuditLog(logEntry *models.AuditLog) error {
	return p.Publish(QueueAuditLogs, Event{
		Type:      string(logEntry.EventType),
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"user_id":    logEntry.UserID,
			"booking_id": logEntry.BookingID,
			"seat_code":  logEntry.SeatCode,
			"details":    logEntry.Details,
		},
	})
}

func (p *Publisher) PublishNotification(userEmail, subject, body string) error {
	return p.Publish(QueueNotifications, Event{
		Type:      "EMAIL_NOTIFICATION",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"email":   userEmail,
			"subject": subject,
			"body":    body,
		},
	})
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

// ─── Consumer ─────────────────────────────────────────────────────────────────

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConsumer(cfg *config.Config) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	return &Consumer{conn: conn, channel: ch}, nil
}

// ConsumeNotifications is a mock consumer that logs notifications.
func (c *Consumer) ConsumeNotifications() {
	msgs, err := c.channel.Consume(QueueNotifications, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume notifications: %v", err)
		return
	}
	go func() {
		for msg := range msgs {
			var event Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				continue
			}
			// Mock: just log the notification
			log.Printf("[NOTIFICATION] To: %v | Subject: %v",
				event.Payload["email"], event.Payload["subject"])
		}
	}()
}

// ConsumeBookingEvents logs booking events.
func (c *Consumer) ConsumeBookingEvents() {
	msgs, err := c.channel.Consume(QueueBookingEvents, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("Failed to consume booking events: %v", err)
		return
	}
	go func() {
		for msg := range msgs {
			var event Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				continue
			}
			log.Printf("[BOOKING EVENT] Type: %s | Payload: %v", event.Type, event.Payload)
		}
	}()
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
