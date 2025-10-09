/* eslint-disable */
// @ts-nocheck
/**
 * k6 Load Testing Script for House Helper API
 * This is a k6 JavaScript file, not TypeScript
 * VS Code may show false errors - ignore them
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const taskCreationTime = new Trend('task_creation_duration');
const taskRetrievalTime = new Trend('task_retrieval_duration');
const totalRequests = new Counter('total_requests');

// Test configuration
export const options = {
  scenarios: {
    // Smoke test - minimal load
    smoke: {
      executor: 'constant-vus',
      vus: 1,
      duration: '1m',
      tags: { test_type: 'smoke' },
      exec: 'smokeTest',
      startTime: '0s',
    },
    
    // Load test - normal load
    load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 10 },  // Ramp up to 10 users
        { duration: '5m', target: 10 },  // Stay at 10 users
        { duration: '2m', target: 20 },  // Ramp up to 20 users
        { duration: '5m', target: 20 },  // Stay at 20 users
        { duration: '2m', target: 0 },   // Ramp down
      ],
      tags: { test_type: 'load' },
      exec: 'loadTest',
      startTime: '1m',
    },
    
    // Stress test - beyond normal load
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 20 },  // Ramp up to 20 users
        { duration: '5m', target: 50 },  // Ramp up to 50 users
        { duration: '5m', target: 100 }, // Ramp up to 100 users
        { duration: '2m', target: 0 },   // Ramp down
      ],
      tags: { test_type: 'stress' },
      exec: 'stressTest',
      startTime: '17m',
    },
    
    // Spike test - sudden traffic spikes
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 10 },   // Normal load
        { duration: '30s', target: 200 },  // Spike!
        { duration: '1m', target: 200 },   // Stay at spike
        { duration: '10s', target: 10 },   // Return to normal
        { duration: '1m', target: 10 },    // Stay at normal
        { duration: '10s', target: 0 },    // Ramp down
      },
      tags: { test_type: 'spike' },
      exec: 'spikeTest',
      startTime: '31m',
    },
  },
  
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],  // 95% of requests under 500ms, 99% under 1s
    'http_req_failed': ['rate<0.01'],                   // Error rate under 1%
    'errors': ['rate<0.01'],
    'task_creation_duration': ['p(95)<600'],
    'task_retrieval_duration': ['p(95)<300'],
  },
};

// Environment configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'test-api-key';

// Test data generators
function generateEmail() {
  return `test-${Date.now()}-${Math.random().toString(36).substring(7)}@example.com`;
}

function generateTask() {
  const dueDate = new Date();
  dueDate.setDate(dueDate.getDate() + Math.floor(Math.random() * 30) + 1);
  
  return {
    title: `Task ${Date.now()}`,
    description: `Description for task created at ${new Date().toISOString()}`,
    due_date: dueDate.toISOString(),
    assigned_to: 'user-123',  // Would be dynamic in real test
    points: Math.floor(Math.random() * 100) + 1,
  };
}

// Authentication helper
function getAuthToken() {
  const payload = JSON.stringify({
    email: 'test@example.com',
    password: 'testpassword',
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
  };
  
  const res = http.post(`${BASE_URL}/api/v1/auth/login`, payload, params);
  
  if (res.status === 200) {
    const body = JSON.parse(res.body);
    return body.token;
  }
  
  return null;
}

// Test scenarios
export function smokeTest() {
  const token = getAuthToken();
  if (!token) {
    errorRate.add(1);
    return;
  }
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };
  
  // Test health endpoint
  const healthRes = http.get(`${BASE_URL}/health`, params);
  check(healthRes, {
    'health check status is 200': (r) => r.status === 200,
  });
  
  // Test list tasks
  const listRes = http.get(`${BASE_URL}/api/v1/tasks`, params);
  check(listRes, {
    'list tasks status is 200': (r) => r.status === 200,
  });
  
  sleep(1);
}

export function loadTest() {
  const token = getAuthToken();
  if (!token) {
    errorRate.add(1);
    return;
  }
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };
  
  totalRequests.add(1);
  
  // Create task
  const taskPayload = JSON.stringify(generateTask());
  const createStart = Date.now();
  const createRes = http.post(`${BASE_URL}/api/v1/tasks`, taskPayload, params);
  const createDuration = Date.now() - createStart;
  
  const createSuccess = check(createRes, {
    'create task status is 201': (r) => r.status === 201,
    'create task response has id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });
  
  if (createSuccess) {
    taskCreationTime.add(createDuration);
    
    // Get task
    const taskId = JSON.parse(createRes.body).id;
    const getStart = Date.now();
    const getRes = http.get(`${BASE_URL}/api/v1/tasks/${taskId}`, params);
    const getDuration = Date.now() - getStart;
    
    check(getRes, {
      'get task status is 200': (r) => r.status === 200,
      'get task response matches': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.id === taskId;
        } catch (e) {
          return false;
        }
      },
    });
    
    taskRetrievalTime.add(getDuration);
    
    // Update task
    const updatePayload = JSON.stringify({ status: 'in_progress' });
    const updateRes = http.patch(`${BASE_URL}/api/v1/tasks/${taskId}`, updatePayload, params);
    check(updateRes, {
      'update task status is 200': (r) => r.status === 200,
    });
  } else {
    errorRate.add(1);
  }
  
  sleep(1);
}

export function stressTest() {
  // Similar to loadTest but with more aggressive patterns
  const token = getAuthToken();
  if (!token) {
    errorRate.add(1);
    return;
  }
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };
  
  // Batch create tasks
  const batchSize = 5;
  for (let i = 0; i < batchSize; i++) {
    const taskPayload = JSON.stringify(generateTask());
    const createRes = http.post(`${BASE_URL}/api/v1/tasks`, taskPayload, params);
    
    check(createRes, {
      'stress create task status is 201': (r) => r.status === 201,
    }) || errorRate.add(1);
  }
  
  // List tasks (potentially large result set)
  const listRes = http.get(`${BASE_URL}/api/v1/tasks?limit=100`, params);
  check(listRes, {
    'stress list tasks status is 200': (r) => r.status === 200,
  });
  
  sleep(0.5);
}

export function spikeTest() {
  // Similar to loadTest but with minimal sleep
  const token = getAuthToken();
  if (!token) {
    errorRate.add(1);
    return;
  }
  
  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };
  
  // Rapid fire requests
  const taskPayload = JSON.stringify(generateTask());
  const createRes = http.post(`${BASE_URL}/api/v1/tasks`, taskPayload, params);
  
  check(createRes, {
    'spike create task status is 201': (r) => r.status === 201 || r.status === 429,  // Allow rate limit
  }) || errorRate.add(1);
  
  sleep(0.1);
}

// Teardown function
export function teardown(data) {
  console.log('Test completed');
  console.log(`Total requests: ${totalRequests.value}`);
}
