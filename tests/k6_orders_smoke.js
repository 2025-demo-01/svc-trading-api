import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = { vus: 5, duration: '30s' };

export default function () {
  const url = __ENV.API_BASE + '/api/v1/trade/orders';
  const payload = JSON.stringify({ symbol: 'BTCUSDT', side: 'buy', price: 50000, qty: 0.01 });
  const headers = { 'Content-Type': 'application/json', 'Idempotency-Key': Math.random().toString(36).slice(2) };
  const res = http.post(url, payload, { headers });
  check(res, { 'status 200': r => r.status === 200 });
  sleep(1);
}
