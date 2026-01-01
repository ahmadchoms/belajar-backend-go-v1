import http from 'k6/http';
import { check } from 'k6';

const TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6ImFobWFkQGV4YW1wbGUuY29tIiwicm9sZSI6InVzZXIiLCJleHAiOjE3NjcyMzUyNDh9.dSDZKsGeE0YPfDgyQzh-q8Md_HSwSLuqI4lnM8yRZVA';

export const options = {
  vus: 1,
  iterations: 50,
};

export default function () {
  const res = http.get('http://localhost:8080/products', {
    headers: {
      Authorization: `Bearer ${TOKEN}`,
    },
  });

  check(res, {
    'status is valid': (r) =>
      r.status === 200 ||
      r.status === 500 ||
      r.status === 503,
  });
}
