CREATE TABLE users (
	id TEXT PRIMARY KEY,
	username TEXT NOT NULL UNIQUE,
	pwd_hash TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_oidc (
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    subject TEXT NOT NULL,
    PRIMARY KEY (provider, subject)
);

CREATE TABLE appointments (
	date_start	  INTEGER NOT NULL,
	date_end	  INTEGER NOT NULL,
	business_id   TEXT NOT NULL,
	client_id	  TEXT NOT NULL,
	UNIQUE (business_id, date_start)
);

CREATE TABLE business_work_rule (
	id TEXT NOT NULL,
	business_id TEXT NOT NULL,
	rule	  TEXT NOT NULL,
	UNIQUE (id, business_id)
);

CREATE TABLE user_bots (
    bot_id          TEXT PRIMARY KEY,
    bot_token_hash  TEXT NOT NULL,
    business_id     TEXT NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT 1,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_used_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE user_tokens (
	token      	TEXT PRIMARY KEY,
	business_id TEXT NOT NULL,
	client_id  	TEXT NOT NULL,
	expires_at  INTEGER NOT NULL,
	is_used     BOOLEAN NOT NULL DEFAULT FALSE,
	UNIQUE (business_id, client_id)
);