import http from 'k6/http';
import { check } from 'k6';

export const options = {
  vus: 1,
  iterations: 1,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const jsonHeaders = { headers: { 'Content-Type': 'application/json' } };
  const tag = `${__VU}_${__ITER}`;
  const username = `reg_${tag}`;

  // === Успешная регистрация ===
  const r = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(r, {
    'register: 201': (r) => r.status === 201,
    'register: cookie set': (r) => !!r.headers['Set-Cookie'],
    'register: id > 0': (r) => r.json('id') > 0,
    'register: correct username': (r) => r.json('username') === username,
    'register: has reg': (r) => !!r.json('reg'),
  });

  // === Дубликат → 400 ===
  const dup = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password: 'otherP@ss' }),
    jsonHeaders,
  );
  check(dup, { 'register duplicate: 400': (r) => r.status === 400 });
}
