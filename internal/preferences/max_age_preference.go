package preferences

import (
	"context"
	"time"
)

func SetMaxTrackAgePreference(durationInMilliseconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	err := SetPreference(ctx, MaxTrackAgeKey, durationInMilliseconds)
	cancel()
	if err != nil {
		return err
	}

	return nil
}

func GetMaxTrackAgePreference() (*Preference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pref, err := GetPreference(ctx, MaxTrackAgeKey)
	cancel()
	if err != nil {
		return nil, err
	}

	return pref, nil
}
