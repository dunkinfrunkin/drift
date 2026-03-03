CREATE OR REPLACE VIEW active_users AS
SELECT id, email, name, created_at
FROM users
WHERE status = 'active';
