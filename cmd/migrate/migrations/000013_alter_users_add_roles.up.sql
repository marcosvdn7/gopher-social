ALTER TABLE IF EXISTS users
ADD COLUMN role_id BIGSERIAL DEFAULT 1;

UPDATE users
SET role_id = (SELECT id from roles where name = 'user');

ALTER TABLE users
ADD CONSTRAINT fk_users_role_id FOREIGN KEY (role_id) REFERENCES roles(id);