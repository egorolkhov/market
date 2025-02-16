package routes

import (
	"avito/internal/app"
	"avito/internal/logger"
	"avito/internal/middleware"
	"github.com/gorilla/mux"
)

func NewHandler(App app.App) *mux.Router {
	handler := mux.NewRouter()

	handler.HandleFunc("/api/info", middleware.Compress(middleware.Cookie(logger.GetLogger(App.Info)))).Methods("GET")
	handler.HandleFunc("/api/sendCoin", middleware.Compress(middleware.Cookie(logger.PostLogger(App.SendCoin)))).Methods("POST")
	handler.HandleFunc("/api/buy/{item}", middleware.Compress(middleware.Cookie(logger.GetLogger(App.Buy)))).Methods("GET")
	handler.HandleFunc("/api/auth", middleware.Compress(logger.PostLogger(App.Auth))).Methods("POST")

	return handler
}
