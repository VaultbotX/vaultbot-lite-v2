package migrations

var Migration003 = &Migration{
	Name: "003-Blacklist",
	Up: `
		CREATE TABLE IF NOT EXISTS blacklist (
		    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		    type VARCHAR(64) NOT NULL,
		    value VARCHAR(64) NOT NULL,
		    blocked_by_user_id INTEGER NOT NULL
		        REFERENCES users(id) ON DELETE SET NULL,
		    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_blacklist_value ON blacklist(value);
	`,
	Down: `DROP TABLE IF EXISTS blacklist;`,
}
