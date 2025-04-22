package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"rwa/internal/model"
	"rwa/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handlers struct {
	UserRepository    *repository.PostgresUserStorage
	V                 *validator.Validate
	ArticleRepository *repository.PostgresArticleStorage
	log               *slog.Logger
}

func NewHandlers(db *pgxpool.Pool, log *slog.Logger) *Handlers {
	return &Handlers{
		UserRepository:    repository.NewPostgresUserStorage(db, log),
		V:                 validator.New(),
		ArticleRepository: repository.NewPostgresArticleStorage(db, log),
		log:               log,
	}
}

func HandleError(w http.ResponseWriter, errMsg string, statusCode int) {
	errJson := model.NewError(errMsg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errJson)
}
