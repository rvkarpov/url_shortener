package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/rvkarpov/url_shortener/internal/storage"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func Authorize(h http.HandlerFunc, logger *zap.SugaredLogger, secretKey string) http.HandlerFunc {
	return func(rsp http.ResponseWriter, rqs *http.Request) {
		cookie, err := rqs.Cookie("session")
		if err != nil {
			userID := uuid.New().String()
			setCookie(userID, secretKey, rsp, logger)
			ctx := context.WithValue(rqs.Context(), storage.UserIDKey{Name: "userID"}, userID)
			h.ServeHTTP(rsp, rqs.WithContext(ctx))
			return
		}

		claims, err := parseJWT(cookie.Value, secretKey)
		if err != nil {
			http.Error(rsp, err.Error(), http.StatusUnauthorized)
			return
		}

		logger.Infof("Authenticated user: %s", claims.UserID)
		ctx := context.WithValue(rqs.Context(), storage.UserIDKey{Name: "userID"}, claims.UserID)
		h.ServeHTTP(rsp, rqs.WithContext(ctx))
	}
}

func setCookie(userID string, secretKey string, rsp http.ResponseWriter, logger *zap.SugaredLogger) {
	tokenString, err := createJWT(userID, secretKey)
	if err != nil {
		http.Error(rsp, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(rsp, &http.Cookie{
		Name:   "session",
		Value:  tokenString,
		MaxAge: 3600,
	})

	logger.Infof("New user ID created: %s", userID)
}

func createJWT(userID string, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		UserID:           userID,
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func parseJWT(tokenString string, secretKey string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while parse token string: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}
