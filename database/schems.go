package database

var Users = `
CREATE TABLE if not exists users (
    id SERIAL PRIMARY KEY,
    user_name VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    last_online TIMESTAMP WITH TIME ZONE
);
`
var Chat_history = `
CREATE TABLE if not exists chat_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    user_id_second INTEGER NOT NULL,
    message VARCHAR(50)
);
`
