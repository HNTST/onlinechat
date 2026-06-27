ALTER TABLE users ADD COLUMN avatar_id INTEGER;

-- 2. Добавляем связь с таблицей files
ALTER TABLE users ADD CONSTRAINT fk_avatar_file FOREIGN KEY (avatar_id) REFERENCES files(id);
