package sqlite

func (store *Store) Migrate() error {
	if _, err := store.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_admin INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS decks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS cards (
			id TEXT PRIMARY KEY,
			deck_id TEXT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			korean TEXT NOT NULL,
			translation TEXT NOT NULL,
			romanization TEXT NOT NULL DEFAULT '',
			example_korean TEXT NOT NULL DEFAULT '',
			example_translation TEXT NOT NULL DEFAULT '',
			tags TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			next_review_at TEXT NOT NULL,
			last_review_at TEXT,
			interval_days INTEGER NOT NULL DEFAULT 0,
			ease_factor REAL NOT NULL DEFAULT 2.5,
			review_count INTEGER NOT NULL DEFAULT 0,
			lapse_count INTEGER NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS reviews (
			id TEXT PRIMARY KEY,
			card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
			rating TEXT NOT NULL,
			reviewed_at TEXT NOT NULL,
			previous_state TEXT NOT NULL,
			next_state TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS lessons (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			level TEXT NOT NULL,
			sort_order INTEGER NOT NULL,
			content TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS lesson_progress (
			user_id TEXT NOT NULL,
			lesson_id TEXT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
			completed INTEGER NOT NULL DEFAULT 0,
			score INTEGER NOT NULL DEFAULT 0,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (user_id, lesson_id)
		);

		CREATE TABLE IF NOT EXISTS journal_entries (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			title TEXT NOT NULL,
			original_text TEXT NOT NULL,
			corrected_text TEXT NOT NULL,
			corrections TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_seed_versions (
			user_id TEXT PRIMARY KEY,
			version INTEGER NOT NULL,
			updated_at TEXT NOT NULL
		);
	`); err != nil {
		return err
	}

	for _, migration := range []struct {
		table      string
		column     string
		definition string
	}{
		{"decks", "user_id", "TEXT NOT NULL DEFAULT 'admin'"},
		{"cards", "user_id", "TEXT NOT NULL DEFAULT 'admin'"},
		{"reviews", "user_id", "TEXT NOT NULL DEFAULT 'admin'"},
	} {
		if err := store.ensureColumn(migration.table, migration.column, migration.definition); err != nil {
			return err
		}
	}

	if _, err := store.db.Exec(`
		CREATE INDEX IF NOT EXISTS decks_user_id_idx ON decks(user_id);
		CREATE INDEX IF NOT EXISTS cards_user_id_idx ON cards(user_id);
		CREATE INDEX IF NOT EXISTS cards_deck_id_idx ON cards(deck_id);
		CREATE INDEX IF NOT EXISTS cards_next_review_at_idx ON cards(next_review_at);
		CREATE INDEX IF NOT EXISTS reviews_user_id_idx ON reviews(user_id);
		CREATE INDEX IF NOT EXISTS reviews_card_id_idx ON reviews(card_id);
		CREATE INDEX IF NOT EXISTS reviews_reviewed_at_idx ON reviews(reviewed_at);
		CREATE INDEX IF NOT EXISTS journal_user_id_idx ON journal_entries(user_id);
	`); err != nil {
		return err
	}

	return store.seedLessons()
}
func (store *Store) ensureColumn(table string, column string, definition string) error {
	rows, err := store.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	_, err = store.db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}
