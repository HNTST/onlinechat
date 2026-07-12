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
  const u1 = `gm_a_${tag}`;
  const u2 = `gm_b_${tag}`;

  const r1 = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username: u1, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  const r2 = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username: u2, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(r1, { 'setup: register u1 201': (r) => r.status === 201 });
  check(r2, { 'setup: register u2 201': (r) => r.status === 201 });
  if (r1.status !== 201 || r2.status !== 201) return;

  const token = extractToken(r1.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };

  const user1Id = r1.json('id');
  const user2Id = r2.json('id');

  const createChat = http.post(`${BASE_URL}/api/v2/chat`,
    JSON.stringify({
      chat_name: `gm_chat_${tag}`,
      author: user1Id,
      target_user: user2Id,
      is_group: false,
    }),
    authed,
  );
  check(createChat, { 'setup: create chat 201': (r) => r.status === 201 });
  const chatId = createChat.json('chat_id');
  if (!chatId) return;

  const members = http.get(`${BASE_URL}/api/v2/chats/${chatId}/members`, authed);
  check(members, {
    'get members: 200': (r) => r.status === 200,
    'get members: has members array': (r) => Array.isArray(r.json('members')),
  });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noAuth = http.get(`${noAuthUrl}/api/v2/chats/${chatId}/members`);
  check(noAuth, { 'get members: no auth 401': (r) => r.status === 401 });
}
