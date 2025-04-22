package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	handler "rwa/internal/handlers"
	"rwa/internal/security"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

var symmetrical_token string

func Init() http.Handler {
	ctx := context.Background()
	getEnv_db := os.Getenv("DB_URL")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	symmetrical_token = os.Getenv("JWT_SECRET")
	if symmetrical_token == "" {
		logger.Error("JWT_SECRET is not set")
		panic("JWT_SECRET is not set")
	}
	err := security.Init(symmetrical_token)
	if getEnv_db == "" {
		logger.Error("DB_URL is not set")
		panic("DB_URL is not set")
	}
	pool, err := pgxpool.New(ctx, getEnv_db)

	if err != nil {
		panic(err)
	}

	handlers := handler.NewHandlers(pool, logger)
	r := mux.NewRouter()
	r.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Hello, world!"))
	})
	r.HandleFunc("/users/login", handlers.LoginUserHandler).Methods(http.MethodPost)
	r.Handle("/users/logout", handlers.AuthMiddleware(http.HandlerFunc(handlers.LogoutHandler))).Methods(http.MethodPost)
	r.HandleFunc("/users", handlers.UserRegisterHandler).Methods(http.MethodPost)
	r.Handle("/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetUserHandler))).Methods(http.MethodGet)
	r.Handle("/users", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateUserHandler))).Methods(http.MethodPut)
	r.Handle("/profiles/{username}/follow", handlers.AuthMiddleware(http.HandlerFunc(handlers.FollowHandler))).Methods(http.MethodPost)
	r.Handle("/profiles/{username}/unfollow", handlers.AuthMiddleware(http.HandlerFunc(handlers.UnFollowHandler))).Methods(http.MethodDelete)
	r.Handle("/profiles/{username}", handlers.AuthMiddleware(http.HandlerFunc(handlers.CheckProfileHandler))).Methods(http.MethodGet)
	r.HandleFunc("/articles", handlers.GetArticleHandler).Methods(http.MethodGet)
	r.Handle("/articles", handlers.AuthMiddleware(http.HandlerFunc(handlers.CreateArticleHandler))).Methods(http.MethodPost)
	return r
}
