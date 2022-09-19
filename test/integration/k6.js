import http from 'k6/http';
import { check, group } from 'k6';

export let options = {
    vus: 1,
};

export default function () {
  group('API health check', () => {
    const response = http.get('http://0.0.0.0:8080/');
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
  });
  group('GET /users', () => {
    const response = http.get('http://0.0.0.0:8080/users');
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 4 users": res => res.json().length === 4,
    });
  });
}
