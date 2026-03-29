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

CREATE TYPE session_status AS ENUM ('active', 'closed');
CREATE TABLE sessions (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL,
    status         session_status NOT NULL,
    public_id      UUID NOT NULL UNIQUE,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT fk_sessions_users
        FOREIGN KEY (user_id) REFERENCES users(id)

);
CREATE INDEX idx_sessions_public_id ON sessions(public_id);
