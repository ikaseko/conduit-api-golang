package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"rwa/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArticleStorage interface {
	CreateArticle(dbArticle model.DBArticle) error
	DeleteArticle(slug string) error
	GetArticles() ([]model.DBArticle, error)
	GetArticlesByAuthor(authorUsername string) ([]model.DBArticle, error)
	GetArticlesByTag(tag string) ([]model.DBArticle, error)
	GetArticlesByAuthorAndTag(authorUsername string, tag string) ([]model.DBArticle, error)
	GetArticleBySlug(slug string) (model.DBArticle, error)
	UpdateArticle(article model.DBArticle) error
	IsSlugExists(slug string) (bool, error)
}

type PostgresArticleStorage struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

const (
	opIsSlugExists              = "repository.PostgresArticleStorage.IsSlugExists"
	opCreateArticle             = "repository.PostgresArticleStorage.CreateArticle"
	opDeleteArticle             = "repository.PostgresArticleStorage.DeleteArticle"
	opGetArticles               = "repository.PostgresArticleStorage.GetArticles"
	opGetArticlesByAuthor       = "repository.PostgresArticleStorage.GetArticlesByAuthor"
	opGetArticlesByTag          = "repository.PostgresArticleStorage.GetArticlesByTag"
	opGetArticlesByAuthorAndTag = "repository.PostgresArticleStorage.GetArticlesByAuthorAndTag"
	opGetArticleBySlug          = "repository.PostgresArticleStorage.GetArticleBySlug"
	opUpdateArticle             = "repository.PostgresArticleStorage.UpdateArticle"
)

func (p PostgresArticleStorage) IsSlugExists(slug string) bool {
	const op = opIsSlugExists
	query := `SELECT slug FROM article WHERE slug = $1`
	ctx := context.Background()
	row := p.db.QueryRow(ctx, query, slug)
	var dbSlug string
	err := row.Scan(&dbSlug)
	if err != nil {
		p.log.Error("failed to check if slug exists", "op", op, "slug", slug, "error", err)

		return !(errors.Is(err, pgx.ErrNoRows))
	}
	return true
}

func NewPostgresArticleStorage(db *pgxpool.Pool, log *slog.Logger) *PostgresArticleStorage {
	return &PostgresArticleStorage{db: db, log: log}
}

func (p PostgresArticleStorage) CreateArticle(dbArticle model.DBArticle) error {
	const op = opCreateArticle
	query := `INSERT INTO article (slug, title, description, body, taglist, created_at, updated_at, author) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	ctx := context.Background()
	_, err := p.db.Exec(ctx, query, dbArticle.Slug, dbArticle.Title, dbArticle.Description, dbArticle.Body, dbArticle.TagList, dbArticle.CreatedAt, dbArticle.UpdatedAt, dbArticle.Author)
	if err != nil {
		p.log.Error("failed to create article", "op", op, "slug", dbArticle.Slug, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (p PostgresArticleStorage) DeleteArticle(slug string) error {
	const op = opDeleteArticle
	query := `DELETE FROM article WHERE slug = $1`
	ctx := context.Background()
	_, err := p.db.Exec(ctx, query, slug)
	if err != nil {
		p.log.Error("failed to delete article", "op", op, "slug", slug, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (p PostgresArticleStorage) GetArticles() ([]model.DBArticleResponseWithAuthorUsername, error) {
	const op = opGetArticles
	query := `SELECT * FROM article`
	ctx := context.Background()
	rows, err := p.db.Query(ctx, query)
	if err != nil {
		p.log.Error("failed to get articles", "op", op, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var articles []model.DBArticleResponseWithAuthorUsername
	for rows.Next() {
		var article model.DBArticleResponseWithAuthorUsername
		err := rows.Scan(&article.Slug, &article.Title, &article.Description, &article.Body, &article.TagList, &article.CreatedAt, &article.UpdatedAt, &article.FavoritesCount, &article.Author.Username)
		if err != nil {
			p.log.Error("failed to scan article", "op", op, "error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (p PostgresArticleStorage) GetArticlesByAuthor(authorUsername string) ([]model.DBArticleResponseWithAuthorUsername, error) {
	const op = opGetArticlesByAuthor
	query := `SELECT * FROM article WHERE author = $1`
	ctx := context.Background()
	rows, err := p.db.Query(ctx, query, authorUsername)
	if err != nil {
		p.log.Error("failed to get articles by author", "op", op, "author", authorUsername, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var articles []model.DBArticleResponseWithAuthorUsername
	for rows.Next() {
		var article model.DBArticleResponseWithAuthorUsername
		err = rows.Scan(&article.Slug, &article.Title, &article.Description, &article.Body, &article.TagList, &article.CreatedAt, &article.UpdatedAt, &article.FavoritesCount, &article.Author.Username)
		if err != nil {
			p.log.Error("failed to scan article", "op", op, "author", authorUsername, "error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (p PostgresArticleStorage) GetArticlesByTag(tag string) ([]model.DBArticleResponseWithAuthorUsername, error) {
	const op = opGetArticlesByTag
	query := `SELECT * FROM article WHERE taglist @> ARRAY[$1]`
	ctx := context.Background()
	rows, err := p.db.Query(ctx, query, tag)
	if err != nil {
		p.log.Error("failed to get articles by tag", "op", op, "tag", tag, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var articles []model.DBArticleResponseWithAuthorUsername
	for rows.Next() {
		var article model.DBArticleResponseWithAuthorUsername
		err = rows.Scan(&article.Slug, &article.Title, &article.Description, &article.Body, &article.TagList, &article.CreatedAt, &article.UpdatedAt, &article.FavoritesCount, &article.Author.Username)
		if err != nil {
			p.log.Error("failed to scan article", "op", op, "tag", tag, "error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (p PostgresArticleStorage) GetArticlesByAuthorAndTag(authorUsername string, tag string) ([]model.DBArticleResponseWithAuthorUsername, error) {
	const op = opGetArticlesByAuthorAndTag
	query := `SELECT * FROM article WHERE author = $1 AND taglist @> ARRAY[$2]`
	ctx := context.Background()
	rows, err := p.db.Query(ctx, query, authorUsername, tag)
	if err != nil {
		p.log.Error("failed to get articles by author and tag", "op", op, "author", authorUsername, "tag", tag, "error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	var articles []model.DBArticleResponseWithAuthorUsername
	for rows.Next() {
		var article model.DBArticleResponseWithAuthorUsername
		err = rows.Scan(&article.Slug, &article.Title, &article.Description, &article.Body, &article.TagList, &article.CreatedAt, &article.UpdatedAt, &article.FavoritesCount, &article.Author.Username)
		if err != nil {
			p.log.Error("failed to scan article", "op", op, "author", authorUsername, "tag", tag, "error", err)
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (p PostgresArticleStorage) GetArticleBySlug(slug string) (model.DBArticle, error) {
	const op = opGetArticleBySlug
	query := `SELECT * FROM article WHERE slug = $1`
	ctx := context.Background()
	row := p.db.QueryRow(ctx, query, slug)
	var article model.DBArticle
	err := row.Scan(&article.Slug, &article.Title, &article.Description, &article.Body, &article.TagList, &article.CreatedAt, &article.UpdatedAt, &article.FavoritesCount, &article.Author)
	if err != nil {
		p.log.Error("failed to get article by slug", "op", op, "slug", slug, "error", err)
		return model.DBArticle{}, fmt.Errorf("%s: %w", op, err)
	}
	return article, nil
}

func (p PostgresArticleStorage) UpdateArticle(article model.DBArticle) error {
	const op = opUpdateArticle
	query := `UPDATE article SET title = $1, description = $2, body = $3, taglist = $4, updated_at = $5 WHERE slug = $6`
	ctx := context.Background()
	_, err := p.db.Exec(ctx, query, article.Title, article.Description, article.Body, article.TagList, article.UpdatedAt, article.Slug)
	if err != nil {
		p.log.Error("failed to update article", "op", op, "slug", article.Slug, "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
