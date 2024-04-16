CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    daily_limit INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS records (
    id INTEGER PRIMARY KEY,
    account_id INTEGER NOT NULL,
    date_record DATETIME NOT NULL,
    date_created DATETIME NOT NULL,
    value INTEGER NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts (id) ON DELETE CASCADE
);
