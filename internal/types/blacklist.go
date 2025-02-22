package types

type BlacklistType int

const (
	Track BlacklistType = iota
	Artist
	Genre
)
