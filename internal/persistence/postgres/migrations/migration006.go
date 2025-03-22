package migrations

var Migration006 = &Migration{
	Name: "006-AdditionalSongFields",
	Up: `
		ALTER TABLE songs
		ADD COLUMN duration INT,
		ADD COLUMN popularity INT,
		ADD COLUMN album_name VARCHAR(255);
	`,
	Down: `
		ALTER TABLE songs
		DROP COLUMN duration,
		DROP COLUMN popularity,
		DROP COLUMN album_name;
	`,
}
