CREATE TABLE crypto_groups (
    id         SERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name       VARCHAR(64) NOT NULL,
    weight     INT NOT NULL DEFAULT 0 CHECK (weight BETWEEN 0 AND 100),
    UNIQUE (profile_id, name),
    UNIQUE (id, profile_id)
);

CREATE TABLE crypto_assets (
    id              BIGSERIAL PRIMARY KEY,
    profile_id      INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    group_id        INT NOT NULL,
    name            TEXT NOT NULL,
    platform        TEXT NOT NULL,
    url             TEXT NOT NULL DEFAULT '',
    annual_interest FLOAT8 NOT NULL CHECK (annual_interest >= 0),
    amount          NUMERIC NOT NULL CHECK (amount > 0),
    UNIQUE (profile_id, name),
    FOREIGN KEY (group_id, profile_id) REFERENCES crypto_groups(id, profile_id)
);

-- Asset IDs are historical identifiers: a full sale deletes the live asset first.
CREATE TABLE crypto_transactions (
    id         BIGSERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    asset_id   BIGINT NOT NULL,
    asset_name TEXT NOT NULL,
    platform   TEXT NOT NULL,
    amount     NUMERIC NOT NULL CHECK (amount > 0),
    buying     BOOLEAN NOT NULL,
    price      FLOAT8 NOT NULL CHECK (price > 0),
    currency   VARCHAR(16) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
