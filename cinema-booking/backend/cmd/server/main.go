package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"cinema-booking/config"
	adminHandler "cinema-booking/internal/admin"
	"cinema-booking/internal/auth"
	"cinema-booking/internal/booking"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/queue"
	"cinema-booking/internal/repository"
	showtimeHandler "cinema-booking/internal/showtime"
	ws "cinema-booking/internal/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	mongo, err := repository.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("MongoDB init failed: %v", err)
	}
	defer mongo.Close(context.Background())

	if err := mongo.SeedIfEmpty(context.Background()); err != nil {
		log.Printf("Seed warning: %v", err)
	}

	redis, err := repository.NewRedisRepo(cfg)
	if err != nil {
		log.Fatalf("Redis init failed: %v", err)
	}

	var publisher *queue.Publisher
	for i := 0; i < 10; i++ {
		publisher, err = queue.NewPublisher(cfg)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ not ready, retrying... (%d/10)", i+1)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("RabbitMQ init failed: %v", err)
	}
	defer publisher.Close()

	consumer, err := queue.NewConsumer(cfg)
	if err != nil {
		log.Printf("RabbitMQ consumer warning: %v", err)
	} else {
		consumer.ConsumeNotifications()
		consumer.ConsumeBookingEvents()
		defer consumer.Close()
	}

	authHandler := auth.NewHandler(cfg, mongo)
	wsHub := ws.NewHub(authHandler)
	bookingHdl := booking.NewHandler(mongo, redis, wsHub, publisher)
	adminHdl := adminHandler.NewHandler(mongo)
	showtimeHdl := showtimeHandler.NewHandler(mongo)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL, "http://localhost:5173", "http://localhost:3000", "http://localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now()})
	})

	// ── Auth ──────────────────────────────────────────────────────────────────
	apiAuth := r.Group("/api/auth")
	{
		apiAuth.GET("/google", authHandler.GoogleLogin)
		apiAuth.GET("/google/callback", authHandler.GoogleCallback)
		apiAuth.GET("/me", middleware.AuthRequired(authHandler), authHandler.Me)
	}

	// ── WebSocket (public) ────────────────────────────────────────────────────
	r.GET("/ws/showtimes/:showtimeId", wsHub.HandleWS)

	// ── PUBLIC routes (no auth required) ─────────────────────────────────────
	public := r.Group("/api")
	{
		public.GET("/showtimes", showtimeHdl.ListShowtimes)          // ✅ public
		public.GET("/showtimes/:id/seats", bookingHdl.ListSeats)     // ✅ public (read-only)
	}

	// ── User API (auth required) ──────────────────────────────────────────────
	api := r.Group("/api", middleware.AuthRequired(authHandler))
	{
		api.POST("/bookings/lock", bookingHdl.LockSeat)
		api.POST("/bookings/:id/confirm", bookingHdl.ConfirmBooking)
		api.DELETE("/bookings/:id", bookingHdl.CancelBooking)
		api.GET("/bookings/me", bookingHdl.MyBookings)
	}

	// ── Admin API (auth + admin role) ─────────────────────────────────────────
	adminAPI := r.Group("/api/admin", middleware.AuthRequired(authHandler), middleware.AdminRequired())
	{
		adminAPI.GET("/bookings", adminHdl.ListBookings)
		adminAPI.GET("/audit-logs", adminHdl.ListAuditLogs)
		adminAPI.GET("/stats", adminHdl.Stats)

		adminAPI.GET("/showtimes", showtimeHdl.ListShowtimes)
		adminAPI.GET("/showtimes/:id", showtimeHdl.GetShowtime)
		adminAPI.POST("/showtimes", showtimeHdl.CreateShowtime)
		adminAPI.PUT("/showtimes/:id", showtimeHdl.UpdateShowtime)
		adminAPI.DELETE("/showtimes/:id", showtimeHdl.DeleteShowtime)
	}

	log.Printf("🎬 Cinema Booking API running on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
