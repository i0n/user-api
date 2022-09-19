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
  group('GET /users?country=USA', () => {
    const response = http.get('http://0.0.0.0:8080/users?country=USA');
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 2 users": res => res.json().length === 2,
    });
  });
  group('GET /users?country=USA&first_name=Hulk', () => {
    const response = http.get('http://0.0.0.0:8080/users?country=USA&first_name=Hulk');
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 1 user": res => res.json().length === 1,
    });
  });
  group('GET /users?country=USA&first_name=Hulk&last_name=Dylan', () => {
    const response = http.get('http://0.0.0.0:8080/users?country=USA&first_name=Hulk&last_name=Dylan');
    check(response, {
      "status code should be 200": res => res.status === 200,
    });
    check(response, {
      "response should have 0 users": res => res.json().length === 0,
    });
  });
}
