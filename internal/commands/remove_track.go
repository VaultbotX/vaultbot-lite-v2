package commands

import "context"

func RemoveTracks(ctx context.Context, trackIds []string) error {
	// 1. Remove tracks from playlist
	// TODO

	// 2. Remove track from database(s) - practically speaking, this will just be a redis cache

	return nil
}
