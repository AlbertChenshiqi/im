-- Users
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(64) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    nickname    VARCHAR(128) NOT NULL DEFAULT '',
    avatar_url  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_devices (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id   VARCHAR(128) NOT NULL,
    push_token  TEXT NOT NULL DEFAULT '',
    platform    VARCHAR(32) NOT NULL DEFAULT 'unknown',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, device_id)
);

-- Friends
CREATE TABLE IF NOT EXISTS friend_requests (
    id          BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL REFERENCES users(id),
    to_user_id   BIGINT NOT NULL REFERENCES users(id),
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (from_user_id, to_user_id)
);

CREATE TABLE IF NOT EXISTS friendships (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, friend_id)
);

-- Groups
CREATE TABLE IF NOT EXISTS groups (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(256) NOT NULL,
    owner_id    BIGINT NOT NULL REFERENCES users(id),
    max_members INT NOT NULL DEFAULT 10000,
    notice      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id    BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(16) NOT NULL DEFAULT 'member',
    muted       BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members(user_id);

-- Conversations
CREATE TABLE IF NOT EXISTS conversations (
    id          VARCHAR(64) PRIMARY KEY,
    type        VARCHAR(16) NOT NULL,
    group_id    BIGINT REFERENCES groups(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS conversation_members (
    conv_id     VARCHAR(64) NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pinned      BOOLEAN NOT NULL DEFAULT FALSE,
    muted       BOOLEAN NOT NULL DEFAULT FALSE,
    last_read_seq BIGINT NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conv_id, user_id)
);

CREATE TABLE IF NOT EXISTS conversation_meta (
    conv_id     VARCHAR(64) PRIMARY KEY REFERENCES conversations(id) ON DELETE CASCADE,
    last_seq    BIGINT NOT NULL DEFAULT 0,
    last_msg_id BIGINT NOT NULL DEFAULT 0,
    last_preview TEXT NOT NULL DEFAULT '',
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Messages (partition-friendly by conv_id hash in app layer)
CREATE TABLE IF NOT EXISTS messages (
    id          BIGINT PRIMARY KEY,
    conv_id     VARCHAR(64) NOT NULL,
    sender_id   BIGINT NOT NULL,
    seq         BIGINT NOT NULL,
    client_msg_id VARCHAR(128) NOT NULL DEFAULT '',
    msg_type    VARCHAR(32) NOT NULL DEFAULT 'text',
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (conv_id, seq)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_client_dedupe
    ON messages (conv_id, client_msg_id) WHERE client_msg_id <> '';

CREATE INDEX IF NOT EXISTS idx_messages_conv_seq ON messages(conv_id, seq DESC);

-- Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(256) NOT NULL,
    body        TEXT NOT NULL,
    category    VARCHAR(64) NOT NULL DEFAULT 'system',
    read        BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, created_at DESC);
