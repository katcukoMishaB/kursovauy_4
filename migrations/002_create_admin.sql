-- Для создания администратора используйте скрипт: go run scripts/create_admin.go
-- Или создайте пользователя вручную через интерфейс и затем обновите роль через SQL:
-- UPDATE user_roles SET is_admin = true, is_organizer = true WHERE user_id = '<user_id>';

INSERT INTO user_roles (user_id, is_participant, is_organizer, is_admin)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    true,
    true,
    true
) ON CONFLICT (user_id) DO UPDATE SET is_admin = true;

