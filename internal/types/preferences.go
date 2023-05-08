package types

type PreferenceKey string

const (
	MaxDurationKey    PreferenceKey = "maxDuration"
	PurgeFrequencyKey PreferenceKey = "purgeFrequency"
)

type Preference struct {
	Id    string        `bson:"_id"`
	Key   PreferenceKey `bson:"key"`
	Value interface{}   `bson:"value"`
}

var AllPreferences = [2]PreferenceKey{MaxDurationKey, PurgeFrequencyKey}

func (key PreferenceKey) DefaultValue() interface{} {
	switch key {
	case MaxDurationKey:
		// 10 minutes in MS
		return 10 * 60 * 1000
	case PurgeFrequencyKey:
		// 12 hours in MS
		return 12 * 60 * 60 * 1000
	default:
		return nil
	}
}
