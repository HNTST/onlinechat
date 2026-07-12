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
  const username = `login_${tag}`;

  // Сначала регистрируем пользователя
  const reg = http.post(`${BASE_URL}/api/v2/auth/register`,
    JSON.stringify({ username, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(reg, { 'login setup: register ok': (r) => r.status === 201 });
  if (reg.status !== 201) return;

  // === Логин с верными данными ===
  const loginOk = http.post(`${BASE_URL}/api/v2/auth/login`,
    JSON.stringify({ username, password: 'str0ngP@ss' }),
    jsonHeaders,
  );
  check(loginOk, {
    'login: 200': (r) => r.status === 200,
    'login: cookie set': (r) => !!r.headers['Set-Cookie'],
    'login: correct username': (r) => r.json('username') === username,
    'login: returns id': (r) => r.json('id') > 0,
  });

  // === Профиль с полученной кукой ===
  const token = extractToken(loginOk.headers['Set-Cookie']);
  const profileOk = http.get(`${BASE_URL}/api/v2/user/profile`, {
    headers: { 'Cookie': `session_token=${token}` },
  });
  check(profileOk, {
    'login: profile 200': (r) => r.status === 200,
    'login: profile correct user': (r) => r.json('username') === username,
  });

  // === Логин с неверным паролем → 401 ===
  const badPass = http.post(`${BASE_URL}/api/v2/auth/login`,
    JSON.stringify({ username, password: 'wrongpass' }),
    jsonHeaders,
  );
  check(badPass, { 'login: bad password 401': (r) => r.status === 401 });

  // === Логин несуществующего пользователя → 401 ===
  const noUser = http.post(`${BASE_URL}/api/v2/auth/login`,
    JSON.stringify({ username: `never_existed_${tag}`, password: 'x' }),
    jsonHeaders,
  );
  check(noUser, { 'login: no user 401': (r) => r.status === 401 });
}
