package migrations

var Migration002 = &Migration{
	Name: "002-Preferences",
	Up: `
		CREATE TABLE IF NOT EXISTS preferences (
		    key VARCHAR(64) PRIMARY KEY,
		    value JSON NOT NULL
		);
	`,
	Down: `DROP TABLE IF EXISTS preferences;`,
}
