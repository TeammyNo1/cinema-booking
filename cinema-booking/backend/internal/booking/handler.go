package booking

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cinema-booking/internal/models"
	"cinema-booking/internal/queue"
	"cinema-booking/internal/repository"
	ws "cinema-booking/internal/websocket"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	db        *repository.MongoDB
	redis     *repository.RedisRepo
	hub       *ws.Hub
	publisher *queue.Publisher
}

func NewHandler(db *repository.MongoDB, redis *repository.RedisRepo, hub *ws.Hub, publisher *queue.Publisher) *Handler {
	return &Handler{db: db, redis: redis, hub: hub, publisher: publisher}
}

// ─── GET /api/showtimes ───────────────────────────────────────────────────────

func (h *Handler) ListShowtimes(c *gin.Context) {
	showtimes, err := h.db.FindShowtimes(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, showtimes)
}

// ─── GET /api/showtimes/:id/seats ─────────────────────────────────────────────

func (h *Handler) ListSeats(c *gin.Context) {
	showtimeID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid showtime id"})
		return
	}

	seats, err := h.db.FindSeatsByShowtime(context.Background(), showtimeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich with Redis lock status
	type SeatWithLock struct {
		models.Seat
		LockedBy  string     `json:"locked_by,omitempty"`
		ExpiresAt *time.Time `json:"expires_at,omitempty"`
	}

	enriched := make([]SeatWithLock, 0, len(seats))
	for _, seat := range seats {
		sl := SeatWithLock{Seat: seat}
		if seat.Status == models.SeatLocked {
			owner, _ := h.redis.GetSeatLockOwner(context.Background(), showtimeID.Hex(), seat.SeatCode)
			sl.LockedBy = owner
			ttl, _ := h.redis.GetSeatLockTTL(context.Background(), showtimeID.Hex(), seat.SeatCode)
			if ttl > 0 {
				exp := time.Now().Add(ttl)
				sl.ExpiresAt = &exp
			}
		}
		enriched = append(enriched, sl)
	}

	c.JSON(http.StatusOK, enriched)
}

// ─── POST /api/bookings/lock ──────────────────────────────────────────────────

type LockRequest struct {
	ShowtimeID string `json:"showtime_id" binding:"required"`
	SeatID     string `json:"seat_id"     binding:"required"`
}

func (h *Handler) LockSeat(c *gin.Context) {
	userID := c.GetString("user_id")

	var req LockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	showtimeOID, err := primitive.ObjectIDFromHex(req.ShowtimeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid showtime_id"})
		return
	}
	seatOID, err := primitive.ObjectIDFromHex(req.SeatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seat_id"})
		return
	}

	ctx := context.Background()

	// 1. Verify seat exists and is available in DB
	seat, err := h.db.FindSeatByID(ctx, seatOID)
	if err != nil || seat == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "seat not found"})
		return
	}
	if seat.Status != models.SeatAvailable {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("seat is %s", seat.Status)})
		return
	}

	// 2. Try Redis distributed lock (SET NX EX — atomic)
	locked, expiresAt, err := h.redis.TryLockSeat(ctx, req.ShowtimeID, seat.SeatCode, userID)
	if err != nil {
		h.logAudit(ctx, models.AuditSystemError, userID, "", seat.SeatCode, req.ShowtimeID,
			map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lock service unavailable"})
		return
	}
	if !locked {
		c.JSON(http.StatusConflict, gin.H{"error": "seat already locked by another user"})
		return
	}

	// 3. Update seat status in MongoDB
	if err := h.db.UpdateSeatStatus(ctx, seatOID, models.SeatLocked); err != nil {
		_ = h.redis.ReleaseSeatLock(ctx, req.ShowtimeID, seat.SeatCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update seat status"})
		return
	}

	// 4. Create booking record with LOCKED status
	booking := &models.Booking{
		UserID:     mustParseOID(userID),
		ShowtimeID: showtimeOID,
		SeatID:     seatOID,
		SeatCode:   seat.SeatCode,
		Status:     models.BookingLocked,
		TotalPrice: seat.Price,
		LockedAt:   time.Now(),
		ExpiresAt:  expiresAt,
	}
	if err := h.db.CreateBooking(ctx, booking); err != nil {
		_ = h.db.UpdateSeatStatus(ctx, seatOID, models.SeatAvailable)
		_ = h.redis.ReleaseSeatLock(ctx, req.ShowtimeID, seat.SeatCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
		return
	}

	// 5. Store owner mapping in Redis
	_ = h.redis.SetBookingOwner(ctx, booking.ID.Hex(), userID)

	// 6. Broadcast seat lock to all WebSocket clients
	h.hub.BroadcastToShowtime(req.ShowtimeID, models.WSMessage{
		Event:      models.WSEventSeatUpdate,
		ShowtimeID: req.ShowtimeID,
		SeatCode:   seat.SeatCode,
		Status:     models.SeatLocked,
		LockedBy:   userID,
		ExpiresAt:  &expiresAt,
	})

	// 7. Audit log
	h.logAudit(ctx, models.AuditBookingLocked, userID, booking.ID.Hex(), seat.SeatCode, req.ShowtimeID,
		map[string]interface{}{"expires_at": expiresAt})

	// 8. Schedule auto-release after lock TTL
	go h.scheduleRelease(booking, seat, req.ShowtimeID)

	c.JSON(http.StatusOK, gin.H{
		"booking_id": booking.ID.Hex(),
		"seat_code":  seat.SeatCode,
		"expires_at": expiresAt,
		"status":     booking.Status,
	})
}

// scheduleRelease releases the seat if payment doesn't happen within lock TTL.
func (h *Handler) scheduleRelease(booking *models.Booking, seat *models.Seat, showtimeID string) {
	ttl := time.Until(booking.ExpiresAt)
	if ttl <= 0 {
		return
	}
	time.Sleep(ttl)

	ctx := context.Background()

	// Check if still locked (not yet confirmed)
	currentBooking, err := h.db.FindBookingByID(ctx, booking.ID)
	if err != nil || currentBooking == nil || currentBooking.Status != models.BookingLocked {
		return // already confirmed or cancelled
	}

	log.Printf("[LOCK EXPIRED] booking=%s seat=%s", booking.ID.Hex(), seat.SeatCode)

	// Update booking to TIMEOUT
	_ = h.db.UpdateBookingStatus(ctx, booking.ID, models.BookingTimeout, bson.M{})

	// Release seat in MongoDB
	_ = h.db.UpdateSeatStatus(ctx, seat.ID, models.SeatAvailable)

	// Release Redis lock (may already be expired, that's OK)
	_ = h.redis.ReleaseSeatLock(ctx, showtimeID, seat.SeatCode)

	// Broadcast seat release
	h.hub.BroadcastToShowtime(showtimeID, models.WSMessage{
		Event:      models.WSEventSeatRelease,
		ShowtimeID: showtimeID,
		SeatCode:   seat.SeatCode,
		Status:     models.SeatAvailable,
	})

	// Publish to MQ
	_ = h.publisher.PublishBookingTimeout(currentBooking)
	h.logAudit(ctx, models.AuditBookingTimeout, booking.UserID.Hex(), booking.ID.Hex(), seat.SeatCode, showtimeID, nil)
	h.logAudit(ctx, models.AuditSeatReleased, booking.UserID.Hex(), booking.ID.Hex(), seat.SeatCode, showtimeID, nil)
}

// ─── POST /api/bookings/:id/confirm ──────────────────────────────────────────

func (h *Handler) ConfirmBooking(c *gin.Context) {
	userID := c.GetString("user_id")
	bookingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	ctx := context.Background()

	booking, err := h.db.FindBookingByID(ctx, bookingID)
	if err != nil || booking == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	// Security: only the owner can confirm
	if booking.UserID.Hex() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your booking"})
		return
	}

	if booking.Status != models.BookingLocked {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("cannot confirm booking with status: %s", booking.Status)})
		return
	}

	// Check lock is still valid
	owner, _ := h.redis.GetSeatLockOwner(ctx, booking.ShowtimeID.Hex(), booking.SeatCode)
	if owner != userID {
		c.JSON(http.StatusConflict, gin.H{"error": "seat lock expired"})
		return
	}

	// Confirm: update seat → BOOKED, booking → CONFIRMED
	now := time.Now()
	if err := h.db.UpdateBookingStatus(ctx, bookingID, models.BookingConfirmed, bson.M{"confirmed_at": now}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to confirm booking"})
		return
	}
	_ = h.db.UpdateSeatStatus(ctx, booking.SeatID, models.SeatBooked)
	_ = h.redis.ReleaseSeatLock(ctx, booking.ShowtimeID.Hex(), booking.SeatCode)

	// Broadcast BOOKED status
	h.hub.BroadcastToShowtime(booking.ShowtimeID.Hex(), models.WSMessage{
		Event:      models.WSEventBookingConfirm,
		ShowtimeID: booking.ShowtimeID.Hex(),
		SeatCode:   booking.SeatCode,
		Status:     models.SeatBooked,
	})

	// Publish success event to MQ (async)
	booking.Status = models.BookingConfirmed
	go func() {
		_ = h.publisher.PublishBookingConfirmed(booking)
		_ = h.publisher.PublishNotification("user@example.com",
			"Booking Confirmed!",
			fmt.Sprintf("Your seat %s is confirmed. Enjoy the movie!", booking.SeatCode))
	}()

	// Audit log
	h.logAudit(ctx, models.AuditBookingSuccess, userID, bookingID.Hex(), booking.SeatCode, booking.ShowtimeID.Hex(), nil)

	c.JSON(http.StatusOK, gin.H{
		"message":     "booking confirmed",
		"booking_id":  bookingID.Hex(),
		"seat_code":   booking.SeatCode,
		"total_price": booking.TotalPrice,
		"status":      models.BookingConfirmed,
	})
}

// ─── DELETE /api/bookings/:id ─────────────────────────────────────────────────

func (h *Handler) CancelBooking(c *gin.Context) {
	userID := c.GetString("user_id")
	bookingID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking id"})
		return
	}

	ctx := context.Background()
	booking, err := h.db.FindBookingByID(ctx, bookingID)
	if err != nil || booking == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	if booking.UserID.Hex() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your booking"})
		return
	}

	if booking.Status != models.BookingLocked {
		c.JSON(http.StatusConflict, gin.H{"error": "can only cancel locked bookings"})
		return
	}

	_ = h.db.UpdateBookingStatus(ctx, bookingID, models.BookingCancelled, bson.M{})
	_ = h.db.UpdateSeatStatus(ctx, booking.SeatID, models.SeatAvailable)
	_ = h.redis.ReleaseSeatLock(ctx, booking.ShowtimeID.Hex(), booking.SeatCode)

	h.hub.BroadcastToShowtime(booking.ShowtimeID.Hex(), models.WSMessage{
		Event:      models.WSEventSeatRelease,
		ShowtimeID: booking.ShowtimeID.Hex(),
		SeatCode:   booking.SeatCode,
		Status:     models.SeatAvailable,
	})

	h.logAudit(ctx, models.AuditSeatReleased, userID, bookingID.Hex(), booking.SeatCode, booking.ShowtimeID.Hex(),
		map[string]interface{}{"reason": "user_cancelled"})

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled"})
}

// ─── GET /api/bookings/me ─────────────────────────────────────────────────────

func (h *Handler) MyBookings(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	bookings, err := h.db.FindUserBookings(context.Background(), oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (h *Handler) logAudit(ctx context.Context, eventType models.AuditEventType,
	userID, bookingID, seatCode, showtimeID string, details map[string]interface{}) {
	entry := &models.AuditLog{
		EventType:  eventType,
		UserID:     userID,
		BookingID:  bookingID,
		SeatCode:   seatCode,
		ShowtimeID: showtimeID,
		Details:    details,
	}
	go func() {
		if err := h.db.InsertAuditLog(context.Background(), entry); err != nil {
			log.Printf("audit log insert error: %v", err)
		}
		_ = h.publisher.PublishAuditLog(entry)
	}()
}

func mustParseOID(hex string) primitive.ObjectID {
	oid, _ := primitive.ObjectIDFromHex(hex)
	return oid
}
