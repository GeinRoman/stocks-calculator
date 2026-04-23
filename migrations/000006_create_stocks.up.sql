CREATE TABLE stocks (
    id          SERIAL PRIMARY KEY,
    group_id    INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    name        VARCHAR(128) NOT NULL,
    code        VARCHAR(32) NOT NULL,
    lot_amount  INT NOT NULL,
    lot_size    INT NOT NULL,
    UNIQUE      (group_id, code)
);
