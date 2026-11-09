CREATE TABLE bond_groups (
    id          SERIAL PRIMARY KEY,
    profile_id  INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name        VARCHAR(64) NOT NULL,
    weight      INT NOT NULL DEFAULT 0,
    UNIQUE      (profile_id, name)
);

CREATE TABLE bonds (
    id          SERIAL PRIMARY KEY,
    group_id    INT NOT NULL REFERENCES bond_groups(id) ON DELETE CASCADE,
    name        VARCHAR(128) NOT NULL,
    code        VARCHAR(32) NOT NULL,
    lot_amount  INT NOT NULL,
    lot_size    INT NOT NULL,
    UNIQUE      (group_id, code)
);

ALTER TABLE transactions RENAME COLUMN stock_code TO instrument_code;
ALTER TABLE transactions ADD COLUMN instrument_type VARCHAR(16);
UPDATE transactions SET instrument_type = 'stock';
ALTER TABLE transactions ALTER COLUMN instrument_type SET NOT NULL;
ALTER TABLE transactions ADD CONSTRAINT transactions_instrument_type_check
    CHECK (instrument_type IN ('stock', 'bond'));
