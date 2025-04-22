package main

import (
	"net/http"
	"rwa/internal/app"
)

// сюда писать код

func GetApp() http.Handler {
	return app.Init()
}
