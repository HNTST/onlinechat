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
  const u1 = `mr_a_${tag}`;
  const u2 = `mr_b_${tag}`;

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
      chat_name: `mr_chat_${tag}`,
      author: user1Id,
      target_user: user2Id,
      is_group: false,
    }),
    authed,
  );
  check(createChat, { 'setup: create chat 201': (r) => r.status === 201 });
  const chatId = createChat.json('chat_id');
  if (!chatId) return;

  const send = http.post(`${BASE_URL}/api/v2/chats/${chatId}/messages`,
    JSON.stringify({ sender_id: user1Id, message: `mark_${tag}` }),
    authed,
  );
  check(send, { 'setup: send message 201': (r) => r.status === 201 });
  if (send.status !== 201) return;

  const getMsgs = http.get(`${BASE_URL}/api/v2/chats/${chatId}/messages`, authed);
  const arr = getMsgs.json('messages');
  if (!Array.isArray(arr) || arr.length === 0) return;
  const messageId = arr[arr.length - 1].id;
  if (!messageId) return;

  const markRes = http.patch(
    `${BASE_URL}/api/v2/chats/${chatId}/messages/read-status`,
    JSON.stringify({ last_read_message_id: messageId }),
    authed,
  );
  check(markRes, { 'mark read: 200': (r) => r.status === 200 });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noAuth = http.patch(
    `${noAuthUrl}/api/v2/chats/${chatId}/messages/read-status`,
    JSON.stringify({ last_read_message_id: messageId }),
  );
  check(noAuth, { 'mark read: no auth 401': (r) => r.status === 401 });
}
