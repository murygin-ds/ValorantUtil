CREATE TYPE valorant_region AS ENUM ('na', 'latam', 'br', 'eu', 'ap', 'kr');
CREATE TYPE valorant_shard AS ENUM ('na', 'pbe', 'eu', 'ap', 'kr');


CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    login       VARCHAR(32) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE valorant_accounts (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    puuid      TEXT NOT NULL UNIQUE,
    region     valorant_region NOT NULL,
    shard      valorant_shard NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_valorant_accounts_user_id ON valorant_accounts(user_id);