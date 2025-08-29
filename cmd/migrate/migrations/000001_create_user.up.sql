CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS "public".users(
    id bigserial PRIMARY KEY,
    email citext unique not null,
    username varchar(255) unique not null,
    password bytea not null,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);