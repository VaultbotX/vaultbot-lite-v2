package migrations

var Migration009 = &Migration{
	Name: "009-DropUsers",
	Up: `
		ALTER TABLE song_archive DROP CONSTRAINT IF EXISTS song_archive_user_id_fkey;
		ALTER TABLE song_archive DROP COLUMN IF EXISTS user_id;
		DROP TABLE IF EXISTS users;
	`,
	Down: ``,
}
