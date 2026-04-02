package migrations

var Migration007 = &Migration{
	Name: "007-TruncateBlacklist",
	Up:   `TRUNCATE TABLE blacklist;`,
	Down: ``,
}
