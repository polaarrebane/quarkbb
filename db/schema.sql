CREATE TABLE users (
    id             BIGSERIAL PRIMARY KEY,
    username       VARCHAR(200) NOT NULL UNIQUE,
    email          VARCHAR(320) NOT NULL,
    password       VARCHAR(255) NOT NULL
);

CREATE TABLE used_refresh_tokens (
    id             BIGSERIAL PRIMARY KEY,
    jti            UUID NOT NULL UNIQUE
);
CREATE INDEX idx_used_refresh_tokens_jti ON used_refresh_tokens (jti);
