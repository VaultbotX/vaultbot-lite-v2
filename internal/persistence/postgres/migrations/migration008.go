package migrations

var Migration008 = &Migration{
	Name: "008-DropBlacklistAndPreferences",
	Up: `
		DROP TABLE IF EXISTS blacklist;
		DROP TABLE IF EXISTS preferences;
	`,
	Down: ``,
}
