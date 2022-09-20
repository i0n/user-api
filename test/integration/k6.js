import http from 'k6/http';
import { check, group } from 'k6';

export let options = {
    vus: 1,
    thresholds: {
      // the rate of successful checks should be 100%
      checks: ['rate>=1'],
    },
};

export default function () {
  group('API health check', () => {
    const response = http.get(`http://${__ENV.USER_API_URL}/`);
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
  });

  group('GET /users', () => {
    const response = http.get(`http://${__ENV.USER_API_URL}/users`);
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 4 users": res => res.json().length === 4,
    });
  });
  group('GET /users?country=USA', () => {
    const response = http.get(`http://${__ENV.USER_API_URL}/users?country=USA`);
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 2 users": res => res.json().length === 2,
    });
  });
  group('GET /users?country=USA&first_name=Hulk', () => {
    const response = http.get(`http://${__ENV.USER_API_URL}/users?country=USA&first_name=Hulk`);
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 1 user": res => res.json().length === 1,
    });
  });
  group('GET /users?country=USA&first_name=Hulk&last_name=Dylan', () => {
    const response = http.get(`http://${__ENV.USER_API_URL}/users?country=USA&first_name=Hulk&last_name=Dylan`);
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 0 users": res => res.json().length === 0,
    });
  });

  group('POST /user', () => {
    const response = http.post(`http://${__ENV.USER_API_URL}/user`);
    check(response, {
      "status code should be 400": res => res.status === 400,
    });
  });

  group('POST /user with email', () => {
    const payload = {
      email: 'test@testing.com'
    };
    const response = http.post(`http://${__ENV.USER_API_URL}/user`, payload);
    check(response, {
      "status code should be 201": res => res.status === 201,
    });
  });

  group('POST /user with duplicate email', () => {
    const payload = {
      email: 'test@testing.com'
    };
    const response = http.post(`http://${__ENV.USER_API_URL}/user`, payload);
    check(response, {
      "status code should be 500": res => res.status === 500,
    });
  });

  group('PATCH /user/:id', () => {
    const response = http.patch(`http://${__ENV.USER_API_URL}/user/5`);
    check(response, {
      "status code should be 400": res => res.status === 400,
    });
  });

  group('PATCH /user/:id with valid kv pairs', () => {
    const payload = {
      first_name: 'Tom',
      last_name: 'Hanks',
      email: 'tom@hanks.com'
    };
    const response = http.patch(`http://${__ENV.USER_API_URL}/user/5`, payload);
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
  });

  group('DELETE /user/:id', () => {
    const response = http.del(`http://${__ENV.USER_API_URL}/user/5`);
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
  });

  group('DELETE /user/:id with an invalid id', () => {
    const response = http.del(`http://${__ENV.USER_API_URL}/user/notanid`);
    console.log(response.body);
    check(response, {
      "status code should be 500": res => res.status === 500,
    });
  });
}
