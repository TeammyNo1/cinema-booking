package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GoogleID  string             `bson:"google_id"     json:"google_id"`
	Email     string             `bson:"email"         json:"email"`
	Name      string             `bson:"name"          json:"name"`
	Avatar    string             `bson:"avatar"        json:"avatar"`
	Role      Role               `bson:"role"          json:"role"`
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
}

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatLocked    SeatStatus = "LOCKED"
	SeatBooked    SeatStatus = "BOOKED"
)

type SeatType string

const (
	SeatTypeNormal  SeatType = "NORMAL"
	SeatTypeVIP     SeatType = "VIP"
	SeatTypeCouple  SeatType = "COUPLE"
)

type Seat struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShowtimeID primitive.ObjectID `bson:"showtime_id"   json:"showtime_id"`
	Row        string             `bson:"row"           json:"row"`
	Number     int                `bson:"number"        json:"number"`
	SeatCode   string             `bson:"seat_code"     json:"seat_code"`
	Status     SeatStatus         `bson:"status"        json:"status"`
	Type       SeatType           `bson:"type"          json:"type"`
	Price      float64            `bson:"price"         json:"price"`
}

type Showtime struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MovieName   string             `bson:"movie_name"    json:"movie_name"`
	Hall        string             `bson:"hall"          json:"hall"`
	StartTime   time.Time          `bson:"start_time"    json:"start_time"`
	EndTime     time.Time          `bson:"end_time"      json:"end_time"`
	PosterEmoji string             `bson:"poster_emoji"  json:"poster_emoji"`
	Genre       string             `bson:"genre"         json:"genre"`
	Rating      string             `bson:"rating"        json:"rating"`
	Duration    int                `bson:"duration"      json:"duration"`
	CreatedAt   time.Time          `bson:"created_at"    json:"created_at"`
}

type BookingStatus string

const (
	BookingLocked    BookingStatus = "LOCKED"
	BookingConfirmed BookingStatus = "CONFIRMED"
	BookingCancelled BookingStatus = "CANCELLED"
	BookingTimeout   BookingStatus = "TIMEOUT"
)

type Booking struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id"       json:"user_id"`
	ShowtimeID  primitive.ObjectID `bson:"showtime_id"   json:"showtime_id"`
	SeatID      primitive.ObjectID `bson:"seat_id"       json:"seat_id"`
	SeatCode    string             `bson:"seat_code"     json:"seat_code"`
	Status      BookingStatus      `bson:"status"        json:"status"`
	TotalPrice  float64            `bson:"total_price"   json:"total_price"`
	LockedAt    time.Time          `bson:"locked_at"     json:"locked_at"`
	ExpiresAt   time.Time          `bson:"expires_at"    json:"expires_at"`
	ConfirmedAt *time.Time         `bson:"confirmed_at"  json:"confirmed_at,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"    json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"    json:"updated_at"`
}

type AuditEventType string

const (
	AuditBookingSuccess AuditEventType = "BOOKING_SUCCESS"
	AuditBookingTimeout AuditEventType = "BOOKING_TIMEOUT"
	AuditSeatReleased   AuditEventType = "SEAT_RELEASED"
	AuditSystemError    AuditEventType = "SYSTEM_ERROR"
	AuditBookingLocked  AuditEventType = "BOOKING_LOCKED"
)

type AuditLog struct {
	ID         primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	EventType  AuditEventType         `bson:"event_type"    json:"event_type"`
	UserID     string                 `bson:"user_id"       json:"user_id"`
	BookingID  string                 `bson:"booking_id"    json:"booking_id"`
	SeatCode   string                 `bson:"seat_code"     json:"seat_code"`
	ShowtimeID string                 `bson:"showtime_id"   json:"showtime_id"`
	Details    map[string]interface{} `bson:"details"       json:"details"`
	CreatedAt  time.Time              `bson:"created_at"    json:"created_at"`
}

type WSEventType string

const (
	WSEventSeatUpdate     WSEventType = "SEAT_UPDATE"
	WSEventBookingLock    WSEventType = "BOOKING_LOCK"
	WSEventBookingConfirm WSEventType = "BOOKING_CONFIRM"
	WSEventSeatRelease    WSEventType = "SEAT_RELEASE"
)

type WSMessage struct {
	Event      WSEventType `json:"event"`
	ShowtimeID string      `json:"showtime_id"`
	SeatCode   string      `json:"seat_code"`
	Status     SeatStatus  `json:"status"`
	LockedBy   string      `json:"locked_by,omitempty"`
	ExpiresAt  *time.Time  `json:"expires_at,omitempty"`
}
