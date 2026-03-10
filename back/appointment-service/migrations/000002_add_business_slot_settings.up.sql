CREATE TABLE business_slot_settings (
    business_id             TEXT PRIMARY KEY,
    default_chunk_minutes   INTEGER NOT NULL,
    max_chunk_minutes       INTEGER NOT NULL
);
