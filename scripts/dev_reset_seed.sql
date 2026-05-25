-- 清空业务表并写入本地开发测试数据（conv_id: group_* / c2c_*）
-- 用法: mysql -uim -pim im < scripts/dev_reset_seed.sql

SET FOREIGN_KEY_CHECKS = 0;

TRUNCATE TABLE messages;
TRUNCATE TABLE conversation_meta;
TRUNCATE TABLE conversation_members;
TRUNCATE TABLE conversations;
TRUNCATE TABLE notifications;
TRUNCATE TABLE group_members;
TRUNCATE TABLE `groups`;
TRUNCATE TABLE friendships;
TRUNCATE TABLE friend_requests;
TRUNCATE TABLE user_devices;
TRUNCATE TABLE users;

SET FOREIGN_KEY_CHECKS = 1;

INSERT INTO users (id, username, password_hash, nickname) VALUES
    (1, 'dev_1', '', 'Dev 1'),
    (2, 'dev_2', '', 'Dev 2'),
    (3, 'dev_3', '', 'Dev 3');

INSERT INTO friendships (user_id, friend_id) VALUES (1, 2), (2, 1);

INSERT INTO `groups` (id, name, owner_id, max_members, notice) VALUES
    (1, '测试群', 1, 10000, '欢迎加入测试群');

INSERT INTO group_members (group_id, user_id, role) VALUES
    (1, 1, 'owner'),
    (1, 2, 'member'),
    (1, 3, 'member');

INSERT INTO conversations (id, type, group_id) VALUES
    ('group_1', 'group', 1),
    ('c2c_1_2', 'c2c', NULL),
    ('c2c_1_3', 'c2c', NULL);

INSERT INTO conversation_members (conv_id, user_id) VALUES
    ('group_1', 1), ('group_1', 2), ('group_1', 3),
    ('c2c_1_2', 1), ('c2c_1_2', 2),
    ('c2c_1_3', 1), ('c2c_1_3', 3);

INSERT INTO conversation_meta (conv_id, last_seq, last_msg_id, last_preview, updated_at) VALUES
    ('group_1', 2, 1002, '大家好', NOW(3)),
    ('c2c_1_2', 1, 2001, '单聊你好', NOW(3) - INTERVAL 1 HOUR),
    ('c2c_1_3', 0, 0, '', NOW(3) - INTERVAL 2 DAY);

INSERT INTO messages (id, conv_id, sender_id, seq, client_msg_id, input) VALUES
    (1001, 'group_1', 1, 1, 'seed-g1-1', '{"input":[{"msgType":"text","content":"{\\"text\\":\\"群聊第一条\\"}"}]}'),
    (1002, 'group_1', 2, 2, 'seed-g1-2', '{"input":[{"msgType":"text","content":"{\\"text\\":\\"大家好\\"}"}]}'),
    (2001, 'c2c_1_2', 1, 1, 'seed-c12-1', '{"input":[{"msgType":"text","content":"{\\"text\\":\\"单聊你好\\"}"}]}');
