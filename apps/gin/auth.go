package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const tokenLifetime = 12 * time.Hour

// tokens issues and checks opaque bearer tokens: an expiry time and an HMAC
// of it under a per-process secret.
type tokens struct{ secret []byte }

func newTokens() *tokens {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic(err) // the system has no randomness; nothing sensible can run
	}
	return &tokens{secret: secret}
}

func (t *tokens) sign(expiry string) string {
	mac := hmac.New(sha256.New, t.secret)
	mac.Write([]byte(expiry))
	return hex.EncodeToString(mac.Sum(nil))
}

func (t *tokens) issue(now time.Time) string {
	expiry := strconv.FormatInt(now.Add(tokenLifetime).Unix(), 10)
	return expiry + "." + t.sign(expiry)
}

func (t *tokens) valid(token string, now time.Time) bool {
	expiry, sig, ok := strings.Cut(token, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(t.sign(expiry))) {
		return false
	}
	until, err := strconv.ParseInt(expiry, 10, 64)
	return err == nil && now.Unix() < until
}

func (s *server) login(c *gin.Context) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		writeError(c, http.StatusBadRequest, "body is not valid JSON")
		return
	}
	hash, err := s.q.GetStaffHash(c.Request.Context(), in.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		writeError(c, http.StatusUnauthorized, "bad credentials")
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": s.tokens.issue(time.Now())})
}

func (s *server) requireToken(c *gin.Context) {
	token, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
	if !ok || !s.tokens.valid(token, time.Now()) {
		writeError(c, http.StatusUnauthorized, "a valid bearer token is required")
		c.Abort()
		return
	}
	c.Next()
}
