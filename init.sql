CREATE TABLE IF NOT EXISTS file_storage (
  file_name VARCHAR(255) PRIMARY KEY,
  file_content BYTEA NOT NULL UNIQUE,
  merkle_root BYTEA
);
