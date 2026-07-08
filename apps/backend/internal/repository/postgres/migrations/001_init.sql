CREATE TABLE IF NOT EXISTS decks (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL
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
  tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL,
  next_review_at TIMESTAMPTZ NOT NULL,
  last_review_at TIMESTAMPTZ NULL,
  interval_days INTEGER NOT NULL DEFAULT 0,
  ease_factor DOUBLE PRECISION NOT NULL DEFAULT 2.5,
  review_count INTEGER NOT NULL DEFAULT 0,
  lapse_count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS cards_deck_id_idx ON cards(deck_id);
CREATE INDEX IF NOT EXISTS cards_next_review_at_idx ON cards(next_review_at);

CREATE TABLE IF NOT EXISTS reviews (
  id TEXT PRIMARY KEY,
  card_id TEXT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
  rating TEXT NOT NULL,
  reviewed_at TIMESTAMPTZ NOT NULL,
  previous_state JSONB NOT NULL,
  next_state JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS reviews_card_id_idx ON reviews(card_id);
CREATE INDEX IF NOT EXISTS reviews_reviewed_at_idx ON reviews(reviewed_at);
