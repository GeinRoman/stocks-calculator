CREATE TABLE profiles (
    id      SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL
);
