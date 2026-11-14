ALTER TABLE deposits ADD COLUMN start_date DATE;
ALTER TABLE insurances ADD COLUMN start_date DATE;

ALTER TABLE deposits ADD CONSTRAINT deposits_date_order CHECK (start_date <= end_date);
ALTER TABLE insurances ADD CONSTRAINT insurances_date_order CHECK (start_date <= end_date);
