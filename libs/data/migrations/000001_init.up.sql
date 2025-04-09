CREATE EXTENSION citext;

CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    avatar text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS accounts (
    id serial PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    user_id integer NOT NULL REFERENCES users ON DELETE CASCADE,
    type VARCHAR(255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    provider_account_id VARCHAR(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamptz NOT NULL,
    scope text NOT NULL
);

