package domain

import (
	"context"
	"encoding/json"
)

type PreferenceKey string

const (
	MaxDurationKey    PreferenceKey = "maxDuration"
	PurgeFrequencyKey PreferenceKey = "purgeFrequency"
	MaxTrackAgeKey    PreferenceKey = "maxTrackAge"
)

var AllPreferences = [3]PreferenceKey{
	MaxDurationKey,
	PurgeFrequencyKey,
	MaxTrackAgeKey,
}

func (key PreferenceKey) DefaultValue() any {
	switch key {
	case MaxDurationKey:
		// 10 minutes in MS
		return 10 * 60 * 1000
	case PurgeFrequencyKey:
		// 12 hours in MS
		return 12 * 60 * 60 * 1000
	case MaxTrackAgeKey:
		// 2 weeks in MS
		return 2 * 7 * 24 * 60 * 60 * 1000
	default:
		return nil
	}
}

type Preference struct {
	Key   PreferenceKey
	Value json.RawMessage
}

func (p Preference) IntValue() (int, error) {
	var jsonNum json.Number
	err := json.Unmarshal(p.Value, &jsonNum)
	if err != nil {
		return 0, err
	}

	num, err := jsonNum.Int64()
	if err != nil {
		return 0, err
	}

	return int(num), nil
}

type PreferenceRepository interface {
	Set(ctx context.Context, preferenceKey PreferenceKey, value any) error
	Get(ctx context.Context, preferenceKey PreferenceKey) (*Preference, error)
	GetAll(ctx context.Context) (map[PreferenceKey]Preference, error)
}

type PreferenceService struct {
	Repo PreferenceRepository
}

func NewPreferenceService(repo PreferenceRepository) *PreferenceService {
	return &PreferenceService{
		Repo: repo,
	}
}
