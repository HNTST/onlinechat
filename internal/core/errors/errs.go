package core_errors

import "errors"

var (
	ErrUserAlreadyExists = errors.New("Пользователь уже существует")
	ErrInternal          = errors.New("Внутренняя ошибка сервера")
	ErrInvalidJSON       = errors.New("Неверный формат JSON")
	ErrParamMissing      = errors.New("Отсутствует обязательный параметр")
	ErrEncodingRequest   = errors.New("Ошибка кодирования JSON ответа")

	ErrUserNotFound    = errors.New("не найден")
	ErrInvalidArgument = errors.New("неверный аргумент")
	ErrConfict         = errors.New("конфликт")

	ErrChatNotExists    = errors.New("не существует")
	ErrMessageNotFound  = errors.New("сообщение не найдено")
)
