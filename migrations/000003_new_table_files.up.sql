CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    file_key VARCHAR(255) UNIQUE NOT NULL,
    original_file_name VARCHAR(255) NOT NULL,
    type_file VARCHAR(255) NOT NULL,
    size_file BIGINT NOT NULL              -- размер файла
);
