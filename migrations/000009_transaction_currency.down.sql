DO $$
BEGIN
    LOCK TABLE transactions IN ACCESS EXCLUSIVE MODE;
    IF EXISTS (SELECT 1 FROM transactions WHERE currency <> 'RUB') THEN
        RAISE EXCEPTION 'cannot roll back while non-RUB transactions exist';
    END IF;
END $$;

ALTER TABLE transactions DROP COLUMN currency;
ALTER TABLE transactions RENAME COLUMN price TO price_in_rub;
