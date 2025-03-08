package helpers

var (
	// AdminPermission is the int64 representation of an admin permission.
	// Not aware of any constants in Discordgo to represent the permission int64s
	// https://discord-api-types.dev/api/discord-api-types-payloads/common#PermissionFlagsBits
	// https://github.com/discordjs/discord-api-types/blob/0e6b19d2bcfe6e9806d3d20125668b3464845517/payloads/common.ts#L26
	AdminPermission int64 = 8

	MinSongDuration float64 = 0
	MaxSongDuration float64 = 120

	MinPurgeFrequency float64 = 1
	MaxPurgeFrequency float64 = 7

	MinTrackAge float64 = 1
	MaxTrackAge float64 = 31
)
