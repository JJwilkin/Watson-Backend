CREATE TABLE IF NOT EXISTS users (
    user_id serial PRIMARY KEY,
    email VARCHAR(300) UNIQUE NOT NULL,
    password VARCHAR(50) NOT NULL
)