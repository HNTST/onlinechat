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
  const username = `avget_${tag}`;

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
  check(uploadRes, { 'setup: upload avatar 201': (r) => r.status === 201 });
  if (uploadRes.status !== 201) return;

  const fileId = uploadRes.json('file_id');

  const avatar = http.get(`${BASE_URL}/api/v2/user/avatar?id=${fileId}`, authed);
  check(avatar, {
    'get avatar: 200': (r) => r.status === 200,
    'get avatar: is image': (r) => {
      const ct = r.headers['Content-Type'];
      return ct && (ct.includes('image') || ct.includes('octet-stream'));
    },
  });

  const noAuthUrl = BASE_URL.replace('localhost', '127.0.0.1');
  const noAuth = http.get(`${noAuthUrl}/api/v2/user/avatar?id=${fileId}`);
  check(noAuth, { 'get avatar: no auth 401': (r) => r.status === 401 });

  const missingParam = http.get(`${BASE_URL}/api/v2/user/avatar`, authed);
  check(missingParam, { 'get avatar: missing id 400': (r) => r.status === 400 });

  const notFound = http.get(`${BASE_URL}/api/v2/user/avatar?id=99999`, authed);
  check(notFound, { 'get avatar: not found 404': (r) => r.status === 404 });
}
