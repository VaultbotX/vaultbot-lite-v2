package migrations

var Migration004 = &Migration{
	Name: "004-EventSourcePreferences",
	Up: `
		ALTER TABLE preferences DROP CONSTRAINT preferences_pkey;
		ALTER TABLE preferences ADD COLUMN id SERIAL PRIMARY KEY;
		ALTER TABLE preferences ADD COLUMN start_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
		ALTER TABLE preferences ADD COLUMN end_time TIMESTAMP;

		-- update the existing preference records to have the start_time equal to 
		-- the created_date of the first song, if exists
		UPDATE preferences
		SET start_time = COALESCE((SELECT created_at FROM songs ORDER BY created_at LIMIT 1), CURRENT_TIMESTAMP);
	`,
	Down: `
 		ALTER TABLE preferences DROP COLUMN start_time;
		ALTER TABLE preferences DROP COLUMN end_time;
		ALTER TABLE preferences DROP CONSTRAINT preferences_pkey;
		ALTER TABLE preferences DROP COLUMN id;
		ALTER TABLE preferences ADD PRIMARY KEY (key);
	`,
}
