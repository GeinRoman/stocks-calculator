CREATE TABLE groups (
    id          SERIAL PRIMARY KEY,
    profile_id  INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name        VARCHAR(64) NOT NULL,
    weight      INT NOT NULL DEFAULT 0,
    UNIQUE      (profile_id, name)
);

