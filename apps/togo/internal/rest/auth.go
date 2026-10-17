package rest

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/crypto/bcrypt"

	"github.com/angvp/tango-benchmarks/apps/togo/internal/app"
)

const (
	tokenLifetime = 12 * time.Hour
	staffUsername = "bench-staff"
)

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

// EnsureStaff creates the benchmark staff account, or resets its password.
func EnsureStaff(ctx context.Context, a *app.App, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = a.SQLDB.ExecContext(ctx, `INSERT INTO staff (username, password_hash) VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash`, staffUsername, string(hash))
	return err
}

type loginInput struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
}

type tokenOutput struct {
	Body struct {
		Token string `json:"token"`
	}
}

func (s *shop) login(ctx context.Context, in *loginInput) (*tokenOutput, error) {
	var hash string
	err := s.a.SQLDB.QueryRowContext(ctx, `SELECT password_hash FROM staff WHERE username = $1`, in.Body.Username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) || err == nil && bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Body.Password)) != nil {
		return nil, failure(http.StatusUnauthorized, "bad credentials")
	}
	if err != nil {
		return nil, err
	}
	out := &tokenOutput{}
	out.Body.Token = s.tokens.issue(time.Now())
	return out, nil
}

// requireToken is the middleware of every route that needs a staff token.
func (s *shop) requireToken(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		token, ok := strings.CutPrefix(ctx.Header("Authorization"), "Bearer ")
		if !ok || !s.tokens.valid(token, time.Now()) {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "a valid bearer token is required")
			return
		}
		next(ctx)
	}
}
