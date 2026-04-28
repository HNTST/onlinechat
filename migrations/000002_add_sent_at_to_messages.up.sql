-- 1. Добавляем колонку, разрешая NULL на долю секунды
ALTER TABLE messages ADD COLUMN sent_at TIMESTAMP WITH TIME ZONE;

-- 2. Заполняем пустые места текущим временем (для старых записей)
UPDATE messages SET sent_at = CURRENT_TIMESTAMP WHERE sent_at IS NULL;

-- 3. Теперь, когда пустых мест нет, ставим запрет на NULL
ALTER TABLE messages ALTER COLUMN sent_at SET NOT NULL;
