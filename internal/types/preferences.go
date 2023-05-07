package types

type PreferenceKey string

const (
	MaxDurationKey PreferenceKey = "maxDuration"
)

type Preference struct {
	Id    string        `bson:"_id"`
	Key   PreferenceKey `bson:"key"`
	Value interface{}   `bson:"value"`
}

var AllPreferences = [1]PreferenceKey{MaxDurationKey}

func (key PreferenceKey) DefaultValue() interface{} {
	switch key {
	case MaxDurationKey:
		// 10 minutes in MS
		return 10 * 60 * 1000
	default:
		return nil
	}
}
