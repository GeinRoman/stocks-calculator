CREATE TABLE deposit_transactions (
    id         BIGSERIAL PRIMARY KEY,
    profile_id INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    bank_name  TEXT NOT NULL,
    amount_rub FLOAT8 NOT NULL CHECK (amount_rub > 0 AND amount_rub < 'Infinity'::FLOAT8),
    buying     BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, profile_id)
);

CREATE TABLE insurance_transactions (
    id           BIGSERIAL PRIMARY KEY,
    profile_id   INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    company_name TEXT NOT NULL,
    amount_rub   FLOAT8 NOT NULL CHECK (amount_rub > 0 AND amount_rub < 'Infinity'::FLOAT8),
    buying       BOOLEAN NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, profile_id)
);

CREATE TABLE deposits (
    id                         BIGSERIAL PRIMARY KEY,
    profile_id                 INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    acquisition_transaction_id BIGINT NOT NULL UNIQUE,
    name                       TEXT NOT NULL,
    bank_name                  TEXT NOT NULL,
    amount_rub                 FLOAT8 NOT NULL CHECK (amount_rub > 0 AND amount_rub < 'Infinity'::FLOAT8),
    annual_interest            FLOAT8 NOT NULL CHECK (annual_interest >= 0 AND annual_interest < 'Infinity'::FLOAT8),
    end_date                   DATE NOT NULL,
    FOREIGN KEY (acquisition_transaction_id, profile_id) REFERENCES deposit_transactions(id, profile_id)
);

CREATE TABLE insurances (
    id                         BIGSERIAL PRIMARY KEY,
    profile_id                 INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    acquisition_transaction_id BIGINT NOT NULL UNIQUE,
    name                       TEXT NOT NULL,
    company_name               TEXT NOT NULL,
    amount_rub                 FLOAT8 NOT NULL CHECK (amount_rub > 0 AND amount_rub < 'Infinity'::FLOAT8),
    annual_interest            FLOAT8 NOT NULL CHECK (annual_interest >= 0 AND annual_interest < 'Infinity'::FLOAT8),
    end_date                   DATE NOT NULL,
    FOREIGN KEY (acquisition_transaction_id, profile_id) REFERENCES insurance_transactions(id, profile_id)
);

CREATE TABLE credit_cards (
    id                           BIGSERIAL PRIMARY KEY,
    profile_id                   INT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    name                         TEXT NOT NULL,
    bank_name                    TEXT NOT NULL,
    mandatory_payment_date       DATE NOT NULL,
    mandatory_payment_amount_rub FLOAT8 NOT NULL CHECK (mandatory_payment_amount_rub >= 0 AND mandatory_payment_amount_rub < 'Infinity'::FLOAT8),
    overall_limit_rub            FLOAT8 NOT NULL CHECK (overall_limit_rub >= 0 AND overall_limit_rub < 'Infinity'::FLOAT8)
);

CREATE INDEX deposit_transactions_profile_id_idx ON deposit_transactions (profile_id);
CREATE INDEX insurance_transactions_profile_id_idx ON insurance_transactions (profile_id);
CREATE INDEX deposits_profile_id_idx ON deposits (profile_id);
CREATE INDEX insurances_profile_id_idx ON insurances (profile_id);
CREATE INDEX credit_cards_profile_id_idx ON credit_cards (profile_id);
