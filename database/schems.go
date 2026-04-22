package database

var Users = `
CREATE TABLE if not exists agents (
    id SERIAL PRIMARY KEY,
    user_name VARCHAR(255) UNIQUE NOT NULL,
    last_online TIMESTAMP WITH TIME ZONE -- Обновляем при каждом получении лога
);
`
var History_messages = `
CREATE TABLE if not exists agents (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    message VARCHAR(50),
);
`
