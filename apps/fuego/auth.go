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

	"github.com/go-fuego/fuego"
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

type loginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenOutput struct {
	Token string `json:"token"`
}

func (sh *shop) login(c fuego.ContextWithBody[loginInput]) (tokenOutput, error) {
	in, err := c.Body()
	if err != nil {
		return tokenOutput{}, err
	}
	hash, err := sh.q.GetStaffHash(c.Context(), in.Username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		return tokenOutput{}, failure(http.StatusUnauthorized, "bad credentials")
	}
	return tokenOutput{Token: sh.tokens.issue(time.Now())}, nil
}

func (sh *shop) requireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || !sh.tokens.valid(token, time.Now()) {
			sendError(w, r, failure(http.StatusUnauthorized, "a valid bearer token is required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
