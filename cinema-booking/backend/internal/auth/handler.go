package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cinema-booking/config"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Handler struct {
	cfg         *config.Config
	db          *repository.MongoDB
	oauthConfig *oauth2.Config
	adminEmails map[string]bool
}

type Claims struct {
	UserID string      `json:"user_id"`
	Email  string      `json:"email"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

func NewHandler(cfg *config.Config, db *repository.MongoDB) *Handler {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	adminEmails := make(map[string]bool)
	for _, email := range strings.Split(cfg.AdminEmails, ",") {
		email = strings.TrimSpace(email)
		if email != "" {
			adminEmails[email] = true
		}
	}

	return &Handler{
		cfg:         cfg,
		db:          db,
		oauthConfig: oauthConfig,
		adminEmails: adminEmails,
	}
}

// GoogleLogin redirects user to Google OAuth consent screen
func (h *Handler) GoogleLogin(c *gin.Context) {
	state := fmt.Sprintf("%d", time.Now().UnixNano())
	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the OAuth callback
func (h *Handler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing code"})
		return
	}

	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oauth exchange failed"})
		return
	}

	userInfo, err := h.fetchGoogleUserInfo(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user info"})
		return
	}

	role := models.RoleUser
	if h.adminEmails[userInfo.Email] {
		role = models.RoleAdmin
	}

	user := &models.User{
		GoogleID:  userInfo.ID,
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		Avatar:    userInfo.Picture,
		Role:      role,
		CreatedAt: time.Now(),
	}

	if err := h.db.UpsertUser(context.Background(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	// Fetch saved user to get ObjectID
	savedUser, err := h.db.FindUserByGoogleID(context.Background(), userInfo.ID)
	if err != nil || savedUser == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found after save"})
		return
	}

	jwtToken, err := h.generateJWT(savedUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Redirect to frontend with token
	redirectURL := fmt.Sprintf("%s/auth/callback?token=%s", h.cfg.FrontendURL, jwtToken)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// Me returns current user info
func (h *Handler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	email := c.GetString("email")
	role := c.GetString("role")
	name := c.GetString("name")

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"name":    name,
	})
}

type googleUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (h *Handler) fetchGoogleUserInfo(accessToken string) (*googleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (h *Handler) generateJWT(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

func (h *Handler) ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
