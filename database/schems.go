package database

var Users = `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL, -- Наше Name из структуры
    password VARCHAR(255) NOT NULL,
    reg TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
`
var Chat = `
CREATE TABLE IF NOT EXISTS chats (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL -- Название комнаты или группы
);
`
var Chat_members = `
CREATE TABLE if not exists chat_members (
    chat_id INTEGER REFERENCES chats(id),
    user_id INTEGER REFERENCES users(id),
    PRIMARY KEY (chat_id, user_id) -- Пара должна быть уникальной
);`

var Messages = `
CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    chat_id INTEGER REFERENCES chats(id) ON DELETE CASCADE, -- Твой ChatID (лучше INTEGER для связи)
    sender_id INTEGER REFERENCES users(id),                  -- Твой Sender_ID
    message TEXT NOT NULL                                   -- Твое Message
);
`
