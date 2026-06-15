package repository

import (
	"context"
	"fmt"
	"time"

	"cinema-booking/config"
	"cinema-booking/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping: %w", err)
	}
	repo := &MongoDB{client: client, db: client.Database(cfg.MongoDB)}
	repo.ensureIndexes(ctx)
	return repo, nil
}

func (m *MongoDB) ensureIndexes(ctx context.Context) {
	m.db.Collection("bookings").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "showtime_id", Value: 1}, {Key: "seat_id", Value: 1}, {Key: "status", Value: 1}},
	})
	m.db.Collection("audit_logs").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "created_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(30 * 24 * 3600),
	})
}

func (m *MongoDB) Close(ctx context.Context) error { return m.client.Disconnect(ctx) }

// ─── Users ────────────────────────────────────────────────────────────────────

func (m *MongoDB) UpsertUser(ctx context.Context, user *models.User) error {
	_, err := m.db.Collection("users").UpdateOne(ctx,
		bson.M{"google_id": user.GoogleID},
		bson.M{"$set": user},
		options.Update().SetUpsert(true))
	return err
}

func (m *MongoDB) FindUserByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	var user models.User
	err := m.db.Collection("users").FindOne(ctx, bson.M{"google_id": googleID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (m *MongoDB) FindUserByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := m.db.Collection("users").FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

// ─── Showtimes ────────────────────────────────────────────────────────────────

func (m *MongoDB) FindShowtimes(ctx context.Context) ([]models.Showtime, error) {
	cursor, err := m.db.Collection("showtimes").Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var result []models.Showtime
	return result, cursor.All(ctx, &result)
}

func (m *MongoDB) FindShowtimeByID(ctx context.Context, id primitive.ObjectID) (*models.Showtime, error) {
	var st models.Showtime
	err := m.db.Collection("showtimes").FindOne(ctx, bson.M{"_id": id}).Decode(&st)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &st, err
}

func (m *MongoDB) CreateShowtime(ctx context.Context, st *models.Showtime) (primitive.ObjectID, error) {
	st.ID = primitive.NewObjectID()
	st.CreatedAt = time.Now()
	_, err := m.db.Collection("showtimes").InsertOne(ctx, st)
	return st.ID, err
}

func (m *MongoDB) UpdateShowtime(ctx context.Context, id primitive.ObjectID, st *models.Showtime) error {
	_, err := m.db.Collection("showtimes").UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{
			"movie_name":   st.MovieName,
			"hall":         st.Hall,
			"start_time":   st.StartTime,
			"end_time":     st.EndTime,
			"poster_emoji": st.PosterEmoji,
			"genre":        st.Genre,
			"rating":       st.Rating,
			"duration":     st.Duration,
		}})
	return err
}

func (m *MongoDB) DeleteShowtime(ctx context.Context, id primitive.ObjectID) error {
	_, err := m.db.Collection("showtimes").DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// ─── Seats ────────────────────────────────────────────────────────────────────

func (m *MongoDB) FindSeatsByShowtime(ctx context.Context, showtimeID primitive.ObjectID) ([]models.Seat, error) {
	cursor, err := m.db.Collection("seats").Find(ctx,
		bson.M{"showtime_id": showtimeID},
		options.Find().SetSort(bson.D{{Key: "row", Value: 1}, {Key: "number", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var result []models.Seat
	return result, cursor.All(ctx, &result)
}

func (m *MongoDB) FindSeatByID(ctx context.Context, seatID primitive.ObjectID) (*models.Seat, error) {
	var seat models.Seat
	err := m.db.Collection("seats").FindOne(ctx, bson.M{"_id": seatID}).Decode(&seat)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &seat, err
}

func (m *MongoDB) UpdateSeatStatus(ctx context.Context, seatID primitive.ObjectID, status models.SeatStatus) error {
	_, err := m.db.Collection("seats").UpdateOne(ctx,
		bson.M{"_id": seatID},
		bson.M{"$set": bson.M{"status": status}})
	return err
}

func (m *MongoDB) DeleteSeatsByShowtime(ctx context.Context, showtimeID primitive.ObjectID) error {
	_, err := m.db.Collection("seats").DeleteMany(ctx, bson.M{"showtime_id": showtimeID})
	return err
}

func (m *MongoDB) CreateSeatsForShowtime(ctx context.Context, showtimeID primitive.ObjectID, rows []string, seatsPerRow int, priceNormal, priceVIP float64) error {
	var docs []interface{}
	vipRows := map[string]bool{}
	// First 2 rows are VIP
	if len(rows) >= 2 {
		vipRows[rows[0]] = true
		vipRows[rows[1]] = true
	}
	for _, row := range rows {
		seatType := models.SeatTypeNormal
		price := priceNormal
		if vipRows[row] {
			seatType = models.SeatTypeVIP
			price = priceVIP
		}
		for num := 1; num <= seatsPerRow; num++ {
			docs = append(docs, models.Seat{
				ID:         primitive.NewObjectID(),
				ShowtimeID: showtimeID,
				Row:        row,
				Number:     num,
				SeatCode:   fmt.Sprintf("%s%d", row, num),
				Status:     models.SeatAvailable,
				Type:       seatType,
				Price:      price,
			})
		}
	}
	_, err := m.db.Collection("seats").InsertMany(ctx, docs)
	return err
}

// ─── Bookings ─────────────────────────────────────────────────────────────────

func (m *MongoDB) CreateBooking(ctx context.Context, booking *models.Booking) error {
	booking.ID = primitive.NewObjectID()
	booking.CreatedAt = time.Now()
	booking.UpdatedAt = time.Now()
	_, err := m.db.Collection("bookings").InsertOne(ctx, booking)
	return err
}

func (m *MongoDB) FindBookingByID(ctx context.Context, id primitive.ObjectID) (*models.Booking, error) {
	var b models.Booking
	err := m.db.Collection("bookings").FindOne(ctx, bson.M{"_id": id}).Decode(&b)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &b, err
}

func (m *MongoDB) UpdateBookingStatus(ctx context.Context, id primitive.ObjectID, status models.BookingStatus, extra bson.M) error {
	update := bson.M{"status": status, "updated_at": time.Now()}
	for k, v := range extra {
		update[k] = v
	}
	_, err := m.db.Collection("bookings").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (m *MongoDB) FindBookingsForAdmin(ctx context.Context, status, showtimeID string, limit, skip int64) ([]models.Booking, int64, error) {
	col := m.db.Collection("bookings")
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}
	if showtimeID != "" {
		if oid, err := primitive.ObjectIDFromHex(showtimeID); err == nil {
			filter["showtime_id"] = oid
		}
	}
	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit).SetSkip(skip)
	}
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var result []models.Booking
	return result, total, cursor.All(ctx, &result)
}

func (m *MongoDB) FindUserBookings(ctx context.Context, userID primitive.ObjectID) ([]models.Booking, error) {
	cursor, err := m.db.Collection("bookings").Find(ctx,
		bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var result []models.Booking
	return result, cursor.All(ctx, &result)
}

// ─── Audit Logs ───────────────────────────────────────────────────────────────

func (m *MongoDB) InsertAuditLog(ctx context.Context, log *models.AuditLog) error {
	log.ID = primitive.NewObjectID()
	log.CreatedAt = time.Now()
	_, err := m.db.Collection("audit_logs").InsertOne(ctx, log)
	return err
}

func (m *MongoDB) FindAuditLogs(ctx context.Context, eventType, bookingID string, limit, skip int64) ([]models.AuditLog, int64, error) {
	col := m.db.Collection("audit_logs")
	filter := bson.M{}
	if eventType != "" {
		filter["event_type"] = eventType
	}
	if bookingID != "" {
		filter["booking_id"] = bookingID
	}
	total, _ := col.CountDocuments(ctx, filter)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit).SetSkip(skip)
	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	var result []models.AuditLog
	return result, total, cursor.All(ctx, &result)
}

// ─── Seed ─────────────────────────────────────────────────────────────────────

func (m *MongoDB) SeedIfEmpty(ctx context.Context) error {
	col := m.db.Collection("showtimes")
	count, _ := col.CountDocuments(ctx, bson.M{})
	if count > 0 {
		return nil
	}

	now := time.Now()
	type movieDef struct {
		name  string
		emoji string
		genre string
		dur   int
		rate  string
	}
	movies := []movieDef{
		{"Avengers: Secret Wars", "⚡", "Action", 180, "PG-13"},
		{"Dune: Messiah", "🏜️", "Sci-Fi", 165, "PG-13"},
		{"Spider-Man: Beyond the Spider-Verse", "🕷️", "Action", 140, "PG"},
		{"Inception 2", "🌀", "Sci-Fi", 155, "PG-13"},
		{"The Batman: Dark Knight Returns", "🦇", "Action", 150, "PG-13"},
		{"Interstellar: Return", "🚀", "Sci-Fi", 170, "PG"},
		{"A Quiet Place: Part IV", "🤫", "Horror", 100, "R"},
		{"Oppenheimer 2", "💥", "Drama", 180, "R"},
		{"Inside Out 3", "🎭", "Animation", 110, "G"},
		{"Moana 3", "🌊", "Animation", 115, "G"},
	}

	halls := []string{"Hall A", "Hall B", "Hall C", "Hall D", "Hall E", "IMAX"}
	var stIDs []primitive.ObjectID
	var stDocs []interface{}

	for i, mv := range movies {
		for round := 0; round < 2; round++ {
			startOffset := time.Duration(i*90+round*5*60) * time.Minute
			st := models.Showtime{
				ID:          primitive.NewObjectID(),
				MovieName:   mv.name,
				Hall:        halls[(i+round)%len(halls)],
				StartTime:   now.Add(startOffset),
				EndTime:     now.Add(startOffset + time.Duration(mv.dur)*time.Minute),
				PosterEmoji: mv.emoji,
				Genre:       mv.genre,
				Rating:      mv.rate,
				Duration:    mv.dur,
				CreatedAt:   now,
			}
			stDocs = append(stDocs, st)
			stIDs = append(stIDs, st.ID)
		}
	}

	if _, err := col.InsertMany(ctx, stDocs); err != nil {
		return err
	}

	rows := []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K"}
	seatsCol := m.db.Collection("seats")
	for _, stID := range stIDs {
		var seatDocs []interface{}
		for ri, row := range rows {
			seatType := models.SeatTypeNormal
			price := 300.0
			if ri < 2 {
				seatType = models.SeatTypeVIP
				price = 400.0
			}
			for num := 1; num <= 14; num++ {
				seatDocs = append(seatDocs, models.Seat{
					ID:         primitive.NewObjectID(),
					ShowtimeID: stID,
					Row:        row,
					Number:     num,
					SeatCode:   fmt.Sprintf("%s%d", row, num),
					Status:     models.SeatAvailable,
					Type:       seatType,
					Price:      price,
				})
			}
		}
		if _, err := seatsCol.InsertMany(ctx, seatDocs); err != nil {
			return err
		}
	}
	return nil
}
