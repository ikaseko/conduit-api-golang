package model

import "time"

type DBArticle struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	FavoritesCount int       `json:"favoritesCount"`
	Author         string    `json:"author"`
}

type Author struct {
	Username string `json:"username"`
	Bio      string `json:"bio"`
}
type DBArticleResponse struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int       `json:"favoritesCount"`
	Author         Author    `json:"author"`
}

type CreateArticleRequest struct {
	Article struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Body        string   `json:"body"`
		TagList     []string `json:"tagList"`
	} `json:"article"`
}

type DBArticleResponseWithUsernameJson struct {
	Articles      []DBArticleResponseWithAuthorUsername `json:"articles"`
	ArticlesCount int                                   `json:"articlesCount"`
}

type DBArticleResponseWithUsernameJsonWithoutCount struct {
	Articles DBArticleResponseWithAuthorUsername `json:"article"`
}

type DBArticleResponseWithAuthorUsername struct {
	Slug           string         `json:"slug"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	Body           string         `json:"body"`
	TagList        []string       `json:"tagList"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	Favorited      bool           `json:"favorited"`
	FavoritesCount int            `json:"favoritesCount"`
	Author         AuthorUsername `json:"author"`
}

type AuthorUsername struct {
	Username string `json:"username"`
}
