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

function randName(tag) {
  return `usr_${tag}`;
}

export default function () {
  const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };
  const tag = `${__VU}_${__ITER}`;
  const username = randName(tag);
  const password = 'str0ngP@ss';

  const reg = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password }),
    jsonHeaders,
  );
  check(reg, { 'profile setup: register 201': (r) => r.status === 201 });
  if (reg.status !== 201) return;

  const token = extractToken(reg.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };

  const profile = http.get(`${BASE_URL}/api/v2/user/profile`, authed);
  check(profile, {
    'profile: 200': (r) => r.status === 200,
    'profile: correct username': (r) => r.json('username') === username,
    'profile: id > 0': (r) => r.json('id') > 0,
    'profile: has reg': (r) => !!r.json('reg'),
  });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noAuth = http.get(`${noAuthUrl}/api/v2/user/profile`);
  check(noAuth, { 'profile: no auth 401': (r) => r.status === 401 });

  const userId = profile.json('id');
  if (!userId) return;

  const userById = http.get(`${BASE_URL}/api/v2/users?id=${userId}`, authed);
  check(userById, {
    'users: 200': (r) => r.status === 200,
    'users: correct username': (r) => r.json('username') === username,
    'users: id matches': (r) => r.json('id') === userId,
  });

  const noId = http.get(`${BASE_URL}/api/v2/users`, authed);
  check(noId, { 'users: missing id 400': (r) => r.status === 400 });

  const notFound = http.get(`${BASE_URL}/api/v2/users?id=99999`, authed);
  check(notFound, { 'users: not found 404': (r) => r.status === 404 });
}
