# Benchmark: gRPC vs HTTP

## Executive Summary

This benchmark evaluates the performance benefits of migrating the authorization service from **HTTP/JSON** to 
**gRPC/Protobuf**. Under controlled conditions with `GOMAXPROCS=2`, gRPC demonstrated **3.7x higher throughput** and 
**92.5% lower P95 latency** compared to HTTP. These metrics establish a performance baseline for future optimizations 
like Redis multi-level caching.

## Test Strategy & Environment

### Runtime Configuration

| Config        | Parameter                    | Description                                                                     |
| ------------- | ---------------------------- | ------------------------------------------------------------------------------- |
| CPU Limit     | `GOMAXPROCS=2`               | Simulates resource-constrained environment to stress protocol efficiency        |
| OpenFGA Cache | `enabled` (always hit cache) | Eliminates database query bottleneck, isolating protocol/serialization overhead |
| Network       | `localhost`                  | Eliminates network transmission latency, focusing on processing efficiency      |

### Pressure measurement strategy (k6)

k6 was used for stepped load testing without sleep intervals to stress the system to its saturation point.

| Stage | Duration | Virtual Users | Description             |
| ----- | -------- | ------------- | ----------------------- |
| 1     | 30 s     | 50            | Initial ramp-up         |
| 2     | 60 s     | 50            | Steady state at 50 VUs  |
| 3     | 30 s     | 100           | Ramp-up to 100 VUs      |
| 4     | 60 s     | 100           | Steady state at 100 VUs |
| 5     | 30 s     | 200           | Ramp-up to 200 VUs      |
| 6     | 60 s     | 200           | Steady state at 200 VUs |
| 7     | 30 s     | 0             | Graceful shutdown       |

Thresholds: P95 latency < 20ms, Check success rate > 99%

### Performance Comparison

| Metrics          | HTTP/JSON | gRPC/Protobuf |
| ---------------- | --------- | ------------- |
| Throughput (QPS) | 849       | 3,139         |
| P95 Latency      | 1,000 ms  | 74.42 ms      |
| P90 Latency      | 474.7 ms  | 63.69 ms      |
| Avg Latency      | 123.85 ms | 33.25 ms      |
| Max Latency      | 2.17 s    | 1.12 s        |
| Error Rate       | 0.00 %    | 0.00 %        |

### Analysis & Conclusions

#### Why gRPC Outperforms HTTP

1. **Binary Protocol Efficiency:** Protobuf’s compact binary format reduces data size by 30-50% compared to JSON, 
    minimizing serialization/deserialization CPU overhead.
2. **HTTP/2 Multiplexing:** Single TCP connection handles concurrent requests, eliminating connection establishment 
    overhead and reducing  network latency.
3. **Header Compression**: HPACK algorithm significantly reduces metadata transmission, particularly beneficial for 
    small request/response payloads like authorization checks.

#### Tail Latency Explanation

Despite business logic taking only ~4ms (per OTel), P95 latency reached 74ms. This stems from:

- **CPU Scheduling Contention:** 200 VUs competing for 2 CPU cores causes Go runtime scheduling delays
- **Network Stack Queueing:** OS TCP buffers queue packets under extreme concurrency
- **GC Pauses:** High QPS generates temporary objects, triggering occasional Go garbage collection pauses

> [!NOTE]
>
> In production environments with more CPU cores, tail latency would be substantially lower.
> This benchmark establishes relative performance gains rather than absolute production capacity.

### Appendices

#### Test Scripts

##### HTTP

```js
import http from 'k6/http';
import { check, sleep } from 'k6';

const URL = 'http://127.0.0.1:8080/api/authz/check';  // HTTP server

export const options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '30s', target: 200 },
        { duration: '1m', target: 200 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        'http_req_duration': ['p(95)<20'],
        'checks': ['rate>0.99'],
    },
};

export default () => {
    // Test use case
    const payload = JSON.stringify({
        "object": "system:community",
        "relation": "can_publish_post",
        "user": "user:32803504934359040"
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const response = http.post(URL, payload, params);

    check(response, {
        'status is OK': (r) => r.status === 200,
    });
}
```
##### gRPC

```js
import grpc from 'k6/net/grpc';
import { check, sleep } from 'k6';

const client = new grpc.Client();
client.load(null, 'authz.proto');

let isConnected = false;

export const options = {
    stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 100 },
        { duration: '1m', target: 100 },
        { duration: '30s', target: 200 },
        { duration: '1m', target: 200 },
        { duration: '30s', target: 0 },
    ],
    thresholds: {
        'grpc_req_duration': ['p(95)<20'],
        'checks': ['rate>0.99'],
    },
};

export default () => {
    if (!isConnected) {
        client.connect('127.0.0.1:50051', { plaintext: true }); // gRPC server
        isConnected = true;
    }

    // Test use case
    const data = {
        "object": "system:community",
        "relation": "can_publish_post",
        "user": "user:32803504934359040"
    }

    const response = client.invoke('/authz.v1.AuthzService/Check', data)

    check(response, {
        'status is OK': (r) => r && r.status === grpc.StatusOK,
    });
}

export function teardown(data) {
    client.close();
}

```

#### Raw Test Results

##### HTTP

```bash
     execution: local
        script: bench_http.js
        output: -

     scenarios: (100.00%) 1 scenario, 200 max VUs, 5m30s max duration (incl. graceful stop):
              * default: Up to 200 looping VUs for 5m0s over 7 stages (gracefulRampDown: 30s, gracefulStop: 30s)



  █ THRESHOLDS 

    checks
    ✓ 'rate>0.99' rate=100.00%

    http_req_duration
    ✗ 'p(95)<20' p(95)=1s


  █ TOTAL RESULTS 

    checks_total.......: 254733  849.061246/s
    checks_succeeded...: 100.00% 254733 out of 254733
    checks_failed......: 0.00%   0 out of 254733

    ✓ status is OK

    HTTP
    http_req_duration..............: avg=123.85ms min=1.27ms med=2.72ms max=2.17s p(90)=474.7ms  p(95)=1s
      { expected_response:true }...: avg=123.85ms min=1.27ms med=2.72ms max=2.17s p(90)=474.7ms  p(95)=1s
    http_req_failed................: 0.00%  0 out of 254733
    http_reqs......................: 254733 849.061246/s

    EXECUTION
    iteration_duration.............: avg=123.98ms min=1.35ms med=2.84ms max=2.17s p(90)=474.85ms p(95)=1s
    iterations.....................: 254733 849.061246/s
    vus............................: 1      min=1           max=200
    vus_max........................: 200    min=200         max=200

    NETWORK
    data_received..................: 44 MB  148 kB/s
    data_sent......................: 58 MB  194 kB/s




running (5m00.0s), 000/200 VUs, 254733 complete and 0 interrupted iterations
default ✓ [======================================] 000/200 VUs  5m0s
ERRO[0300] thresholds on metrics 'http_req_duration' have been crossed
```

#### gRPC

```bash
     execution: local
        script: bench_grpc.js
        output: -

     scenarios: (100.00%) 1 scenario, 200 max VUs, 5m30s max duration (incl. graceful stop):
              * default: Up to 200 looping VUs for 5m0s over 7 stages (gracefulRampDown: 30s, gracefulStop: 30s)



  █ THRESHOLDS 

    checks
    ✓ 'rate>0.99' rate=100.00%

    grpc_req_duration
    ✗ 'p(95)<20' p(95)=74.42ms


  █ TOTAL RESULTS 

    checks_total.......: 942076  3139.300632/s
    checks_succeeded...: 100.00% 942076 out of 942076
    checks_failed......: 0.00%   0 out of 942076

    ✓ status is OK

    EXECUTION
    iteration_duration...: avg=33.4ms  min=1.21ms med=27.11ms max=1.12s p(90)=63.86ms p(95)=74.63ms
    iterations...........: 942076 3139.300632/s
    vus..................: 1      min=1         max=200
    vus_max..............: 200    min=200       max=200

    NETWORK
    data_received........: 80 MB  267 kB/s
    data_sent............: 134 MB 446 kB/s

    GRPC
    grpc_req_duration....: avg=33.25ms min=1.16ms med=26.97ms max=1.12s p(90)=63.69ms p(95)=74.42ms




running (5m00.1s), 000/200 VUs, 942076 complete and 0 interrupted iterations
default ✓ [======================================] 000/200 VUs  5m0s
ERRO[0300] thresholds on metrics 'grpc_req_duration' have been crossed
```
