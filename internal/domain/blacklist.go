package domain

import (
	"context"
	"time"
)

type BlacklistType int

const (
	Track BlacklistType = iota
	Artist
	Genre
)

type BlacklistRepository interface {
	AddToBlacklist(ctx context.Context, blacklistType BlacklistType, id string,
		userFields *UserFields, now time.Time) error
	RemoveFromBlacklist(ctx context.Context, blacklistType BlacklistType, id string) error
	CheckBlacklistItem(ctx context.Context, blacklistType BlacklistType, id string) (bool, error)
}

type BlacklistService struct {
	Repo BlacklistRepository
}

func NewBlacklistService(repo BlacklistRepository) *BlacklistService {
	return &BlacklistService{
		Repo: repo,
	}
}
