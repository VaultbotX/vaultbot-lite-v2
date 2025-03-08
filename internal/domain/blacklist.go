package domain

import (
	"context"
)

type BlacklistRepository interface {
	AddToBlacklist(ctx context.Context, blacklistType EntityType, id string,
		userFields *UserFields) error
	RemoveFromBlacklist(ctx context.Context, blacklistType EntityType, id string) error
	CheckBlacklistItem(ctx context.Context, blacklistType EntityType, id string) (bool, error)
}

type BlacklistService struct {
	Repo BlacklistRepository
}

func NewBlacklistService(repo BlacklistRepository) *BlacklistService {
	return &BlacklistService{
		Repo: repo,
	}
}
