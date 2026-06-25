CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    file_key VARCHAR(255) UNIQUE NOT NULL,
    original_file_name VARCHAR(255) NOT NULL,
    type_file VARCHAR(255) NOT NULL,
    size_file BIGINT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    reg TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chats (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS chat_members (
    chat_id INTEGER REFERENCES chats(id),
    user_id INTEGER REFERENCES users(id),
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE,
    sender_id INTEGER REFERENCES users(id),
    message TEXT NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='users' AND column_name='avatar_id'
    ) THEN
        ALTER TABLE users ADD COLUMN avatar_id INTEGER;
    END IF;
END;
$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name='fk_avatar_file' AND table_name='users'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT fk_avatar_file FOREIGN KEY (avatar_id) REFERENCES files(id);
    END IF;
END;
$$;
