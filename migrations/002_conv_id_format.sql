-- 可选：将旧 conv_id（g_* / d_*）迁移为新格式（group_* / c2c_*）
-- 仅在已有开发数据且需统一 ID 时执行；新环境直接用 001_init + 新格式即可。

BEGIN;

UPDATE conversations SET type = 'c2c' WHERE type = 'direct';

-- 群：g_{id} -> group_{id}
UPDATE conversations
SET id = 'group_' || substring(id from 3)
WHERE id LIKE 'g\_%' ESCAPE '\';

UPDATE conversation_members cm
SET conv_id = 'group_' || substring(cm.conv_id from 3)
WHERE cm.conv_id LIKE 'g\_%' ESCAPE '\';

UPDATE conversation_meta m
SET conv_id = 'group_' || substring(m.conv_id from 3)
WHERE m.conv_id LIKE 'g\_%' ESCAPE '\';

UPDATE messages
SET conv_id = 'group_' || substring(conv_id from 3)
WHERE conv_id LIKE 'g\_%' ESCAPE '\';

-- 私信：d_{a}_{b} -> c2c_{a}_{b}
UPDATE conversations
SET id = 'c2c_' || substring(id from 3)
WHERE id LIKE 'd\_%' ESCAPE '\';

UPDATE conversation_members cm
SET conv_id = 'c2c_' || substring(cm.conv_id from 3)
WHERE cm.conv_id LIKE 'd\_%' ESCAPE '\';

UPDATE conversation_meta m
SET conv_id = 'c2c_' || substring(m.conv_id from 3)
WHERE m.conv_id LIKE 'd\_%' ESCAPE '\';

UPDATE messages
SET conv_id = 'c2c_' || substring(conv_id from 3)
WHERE conv_id LIKE 'd\_%' ESCAPE '\';

COMMIT;
