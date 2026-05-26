-- 已有库：删除 001_init 旧版外键（新库请直接用无 FK 的 001_init）
-- 若某约束不存在会报错，可忽略该行后重跑其余语句。

ALTER TABLE user_devices DROP FOREIGN KEY fk_user_devices_user;

ALTER TABLE friend_requests DROP FOREIGN KEY fk_friend_req_from;
ALTER TABLE friend_requests DROP FOREIGN KEY fk_friend_req_to;

ALTER TABLE friendships DROP FOREIGN KEY fk_friendships_user;
ALTER TABLE friendships DROP FOREIGN KEY fk_friendships_friend;

ALTER TABLE `groups` DROP FOREIGN KEY fk_groups_owner;

ALTER TABLE group_members DROP FOREIGN KEY fk_group_members_group;
ALTER TABLE group_members DROP FOREIGN KEY fk_group_members_user;

ALTER TABLE conversations DROP FOREIGN KEY fk_conversations_group;

ALTER TABLE conversation_members DROP FOREIGN KEY fk_conv_members_conv;
ALTER TABLE conversation_members DROP FOREIGN KEY fk_conv_members_user;

ALTER TABLE conversation_meta DROP FOREIGN KEY fk_conv_meta_conv;

ALTER TABLE notifications DROP FOREIGN KEY fk_notifications_user;
