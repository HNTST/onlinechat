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
  const username = `logout_${tag}`;

  const reg = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(reg, { 'logout setup: register 201': (r) => r.status === 201 });
  if (reg.status !== 201) return;

  const token = extractToken(reg.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };

  const before = http.get(`${BASE_URL}/api/v2/user/profile`, authed);
  check(before, { 'logout: profile before 200': (r) => r.status === 200 });

  const logout = http.post(`${BASE_URL}/api/v2/auth/logout`, null, authed);
  check(logout, { 'logout: 200': (r) => r.status === 200 });

  const after = http.get(`${BASE_URL}/api/v2/user/profile`, authed);
  check(after, { 'logout: profile after 401': (r) => r.status === 401 });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noCookie = http.post(`${noAuthUrl}/api/v2/auth/logout`, null);
  check(noCookie, { 'logout: no cookie 401': (r) => r.status === 401 });

  const noAuth = http.get(`${noAuthUrl}/api/v2/user/profile`);
  check(noAuth, { 'logout: no auth 401': (r) => r.status === 401 });
}
