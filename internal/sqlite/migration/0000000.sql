CREATE TABLE users (
	id           NUMERIC PRIMARY KEY,
	username     TEXT NOT NULL UNIQUE,
	seed 				 TEXT NOT NULL,
	created_at   TEXT NOT NULL,
	updated_at   TEXT NOT NULL
);

CREATE INDEX users_username_idx ON users (username);

CREATE TABLE sessions (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id       INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	refresh_token TEXT NOT NULL,
	expires_at    TEXT NOT NULL,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);

CREATE TABLE bookmarks (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id     INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
	name        TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT "",
	url					TEXT NOT NULL,
	created_at  TEXT NOT NULL,
	updated_at  TEXT NOT NULL
);

CREATE INDEX bookmarks_user_id_idx ON bookmarks (user_id);