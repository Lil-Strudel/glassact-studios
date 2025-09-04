CREATE EXTENSION citext;
CREATE EXTENSION postgis;

CREATE TABLE IF NOT EXISTS dealerships (
    id serial PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name text NOT NULL,    
    street text NOT NULL,
    street_ext text NOT NULL,
    city text NOT NULL,
    state text NOT NULL,
    postal_code text NOT NULL,
    country text NOT NULL,
    location GEOGRAPHY(Point, 4326) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS users (
    id serial PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    avatar text NOT NULL,
    dealership_id integer NOT NULL REFERENCES dealerships ON DELETE RESTRICT,
    updated_at timestamptz NOT NULL DEFAULT now(),
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
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamptz NOT NULL,
    scope text NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
    id serial PRIMARY KEY,
    uuid UUID DEFAULT gen_random_uuid() UNIQUE NOT NULL,
    name text NOT NULL,
    status VARCHAR(255) NOT NULL CHECK (status IN ('awaiting-proof', 'proof-in-revision', 'all-proofs-accepted', 'cancelled', 'ordered', 'in-production', 'awaiting-payment', 'completed')),
    approved boolean NOT NULL,
    dealership_id integer NOT NULL REFERENCES dealerships ON DELETE RESTRICT,
    updated_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    version integer NOT NULL DEFAULT 1
);

