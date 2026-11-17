DO $$
BEGIN
    LOCK TABLE transactions IN ACCESS EXCLUSIVE MODE;
    IF EXISTS (SELECT 1 FROM transactions WHERE instrument_type = 'bond') THEN
        RAISE EXCEPTION 'cannot roll back bond migration while bond transactions exist';
    END IF;
END $$;

ALTER TABLE transactions DROP CONSTRAINT transactions_instrument_type_check;
ALTER TABLE transactions DROP COLUMN instrument_type;
ALTER TABLE transactions RENAME COLUMN instrument_code TO stock_code;
DROP TABLE bonds;
DROP TABLE bond_groups;
