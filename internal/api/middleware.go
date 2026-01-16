package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			// For development, allow user_id query param
			if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err == nil && userID > 0 {
					ctx := context.WithValue(r.Context(), userIDKey, userID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			http.Error(w, "missing init data", http.StatusUnauthorized)
			return
		}

		userID, err := s.validateInitData(initData)
		if err != nil {
			http.Error(w, "invalid init data", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) validateInitData(initData string) (int64, error) {
	// Parse init data
	values, err := url.ParseQuery(initData)
	if err != nil {
		return 0, err
	}

	hash := values.Get("hash")
	if hash == "" {
		return 0, fmt.Errorf("missing hash")
	}

	// Build data check string
	var keys []string
	for k := range values {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var dataCheckString strings.Builder
	for i, k := range keys {
		if i > 0 {
			dataCheckString.WriteString("\n")
		}
		dataCheckString.WriteString(k + "=" + values.Get(k))
	}

	// Validate HMAC
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(s.botToken))

	h := hmac.New(sha256.New, secretKey.Sum(nil))
	h.Write([]byte(dataCheckString.String()))

	if hex.EncodeToString(h.Sum(nil)) != hash {
		return 0, fmt.Errorf("invalid hash")
	}

	// Extract user ID from user JSON
	userJSON := values.Get("user")
	if userJSON == "" {
		return 0, fmt.Errorf("missing user data")
	}

	var userData struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userJSON), &userData); err != nil {
		return 0, fmt.Errorf("parse user data: %w", err)
	}

	if userData.ID == 0 {
		return 0, fmt.Errorf("invalid user id")
	}

	return userData.ID, nil
}

func getUserID(ctx context.Context) int64 {
	if id, ok := ctx.Value(userIDKey).(int64); ok {
		return id
	}
	return 0
}
