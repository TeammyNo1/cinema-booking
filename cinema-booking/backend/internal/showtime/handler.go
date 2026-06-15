package showtime

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cinema-booking/internal/models"
	"cinema-booking/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	db *repository.MongoDB
}

func NewHandler(db *repository.MongoDB) *Handler {
	return &Handler{db: db}
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

type CreateShowtimeRequest struct {
	MovieName   string  `json:"movie_name"   binding:"required"`
	Hall        string  `json:"hall"         binding:"required"`
	StartTime   string  `json:"start_time"   binding:"required"` // RFC3339
	EndTime     string  `json:"end_time"     binding:"required"`
	PosterEmoji string  `json:"poster_emoji"`
	Genre       string  `json:"genre"`
	Rating      string  `json:"rating"`
	Duration    int     `json:"duration"`     // minutes
	PriceNormal float64 `json:"price_normal"` // default 250
	PriceVIP    float64 `json:"price_vip"`    // default 350
	RowsCount   int     `json:"rows_count"`   // default 8
	SeatsPerRow int     `json:"seats_per_row"` // default 12
}

// ─── GET /api/showtimes (public) ──────────────────────────────────────────────

func (h *Handler) ListShowtimes(c *gin.Context) {
	showtimes, err := h.db.FindShowtimes(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, showtimes)
}

// ─── POST /api/admin/showtimes ────────────────────────────────────────────────

func (h *Handler) CreateShowtime(c *gin.Context) {
	var req CreateShowtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format (use RFC3339)"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format (use RFC3339)"})
		return
	}

	// Defaults
	if req.PriceNormal == 0 {
		req.PriceNormal = 250
	}
	if req.PriceVIP == 0 {
		req.PriceVIP = 350
	}
	if req.RowsCount == 0 {
		req.RowsCount = 8
	}
	if req.SeatsPerRow == 0 {
		req.SeatsPerRow = 12
	}
	if req.PosterEmoji == "" {
		req.PosterEmoji = "🎬"
	}
	if req.Genre == "" {
		req.Genre = "Action"
	}
	if req.Rating == "" {
		req.Rating = "G"
	}
	if req.Duration == 0 {
		req.Duration = 120
	}

	st := &models.Showtime{
		MovieName:   req.MovieName,
		Hall:        req.Hall,
		StartTime:   startTime,
		EndTime:     endTime,
		PosterEmoji: req.PosterEmoji,
		Genre:       req.Genre,
		Rating:      req.Rating,
		Duration:    req.Duration,
		CreatedAt:   time.Now(),
	}

	id, err := h.db.CreateShowtime(context.Background(), st)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create seats for this showtime
	rows := generateRows(req.RowsCount)
	if err := h.db.CreateSeatsForShowtime(context.Background(), id, rows, req.SeatsPerRow, req.PriceNormal, req.PriceVIP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "showtime created but seats failed: " + err.Error()})
		return
	}

	st.ID = id
	c.JSON(http.StatusCreated, st)
}

// ─── PUT /api/admin/showtimes/:id ─────────────────────────────────────────────

func (h *Handler) UpdateShowtime(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req CreateShowtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time"})
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time"})
		return
	}

	update := &models.Showtime{
		MovieName:   req.MovieName,
		Hall:        req.Hall,
		StartTime:   startTime,
		EndTime:     endTime,
		PosterEmoji: req.PosterEmoji,
		Genre:       req.Genre,
		Rating:      req.Rating,
		Duration:    req.Duration,
	}

	if err := h.db.UpdateShowtime(context.Background(), id, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "showtime updated", "id": id.Hex()})
}

// ─── DELETE /api/admin/showtimes/:id ─────────────────────────────────────────

func (h *Handler) DeleteShowtime(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := context.Background()

	// Check no active bookings
	bookings, _, err := h.db.FindBookingsForAdmin(ctx, "LOCKED", id.Hex(), 1, 0)
	if err == nil && len(bookings) > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete showtime with active locked bookings"})
		return
	}

	if err := h.db.DeleteShowtime(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Also delete seats
	_ = h.db.DeleteSeatsByShowtime(ctx, id)

	c.JSON(http.StatusOK, gin.H{"message": "showtime deleted"})
}

// ─── GET /api/admin/showtimes/:id ────────────────────────────────────────────

func (h *Handler) GetShowtime(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	st, err := h.db.FindShowtimeByID(context.Background(), id)
	if err != nil || st == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, st)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func generateRows(count int) []string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rows := make([]string, count)
	for i := 0; i < count && i < len(letters); i++ {
		rows[i] = fmt.Sprintf("%c", letters[i])
	}
	return rows
}
