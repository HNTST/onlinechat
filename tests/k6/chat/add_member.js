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
  const u1 = `am_a_${tag}`;
  const u2 = `am_b_${tag}`;
  const u3 = `am_c_${tag}`;

  const r1 = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username: u1, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  const r2 = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username: u2, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  const r3 = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username: u3, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(r1, { 'setup: register u1 201': (r) => r.status === 201 });
  check(r2, { 'setup: register u2 201': (r) => r.status === 201 });
  check(r3, { 'setup: register u3 201': (r) => r.status === 201 });
  if (r1.status !== 201 || r2.status !== 201 || r3.status !== 201) return;

  const token = extractToken(r1.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };

  const user1Id = r1.json('id');
  const user2Id = r2.json('id');
  const user3Id = r3.json('id');

  const createChat = http.post(`${BASE_URL}/api/v2/chat`,
    JSON.stringify({
      chat_name: `am_chat_${tag}`,
      author: user1Id,
      target_user: user2Id,
      is_group: false,
    }),
    authed,
  );
  check(createChat, { 'setup: create chat 201': (r) => r.status === 201 });
  const chatId = createChat.json('chat_id');
  if (!chatId) return;

  const addRes = http.post(`${BASE_URL}/api/v2/chat/${chatId}/members`,
    JSON.stringify({ users_ids: [user3Id] }),
    authed,
  );
  check(addRes, {
    'add member: 200': (r) => r.status === 200,
    'add member: returns chat_id': (r) => r.json('Chat_id') === chatId,
    'add member: returns users_ids': (r) => Array.isArray(r.json('Users_ids')),
  });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noAuth = http.post(`${noAuthUrl}/api/v2/chat/${chatId}/members`,
    JSON.stringify({ users_ids: [user3Id] }),
  );
  check(noAuth, { 'add member: no auth 401': (r) => r.status === 401 });
}
