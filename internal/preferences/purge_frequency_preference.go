package preferences

import (
	"context"
	"time"
)

func GetPurgeFrequencyPreference() (*Preference, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	pref, err := GetPreference(ctx, PurgeFrequencyKey)
	cancel()
	if err != nil {
		return nil, err
	}

	return pref, nil
}

func SetPurgeFrequencyPreference(durationInMilliseconds int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	err := SetPreference(ctx, PurgeFrequencyKey, durationInMilliseconds)
	cancel()
	if err != nil {
		return err
	}

	return nil
}
