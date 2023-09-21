CREATE TABLE IF NOT EXISTS merkle_leaves (
    id SERIAL PRIMARY KEY,
    leaf_content BYTEA NOT NULL
);