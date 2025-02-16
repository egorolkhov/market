package app

import (
	"avito/internal/app/services/auth"
	"avito/internal/app/services/buy"
	"avito/internal/app/services/info"
	"avito/internal/app/services/sendCoin"
	"avito/internal/cache"
	"avito/internal/models"
	"context"
)

type Storage interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error

	GetInfo(ctx context.Context, uuid string) (*models.Info, error)

	Send(ctx context.Context, uuid string, toUser string, amount int) error
	GetUuidByUsername(ctx context.Context, username string) (string, error)

	Buy(ctx context.Context, uuid string, item string) error
}

type App struct {
	auth.AuthController
	info.InfoController
	sendCoin.SendController
	buy.BuyController
	Lfu *cache.LFUCache
}

func NewApp(storage Storage, cache *cache.LFUCache) *App {
	return &App{
		AuthController: auth.AuthController{Storage: storage},
		InfoController: info.InfoController{Storage: storage, Lfu: cache},
		SendController: sendCoin.SendController{Storage: storage, Lfu: cache},
		BuyController:  buy.BuyController{Storage: storage, Lfu: cache},
		Lfu:            cache,
	}
}
