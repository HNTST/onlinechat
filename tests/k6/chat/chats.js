import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: parseInt(__ENV.VUS),
  duration: __ENV.DURATION,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

function extractToken(setCookie) {
  if (!setCookie) return '';
  const m = setCookie.match(/session_token=([^;]+)/);
  return m ? m[1] : '';
}

export default function () {
  const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };
  const tag = `${__VU}_${__ITER}`;

  const u1 = `chat_a_${tag}`;
  const u2 = `chat_b_${tag}`;

  const r1 = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username: u1, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  const r2 = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username: u2, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(r1, { 'chats setup: register user1 201': (r) => r.status === 201 });
  check(r2, { 'chats setup: register user2 201': (r) => r.status === 201 });
  if (r1.status !== 201 || r2.status !== 201) return;

  const token = extractToken(r1.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };

  const user1Id = r1.json('id');
  const user2Id = r2.json('id');
  if (!user1Id || !user2Id) return;

  const createChat = http.post(`${BASE_URL}/api/v2/chat`,
    JSON.stringify({
      chat_name: `test_chat_${tag}`,
      author: user1Id,
      target_user: user2Id,
      is_group: false,
    }),
    authed,
  );
  check(createChat, {
    'create chat: 201': (r) => r.status === 201,
    'create chat: has chat_id': (r) => r.json('chat_id') > 0,
  });
  const chatId = createChat.json('chat_id');
  if (!chatId) return;

  const chats = http.get(`${BASE_URL}/api/v2/chats`, authed);
  check(chats, {
    'get chats: 200': (r) => r.status === 200,
    'get chats: has chats array': (r) => Array.isArray(r.json('chats')),
    'get chats: contains created': (r) => {
      const arr = r.json('chats');
      return Array.isArray(arr) && arr.some(c => c.id === chatId);
    },
  });

  const dup = http.post(`${BASE_URL}/api/v2/chat`,
    JSON.stringify({
      chat_name: `test_chat_${tag}`,
      author: user1Id,
      target_user: user2Id,
      is_group: false,
    }),
    authed,
  );
  check(dup, {
    'create chat duplicate: 201': (r) => r.status === 201,
    'create chat duplicate: same chat_id': (r) => r.json('chat_id') === chatId,
  });
}
