-- MySQL 8+ 初始化（utf8mb4）
-- 约定：不使用外键，引用完整性由应用层保证。

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS users (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '用户主键',
    username      VARCHAR(64) NOT NULL COMMENT '登录名，全局唯一',
    password_hash TEXT NOT NULL COMMENT '密码哈希',
    nickname      VARCHAR(128) NOT NULL DEFAULT '' COMMENT '昵称',
    avatar_url    TEXT NOT NULL DEFAULT '' COMMENT '头像 URL',
    created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    updated_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_devices (
    user_id     BIGINT NOT NULL COMMENT '用户 ID',
    device_id   VARCHAR(128) NOT NULL COMMENT '设备标识',
    push_token  TEXT NOT NULL COMMENT '推送 Token',
    platform    VARCHAR(32) NOT NULL DEFAULT 'unknown' COMMENT '平台',
    updated_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (user_id, device_id),
    KEY idx_user_devices_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS friend_requests (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    from_user_id  BIGINT NOT NULL,
    to_user_id    BIGINT NOT NULL,
    status        VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_friend_req_pair (from_user_id, to_user_id),
    KEY idx_friend_req_to (to_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS friendships (
    user_id     BIGINT NOT NULL,
    friend_id   BIGINT NOT NULL,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (user_id, friend_id),
    KEY idx_friendships_friend (friend_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `groups` (
    id            BIGINT AUTO_INCREMENT PRIMARY KEY,
    name          VARCHAR(256) NOT NULL,
    owner_id      BIGINT NOT NULL,
    max_members   INT NOT NULL DEFAULT 10000,
    notice        TEXT NOT NULL,
    created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    KEY idx_groups_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS group_members (
    group_id    BIGINT NOT NULL,
    user_id     BIGINT NOT NULL,
    role        VARCHAR(16) NOT NULL DEFAULT 'member',
    muted       TINYINT(1) NOT NULL DEFAULT 0,
    joined_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (group_id, user_id),
    KEY idx_group_members_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS conversations (
    id          VARCHAR(64) PRIMARY KEY,
    type        VARCHAR(16) NOT NULL,
    group_id    BIGINT NULL,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_conversations_group (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS conversation_members (
    conv_id         VARCHAR(64) NOT NULL,
    user_id         BIGINT NOT NULL,
    pinned          TINYINT(1) NOT NULL DEFAULT 0,
    muted           TINYINT(1) NOT NULL DEFAULT 0,
    last_read_seq   BIGINT NOT NULL DEFAULT 0,
    updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (conv_id, user_id),
    KEY idx_conv_members_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS conversation_meta (
    conv_id       VARCHAR(64) PRIMARY KEY,
    last_seq      BIGINT NOT NULL DEFAULT 0,
    last_msg_id   BIGINT NOT NULL DEFAULT 0,
    last_preview  TEXT NOT NULL,
    updated_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS messages (
    id            BIGINT PRIMARY KEY,
    conv_id       VARCHAR(64) NOT NULL,
    sender_id     BIGINT NOT NULL,
    seq           BIGINT NOT NULL,
    client_msg_id VARCHAR(128) NOT NULL DEFAULT '',
    input         JSON NOT NULL COMMENT '消息体 {"input":[{"msgType","content"}]}',
    created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_messages_conv_seq (conv_id, seq),
    KEY idx_messages_conv_seq (conv_id, seq DESC),
    KEY idx_messages_sender (sender_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notifications (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id     BIGINT NOT NULL,
    title       VARCHAR(256) NOT NULL,
    body        TEXT NOT NULL,
    category    VARCHAR(64) NOT NULL DEFAULT 'system',
    is_read     TINYINT(1) NOT NULL DEFAULT 0,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_notifications_user (user_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
