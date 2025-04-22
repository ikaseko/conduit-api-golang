package handler

import (
	"context"
	"net/http"
	"rwa/internal/security"
	"time"
)

// userCtxKey is the context key for the user ID.

func (h *Handlers) AuthMiddleware(next http.Handler) http.Handler {
	const op = "handler.AuthMiddleware"

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authHeader := request.Header.Get("Authorization")
		if authHeader == "" {
			h.log.Warn("Missing Authorization header", "op", op)
			HandleError(writer, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[6:] // Get token part

		uid, err := security.DecodeToken(tokenString)
		if err != nil {
			h.log.Error("Failed to decode token", "op", op, "error", err)
			HandleError(writer, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		token, err := h.UserRepository.GetToken(tokenString)
		if err != nil {
			// Consider if this should be 500 or 401 depending on expected errors
			h.log.Error("Failed to retrieve token from repository", "op", op, "error", err)
			// Decide if token not found is a client error (401) or server error (500)
			HandleError(writer, "Failed to validate token", http.StatusUnauthorized) // Or Internal Server Error 500
			return
		}

		if token.UID != uid || !token.EndDate.After(time.Now()) {
			h.log.Warn("Token validation failed: UID mismatch or token expired", "op", op, "tokenUID", token.UID, "decodedUID", uid, "expiry", token.EndDate)
			HandleError(writer, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Use request's context as base
		ctx := context.WithValue(request.Context(), "uid", token.UID)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
