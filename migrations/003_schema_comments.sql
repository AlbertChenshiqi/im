-- 表/字段注释（PostgreSQL COMMENT ON）
-- 新库随 docker-entrypoint-initdb.d 执行；已有库: psql ... -f migrations/003_schema_comments.sql

-- users
COMMENT ON TABLE users IS '用户账号';
COMMENT ON COLUMN users.id IS '用户主键';
COMMENT ON COLUMN users.username IS '登录名，全局唯一';
COMMENT ON COLUMN users.password_hash IS '密码哈希；开发期 dev 用户可为空';
COMMENT ON COLUMN users.nickname IS '昵称';
COMMENT ON COLUMN users.avatar_url IS '头像 URL';
COMMENT ON COLUMN users.created_at IS '创建时间';
COMMENT ON COLUMN users.updated_at IS '更新时间';

-- user_devices
COMMENT ON TABLE user_devices IS '用户推送设备';
COMMENT ON COLUMN user_devices.user_id IS '用户 ID';
COMMENT ON COLUMN user_devices.device_id IS '设备标识';
COMMENT ON COLUMN user_devices.push_token IS 'APNs/FCM 等推送 Token';
COMMENT ON COLUMN user_devices.platform IS '平台：ios/android/web 等';
COMMENT ON COLUMN user_devices.updated_at IS 'Token 更新时间';

-- friend_requests
COMMENT ON TABLE friend_requests IS '好友申请';
COMMENT ON COLUMN friend_requests.id IS '申请主键';
COMMENT ON COLUMN friend_requests.from_user_id IS '发起人';
COMMENT ON COLUMN friend_requests.to_user_id IS '接收人';
COMMENT ON COLUMN friend_requests.status IS '状态：pending/accepted/rejected';
COMMENT ON COLUMN friend_requests.created_at IS '申请时间';

-- friendships
COMMENT ON TABLE friendships IS '好友关系（单向一行，互为好友需两行）';
COMMENT ON COLUMN friendships.user_id IS '用户 ID';
COMMENT ON COLUMN friendships.friend_id IS '好友用户 ID';
COMMENT ON COLUMN friendships.created_at IS '成为好友时间';

-- groups
COMMENT ON TABLE groups IS '群组';
COMMENT ON COLUMN groups.id IS '群 ID';
COMMENT ON COLUMN groups.name IS '群名称';
COMMENT ON COLUMN groups.owner_id IS '群主用户 ID';
COMMENT ON COLUMN groups.max_members IS '成员上限';
COMMENT ON COLUMN groups.notice IS '群公告';
COMMENT ON COLUMN groups.created_at IS '创建时间';
COMMENT ON COLUMN groups.updated_at IS '更新时间';

-- group_members
COMMENT ON TABLE group_members IS '群成员';
COMMENT ON COLUMN group_members.group_id IS '群 ID';
COMMENT ON COLUMN group_members.user_id IS '成员用户 ID';
COMMENT ON COLUMN group_members.role IS '角色：owner/admin/member';
COMMENT ON COLUMN group_members.muted IS '是否屏蔽群消息提醒';
COMMENT ON COLUMN group_members.joined_at IS '入群时间';

-- conversations
COMMENT ON TABLE conversations IS '会话；id 即 conv_id，全局唯一';
COMMENT ON COLUMN conversations.id IS '会话 ID：群 group_{group_id}，私信 c2c_{小uid}_{大uid}';
COMMENT ON COLUMN conversations.type IS '会话类型：group 群聊，c2c 单聊（兼容旧值 direct）';
COMMENT ON COLUMN conversations.group_id IS '群聊时关联 groups.id；私信为 NULL';
COMMENT ON COLUMN conversations.created_at IS '会话创建时间';

-- conversation_members
COMMENT ON TABLE conversation_members IS '用户与会话关系（会话列表成员维度）';
COMMENT ON COLUMN conversation_members.conv_id IS '会话 ID，同 conversations.id';
COMMENT ON COLUMN conversation_members.user_id IS '用户 ID';
COMMENT ON COLUMN conversation_members.pinned IS '是否置顶';
COMMENT ON COLUMN conversation_members.muted IS '是否免打扰';
COMMENT ON COLUMN conversation_members.last_read_seq IS '已读到的消息序号';
COMMENT ON COLUMN conversation_members.updated_at IS '成员会话状态更新时间';

-- conversation_meta
COMMENT ON TABLE conversation_meta IS '会话摘要（最后一条消息等）';
COMMENT ON COLUMN conversation_meta.conv_id IS '会话 ID';
COMMENT ON COLUMN conversation_meta.last_seq IS '当前会话最大消息序号';
COMMENT ON COLUMN conversation_meta.last_msg_id IS '最后一条消息 ID';
COMMENT ON COLUMN conversation_meta.last_preview IS '列表展示用的最后一条预览文案';
COMMENT ON COLUMN conversation_meta.updated_at IS '摘要更新时间';

-- messages
COMMENT ON TABLE messages IS '消息正文（按 conv_id 存储）';
COMMENT ON COLUMN messages.id IS '消息 ID（雪花）';
COMMENT ON COLUMN messages.conv_id IS '所属会话 ID';
COMMENT ON COLUMN messages.sender_id IS '发送者用户 ID';
COMMENT ON COLUMN messages.seq IS '会话内递增序号';
COMMENT ON COLUMN messages.client_msg_id IS '客户端幂等 ID；非空时在 (conv_id, client_msg_id) 上部分唯一';
COMMENT ON INDEX idx_messages_client_dedupe IS 'client_msg_id 非空时的幂等去重索引';
COMMENT ON COLUMN messages.msg_type IS '消息类型：text 等';
COMMENT ON COLUMN messages.content IS '消息内容';
COMMENT ON COLUMN messages.created_at IS '发送时间';

-- notifications
COMMENT ON TABLE notifications IS '系统通知收件箱';
COMMENT ON COLUMN notifications.id IS '通知主键';
COMMENT ON COLUMN notifications.user_id IS '接收用户 ID';
COMMENT ON COLUMN notifications.title IS '标题';
COMMENT ON COLUMN notifications.body IS '正文';
COMMENT ON COLUMN notifications.category IS '分类：system 等';
COMMENT ON COLUMN notifications.read IS '是否已读';
COMMENT ON COLUMN notifications.created_at IS '创建时间';
