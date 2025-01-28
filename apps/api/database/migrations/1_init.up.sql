CREATE TABLE IF NOT EXISTS verification_tokens (
  identifier TEXT NOT NULL,
  expires TIMESTAMPTZ NOT NULL,
  token TEXT NOT NULL,
 
  PRIMARY KEY (identifier, token)
);
 
CREATE TABLE IF NOT EXISTS accounts (
  id SERIAL,
  user_id INTEGER NOT NULL,
  type VARCHAR(255) NOT NULL,
  provider VARCHAR(255) NOT NULL,
  provider_account_id VARCHAR(255) NOT NULL,
  refresh_token TEXT,
  access_token TEXT,
  expires TIMESTAMPTZ,
  id_token TEXT,
  scope TEXT,
  session_state TEXT,
  token_type TEXT,
 
  PRIMARY KEY (id)
);
 
CREATE TABLE IF NOT EXISTS sessions (
  id VARCHAR(64),
  user_id INTEGER,
  expires BIGINT NOT NULL,
  data BYTEA NOT NULL,
 
  PRIMARY KEY (id)
);
 
CREATE TABLE IF NOT EXISTS users (
  id SERIAL,
  name VARCHAR(255),
  email VARCHAR(255),
  email_verified TIMESTAMPTZ,
  image TEXT,
 
  PRIMARY KEY (id)
);


CREATE TABLE IF NOT EXISTS cats (
  id      SERIAL PRIMARY KEY,
  name    VARCHAR(40) NOT NULL
);

