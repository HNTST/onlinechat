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
  const username = `uid_${tag}`;

  const reg = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(reg, { 'setup: register 201': (r) => r.status === 201 });
  if (reg.status !== 201) return;

  const token = extractToken(reg.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };
  const userId = reg.json('id');

  const byUsername = http.get(`${BASE_URL}/api/v2/user_id?username=${username}`, authed);
  check(byUsername, {
    'get user_id by username: 200': (r) => r.status === 200,
    'get user_id: has user_id': (r) => r.json('user_id') > 0,
    'get user_id: matches': (r) => r.json('user_id') === userId,
  });

  const missingParam = http.get(`${BASE_URL}/api/v2/user_id`, authed);
  check(missingParam, { 'get user_id: missing param 500': (r) => r.status === 500 });

  const notFound = http.get(`${BASE_URL}/api/v2/user_id?username=nonexistent_${tag}`, authed);
  check(notFound, { 'get user_id: not found 404': (r) => r.status === 404 });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noAuth = http.get(`${noAuthUrl}/api/v2/user_id?username=${username}`);
  check(noAuth, { 'get user_id: no auth 401': (r) => r.status === 401 });
}
