-- users.avatar_url 补默认值，避免 INSERT 未带该列时报 1364
ALTER TABLE users
    MODIFY COLUMN avatar_url TEXT NOT NULL DEFAULT '' COMMENT '头像 URL';
