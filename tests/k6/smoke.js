import http from 'k6/http';
import { check, sleep } from 'k6';

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
  const username = `smoke_${tag}`;

  const reg = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(reg, { 'smoke setup: register 201': (r) => r.status === 201 });

  const token = extractToken(reg.headers['Set-Cookie']);
  const authed = { headers: { 'Cookie': `session_token=${token}` } };

  const userId = reg.json('id');
  const res = http.get(`${BASE_URL}/api/v2/users?id=${userId}`, authed);
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(1);
}
