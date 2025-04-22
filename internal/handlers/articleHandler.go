package handler

import (
	"encoding/json"
	"net/http"
	"rwa/internal/model"
	"rwa/internal/pkg"
	"time"
)

func (h *Handlers) CreateArticleHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handler.CreateArticleHandler"
	h.log.With("op", op).Info("Attempting to create article")

	uid := r.Context().Value("uid").(string)
	user, err := h.UserRepository.GetUserForApi(uid)
	if err != nil {
		h.log.With("op", op, "uid", uid).Error("User not found", "error", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	request := model.CreateArticleRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.log.With("op", op).Error("Failed to decode request body", "error", err)
		http.Error(w, "Failed to decode request body", http.StatusUnprocessableEntity)
		return
	}
	index := 1
	slug := pkg.Slugify(request.Article.Title, index)
	var shouldFindSlug bool = true
	for shouldFindSlug {
		shouldFindSlug = h.ArticleRepository.IsSlugExists(slug)
		if shouldFindSlug {
			index++
			slug = pkg.Slugify(request.Article.Title, index)
		} else {
			shouldFindSlug = false
		}
	}
	article := model.DBArticle{
		Slug:           slug,
		Title:          request.Article.Title,
		Description:    request.Article.Description,
		Body:           request.Article.Body,
		TagList:        request.Article.TagList,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		FavoritesCount: 0,
		Author:         user.Username,
	}
	err = h.ArticleRepository.CreateArticle(article)
	if err != nil {
		h.log.With("op", op).Error("Failed to create article", "error", err)
		http.Error(w, "Failed to create article", http.StatusUnprocessableEntity)
		return
	}

	responseWithAuthorUsername := model.DBArticleResponseWithAuthorUsername{
		Slug:           slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Favorited:      false,
		FavoritesCount: 0,
		Author:         model.AuthorUsername{Username: user.Username},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	jsonWithoutCount := model.DBArticleResponseWithUsernameJsonWithoutCount{Articles: responseWithAuthorUsername}
	err = json.NewEncoder(w).Encode(&jsonWithoutCount)
	if err != nil {
		h.log.With("op", op).Error("Failed to encode response", "error", err)
		// Note: Cannot send http.Error here as headers are already written
		return
	}
	h.log.With("op", op, "slug", slug).Info("Article created successfully")
	return
}

func (h *Handlers) GetArticleHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handler.GetArticleHandler"

	valueAuthor := r.FormValue("author")
	valueTag := r.FormValue("tag")

	h.log.With("op", op).Info("Attempting to get articles", "author", valueAuthor, "tag", valueTag)

	var articles []model.DBArticleResponseWithAuthorUsername
	var err error
	if valueAuthor == "" && valueTag == "" {
		articles, err = h.ArticleRepository.GetArticles()
		if err != nil {
			h.log.With("op", op).Error("Failed to get articles", "error", err)
			// Removing fmt.Println(err) as we now log the error
			http.Error(w, "Failed to get articles", http.StatusInternalServerError)
			return
		}
	} else {
		if valueAuthor != "" && valueTag != "" {
			articles, err = h.ArticleRepository.GetArticlesByAuthorAndTag(valueAuthor, valueTag)
			if err != nil {
				h.log.With("op", op).Error("Failed to get articles by author and tag", "author", valueAuthor, "tag", valueTag, "error", err)
				http.Error(w, "Failed to get articles by author and tag", http.StatusInternalServerError)
				return
			}
		} else if valueAuthor != "" {
			articles, err = h.ArticleRepository.GetArticlesByAuthor(valueAuthor)
			if err != nil {
				h.log.With("op", op).Error("Failed to get articles by author", "author", valueAuthor, "error", err)
				http.Error(w, "Failed to get articles by author", http.StatusInternalServerError)
				return
			}
		} else if valueTag != "" {
			articles, err = h.ArticleRepository.GetArticlesByTag(valueTag)
			if err != nil {
				h.log.With("op", op).Error("Failed to get articles by tag", "tag", valueTag, "error", err)
				http.Error(w, "Failed to get articles by tag", http.StatusInternalServerError)
				return
			}
		}
	}
	responseJSON := model.DBArticleResponseWithUsernameJson{
		Articles:      articles,
		ArticlesCount: len(articles),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(&responseJSON)
	if err != nil {
		h.log.With("op", op).Error("Failed to encode response", "error", err)
		return
	}
	h.log.With("op", op, "count", len(articles)).Info("Articles retrieved successfully")
	return
}
