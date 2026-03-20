CREATE TABLE users (
    id             BIGSERIAL PRIMARY KEY,
    username       VARCHAR(200) NOT NULL UNIQUE,
    email          VARCHAR(320) NOT NULL,
    password       VARCHAR(255) NOT NULL
);
