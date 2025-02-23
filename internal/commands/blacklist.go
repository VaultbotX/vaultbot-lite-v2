package commands

import (
	"context"
	mongocommands "github.com/vaultbotx/vaultbot-lite/internal/persistence/mongo/commands"
	internaltypes "github.com/vaultbotx/vaultbot-lite/internal/types"
	"time"
)

func Blacklist(ctx context.Context, blacklistType internaltypes.BlacklistType, id string,
	userFields *internaltypes.UserFields) error {
	now := time.Now()
	return mongocommands.Blacklist(ctx, blacklistType, id, userFields, now)
}

func Unblacklist(ctx context.Context, blacklistType internaltypes.BlacklistType, id string,
	userFields *internaltypes.UserFields) error {
	return mongocommands.Unblacklist(ctx, blacklistType, id, userFields)
}
