WebSocket /ws

POST /chat - сохраняет чат 
  JSON {
    `json:"chat_name"` строка
    `json:"author"` int
  } 
  Возвращает id чата
  JSON { "id": id int }

POST /user - сохраняет пользователя
  JSON {
    `json:"username"` строка
    `json:"password"` строка
  }
  Возвращает id пользователя
  JSON { "id": id int }

GET /chats - возвращает чаты пользователя
  JSON {
    `json:"user_id"` int 
  }
  Возвращает Массив JSON 
  [
    {
    "id": id int - чата
    "name": строка название чата
    },
  ]

GET /user возвращает пользователя
  JSON {
    `json:"user_id"` int 
  }
  Возвращает JSON {
  "id":0, int
  "username":"Ivan", string
  "password":"", string
  "reg":"2026-04-23T06:23:38.552926Z" date
  }



"POST /chat"
"POST /user"
"GET /user"
"GET /chats"
