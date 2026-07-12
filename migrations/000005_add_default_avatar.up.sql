INSERT INTO files (id, file_key, original_file_name, type_file, size_file)
VALUES (1, 'avatars/default.jpg', 'default.png', 'image/png', 0)
ON CONFLICT (id) DO NOTHING;

SELECT setval('files_id_seq', GREATEST((SELECT MAX(id) FROM files), 1));