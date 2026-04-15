CREATE TABLE transactions (
    id            SERIAL PRIMARY KEY,
    profile_id    INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    stock_code    VARCHAR(32) NOT NULL,
    lot_amount    INT NOT NULL,
    buying        BOOLEAN NOT NULL,
    lot_size      INT NOT NULL,
    price_in_rub  FLOAT8 NOT NULL,
    datetime      TIMESTAMPTZ NOT NULL
);
