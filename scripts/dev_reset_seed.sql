-- 清空业务表并写入本地开发测试数据（conv_id: group_* / c2c_*）
-- 用法: psql "$POSTGRES_DSN" -f scripts/dev_reset_seed.sql

BEGIN;

TRUNCATE TABLE
    messages,
    conversation_meta,
    conversation_members,
    conversations,
    notifications,
    group_members,
    groups,
    friendships,
    friend_requests,
    user_devices,
    users
RESTART IDENTITY CASCADE;

-- 用户 1–3（与 dev-token user_id 对齐）
INSERT INTO users (id, username, password_hash, nickname) VALUES
    (1, 'dev_1', '', 'Dev 1'),
    (2, 'dev_2', '', 'Dev 2'),
    (3, 'dev_3', '', 'Dev 3');
SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT MAX(id) FROM users));

-- 好友：1 <-> 2
INSERT INTO friendships (user_id, friend_id) VALUES (1, 2), (2, 1);

-- 群 1：成员 1（群主）、2、3
INSERT INTO groups (id, name, owner_id, max_members, notice) VALUES
    (1, '测试群', 1, 10000, '欢迎加入测试群');
SELECT setval(pg_get_serial_sequence('groups', 'id'), (SELECT MAX(id) FROM groups));

INSERT INTO group_members (group_id, user_id, role) VALUES
    (1, 1, 'owner'),
    (1, 2, 'member'),
    (1, 3, 'member');

-- 会话：群 + 私信（含非好友 1-3）
INSERT INTO conversations (id, type, group_id) VALUES
    ('group_1', 'group', 1),
    ('c2c_1_2', 'c2c', NULL),
    ('c2c_1_3', 'c2c', NULL);

INSERT INTO conversation_members (conv_id, user_id) VALUES
    ('group_1', 1), ('group_1', 2), ('group_1', 3),
    ('c2c_1_2', 1), ('c2c_1_2', 2),
    ('c2c_1_3', 1), ('c2c_1_3', 3);

INSERT INTO conversation_meta (conv_id, last_seq, last_msg_id, last_preview, updated_at) VALUES
    ('group_1', 2, 1002, '大家好', NOW()),
    ('c2c_1_2', 1, 2001, '单聊你好', NOW() - INTERVAL '1 hour'),
    ('c2c_1_3', 0, 0, '', NOW() - INTERVAL '2 days');

-- 示例消息（可选，便于拉历史）
INSERT INTO messages (id, conv_id, sender_id, seq, client_msg_id, msg_type, content) VALUES
    (1001, 'group_1', 1, 1, 'seed-g1-1', 'text', '群聊第一条'),
    (1002, 'group_1', 2, 2, 'seed-g1-2', 'text', '大家好'),
    (2001, 'c2c_1_2', 1, 1, 'seed-c12-1', 'text', '单聊你好');

COMMIT;
