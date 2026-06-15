package admin

import (
	"context"
	"net/http"
	"strconv"

	"cinema-booking/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *repository.MongoDB
}

func NewHandler(db *repository.MongoDB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ListBookings(c *gin.Context) {
	status     := c.Query("status")
	showtimeID := c.Query("showtime_id")
	limit      := int64(parseQ(c, "limit", 50))
	page       := int64(parseQ(c, "page", 1))
	skip       := (page - 1) * limit

	bookings, total, err := h.db.FindBookingsForAdmin(context.Background(), status, showtimeID, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": bookings, "total": total, "page": page, "limit": limit})
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	eventType := c.Query("event_type")
	bookingID := c.Query("booking_id")
	limit     := int64(parseQ(c, "limit", 50))
	page      := int64(parseQ(c, "page", 1))
	skip      := (page - 1) * limit

	logs, total, err := h.db.FindAuditLogs(context.Background(), eventType, bookingID, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs, "total": total, "page": page, "limit": limit})
}

func (h *Handler) Stats(c *gin.Context) {
	ctx := context.Background()
	_, total,     _ := h.db.FindBookingsForAdmin(ctx, "", "", 0, 0)
	_, confirmed, _ := h.db.FindBookingsForAdmin(ctx, "CONFIRMED", "", 0, 0)
	_, locked,    _ := h.db.FindBookingsForAdmin(ctx, "LOCKED", "", 0, 0)
	_, timeout,   _ := h.db.FindBookingsForAdmin(ctx, "TIMEOUT", "", 0, 0)

	c.JSON(http.StatusOK, gin.H{
		"total_bookings":     total,
		"confirmed_bookings": confirmed,
		"locked_bookings":    locked,
		"timeout_bookings":   timeout,
	})
}

func parseQ(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" { return def }
	n, err := strconv.Atoi(v)
	if err != nil { return def }
	return n
}
