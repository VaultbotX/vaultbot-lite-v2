package preferences

import (
	"context"
	"time"
)

func SetMaxSongDurationPreference(durationInMilliseconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	err := SetPreference(ctx, MaxDurationKey, durationInMilliseconds)
	cancel()
	if err != nil {
		return err
	}

	return nil
}

func GetMaxSongDurationPreference() (*Preference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pref, err := GetPreference(ctx, MaxDurationKey)
	cancel()
	if err != nil {
		return nil, err
	}

	return pref, nil
}
