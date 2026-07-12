import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: parseInt(__ENV.VUS),
  duration: __ENV.DURATION,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

const pngData = open('../fixtures/test.png', 'b');

function extractToken(setCookie) {
  if (!setCookie) return '';
  const m = setCookie.match(/session_token=([^;]+)/);
  return m ? m[1] : '';
}

export default function () {
  const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };
  const tag = `${__VU}_${__ITER}`;
  const username = `avup_${tag}`;

  const reg = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(reg, { 'setup: register 201': (r) => r.status === 201 });
  if (reg.status !== 201) return;

  const token = extractToken(reg.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };

  const uploadRes = http.post(
    `${BASE_URL}/api/v2/user/avatar`,
    { new_user_avatar: http.file(pngData, 'test.png') },
    { headers: { 'Cookie': `session_token=${token}` } },
  );

  check(uploadRes, {
    'upload avatar: 201': (r) => r.status === 201,
    'upload avatar: has file_id': (r) => {
      try { return r.json('file_id') > 0; } catch (e) { return false; }
    },
  });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noAuth = http.post(`${noAuthUrl}/api/v2/user/avatar`, {
    new_user_avatar: http.file(pngData, 'test.png'),
  });
  check(noAuth, { 'upload avatar: no auth 401': (r) => r.status === 401 });
}
