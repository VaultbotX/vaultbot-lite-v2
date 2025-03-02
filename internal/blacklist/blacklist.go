package blacklist

import (
	"context"
	internaltypes "github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

func Blacklist(ctx context.Context, blacklistType BlacklistType, id string,
	userFields *internaltypes.UserFields) error {
	now := time.Now()
	return AddToBlacklist(ctx, blacklistType, id, userFields, now)
}

func Unblacklist(ctx context.Context, blacklistType BlacklistType, id string,
	userFields *internaltypes.UserFields) error {
	return RemoveFromBlacklist(ctx, blacklistType, id, userFields)
}
