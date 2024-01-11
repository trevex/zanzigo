# Benchmark

The goal of this benchmark was to estimate the cost of having the authorization logic a network hop away.
I also wanted to see how SQLite fares against Postgres for this use-case.
The assumption is that SQLite might have some significant benefits due to data locality.

The benchmark was run twice: one in the same region as `zanzigo` and a second time in a different region, but still Europe (on Google Cloud `europe-west1` and `europe-west4`).

The benchmark scenario is currently hardcoded to my environment, but all the scripts and code used is here to reproduce the results, if required.

The benchmark itself is fairly straight forward we ingest the following tuple set 10000 with different IDs:
```
group:mygroup{ID}#member@user:myuser{ID}"))
doc:mydoc{ID}#parent@folder:myfolder{ID}"))
folder:myfolder{ID}#viewer@group:mygroup{ID}#member
```
The user therefore is indirectly a `viewer` of the document, but tree traversal is required!

The benchmark will randomly run a Check of `doc:mydoc{ID}#viewer@user:myuser{ID}`.

## 0. Setup the infrastructure

Run the following commands from within `bench/`:
```bash
terraform -chdir=0-setup init
terraform -chdir=0-setup apply
```

## 1. Install zanzigo on the machine
```bash
./1-install.sh
```

## 2. Run the server
```bash
./2-server.sh sqlite # pgfunc | pgquery
```

## 3. Ingest the test data
```bash
./3-data.sh
```

## 4. Run the benchmarks
```bash
./4-bench.sh [region]
```

# TLDR Results

For a quick overview of the results check the table below.

All scenarios use `n2-standard-4` (4 vCPU 16GB RAM). If not otherwise mentioned the load was generated on a different VM in the same region, but different zone. Load generated from a different region is suffixed with `sd`.

Therefore the scenario naming schema of the table is `{pgquery|pgfunc|sqlite}-{http|grpc}-{vus}[-sd]` (gRPC without connection re-use is excluded).

| Scenario | Request Duration |
| --- | --- |
| `pgquery-http-100` | `avg=3.91ms   min=1.06ms  med=2.9ms   max=65.78ms  p(90)=4.55ms  p(95)=5.95ms` |
| `pgquery-http-1000` | `avg=9.2ms    min=728.67µs med=2.77ms  max=269.05ms p(90)=5.6ms   p(95)=13.19ms` |
| `pgquery-grpc-100` | `avg=3.94ms min=1.37ms med=3.11ms max=40.09ms p(90)=5.27ms p(95)=8.33ms` |
| `pgquery-grpc-1000` | `avg=11.13ms min=830.06µs med=2.94ms max=399.79ms p(90)=7.28ms p(95)=23.05ms` | 
| `pgfunc-http-100` | `avg=2.68ms  min=748.27µs med=2.25ms  max=36.09ms p(90)=3.45ms  p(95)=4.04ms` |
| `pgfunc-http-1000` | `avg=4.58ms   min=645.74µs med=1.92ms  max=197.74ms p(90)=5.26ms  p(95)=9.09ms` |
| `pgfunc-grpc-100` | `avg=2.71ms min=879.46µs med=2.4ms max=23.42ms p(90)=3.83ms p(95)=5.16ms` |
| `pgfunc-grpc-1000` | `avg=6.18ms min=855.59µs med=2.31ms max=220.38ms p(90)=7.34ms p(95)=19.37ms` |
| `sqlite-http-100` | `avg=2.03ms   min=372.72µs med=1.68ms  max=40.79ms p(90)=3.31ms   p(95)=3.91ms` |
| `sqlite-http-1000` | `avg=4.63ms   min=360.8µs med=1.68ms  max=126.67ms p(90)=4.32ms   p(95)=10.92ms` |
| `sqlite-grpc-100` | `avg=2.27ms min=565.57µs med=1.76ms max=37.28ms p(90)=3.66ms p(95)=5.08ms` |
| `sqlite-grpc-1000` | `avg=3.42ms min=442.94µs med=1.8ms max=275.07ms p(90)=7.58ms p(95)=13.11ms` |
| `sqlite-http-100-dr` | `avg=8.7ms    min=7.07ms med=8.41ms  max=17.25ms p(90)=9.94ms   p(95)=11.4ms` |
| `sqlite-http-1000-dr` | `avg=11.19ms  min=7.01ms med=8.17ms  max=139.6ms  p(90)=10.79ms  p(95)=16.29ms` |
| `sqlite-grpc-100-dr` | `avg=8.75ms min=7.19ms med=8.3ms max=21.31ms p(90)=10.25ms p(95)=12.2ms` |
| `sqlite-grpc-1000-dr` | `avg=11.48ms min=7.19ms med=8.27ms max=180.33ms p(90)=13.1ms p(95)=20.31ms` |

# Full Results

_NOTE_: The same machine type was used both for benchmark generation and zanzigo service

## `pgquery` (Postgres with Queries) + `n2-standard-4` (4 vCPU 16GB RAM)

### HTTP, same region, different zone, 100 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 6000      ✗ 0
     data_received..................: 876 kB  15 kB/s
     data_sent......................: 1.8 MB  29 kB/s
     http_req_blocked...............: avg=60.43µs  min=1.24µs  med=2.69µs  max=32.81ms  p(90)=4.04µs  p(95)=5.07µs
     http_req_connecting............: avg=46.43µs  min=0s      med=0s      max=30.08ms  p(90)=0s      p(95)=0s
     http_req_duration..............: avg=3.91ms   min=1.06ms  med=2.9ms   max=65.78ms  p(90)=4.55ms  p(95)=5.95ms
       { expected_response:true }...: avg=3.91ms   min=1.06ms  med=2.9ms   max=65.78ms  p(90)=4.55ms  p(95)=5.95ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 6000
     http_req_receiving.............: avg=35.69µs  min=10.23µs med=27.99µs max=763.83µs p(90)=50.53µs p(95)=77.91µs
     http_req_sending...............: avg=159.96µs min=5.95µs  med=13.16µs max=31.05ms  p(90)=26.24µs p(95)=53.6µs
     http_req_tls_handshaking.......: avg=0s       min=0s      med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=3.72ms   min=1.02ms  med=2.84ms  max=54.82ms  p(90)=4.5ms   p(95)=5.89ms
     http_reqs......................: 6000    99.467985/s
     iteration_duration.............: avg=1s       min=1s      med=1s      max=1.07s    p(90)=1s      p(95)=1s
     iterations.....................: 6000    99.467985/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (1m00.3s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### HTTP, same region, different zone, 1000 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 59851      ✗ 0
     data_received..................: 8.7 MB  143 kB/s
     data_sent......................: 18 MB   288 kB/s
     http_req_blocked...............: avg=1.13ms   min=1.17µs   med=2.9µs   max=233.49ms p(90)=4.13µs  p(95)=5.19µs
     http_req_connecting............: avg=1.11ms   min=0s       med=0s      max=233.45ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=9.2ms    min=728.67µs med=2.77ms  max=269.05ms p(90)=5.6ms   p(95)=13.19ms
       { expected_response:true }...: avg=9.2ms    min=728.67µs med=2.77ms  max=269.05ms p(90)=5.6ms   p(95)=13.19ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 59851
     http_req_receiving.............: avg=36.1µs   min=9.63µs   med=29.06µs max=1.93ms   p(90)=50.78µs p(95)=76.86µs
     http_req_sending...............: avg=106.26µs min=5.63µs   med=13.28µs max=149.76ms p(90)=26.34µs p(95)=56.03µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=9.06ms   min=685.73µs med=2.71ms  max=265.53ms p(90)=5.54ms  p(95)=13.12ms
     http_reqs......................: 59851   981.200188/s
     iteration_duration.............: avg=1.01s    min=1s       med=1s      max=1.48s    p(90)=1s      p(95)=1.01s
     iterations.....................: 59851   981.200188/s
     vus............................: 242     min=242      max=1000
     vus_max........................: 1000    min=1000     max=1000


running (1m01.0s), 0000/1000 VUs, 59851 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  1m0s
```

### gRPC, same region, different zone, 100 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 6000      ✗ 0
     data_received........: 1.3 MB  22 kB/s
     data_sent............: 1.5 MB  25 kB/s
     grpc_req_duration....: avg=4.37ms min=896.7µs med=3.72ms max=47.63ms p(90)=6.76ms p(95)=8.25ms
     iteration_duration...: avg=1s     min=1s      med=1s     max=1.05s   p(90)=1s     p(95)=1.01s
     iterations...........: 6000    99.318701/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (1m00.4s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), same region, different zone, 100 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 6000      ✗ 0
     data_received........: 643 kB  11 kB/s
     data_sent............: 689 kB  11 kB/s
     grpc_req_duration....: avg=3.94ms min=1.37ms med=3.11ms max=40.09ms p(90)=5.27ms p(95)=8.33ms
     iteration_duration...: avg=1s     min=1s     med=1s     max=1.06s   p(90)=1s     p(95)=1s
     iterations...........: 6000    99.474621/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (1m00.3s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), same region, different zone, 1000 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 59550      ✗ 0
     data_received........: 6.4 MB  105 kB/s
     data_sent............: 6.8 MB  112 kB/s
     grpc_req_duration....: avg=11.13ms min=830.06µs med=2.94ms max=399.79ms p(90)=7.28ms p(95)=23.05ms
     iteration_duration...: avg=1.01s   min=1s       med=1s     max=1.58s    p(90)=1s     p(95)=1.02s
     iterations...........: 59550   976.169937/s
     vus..................: 396     min=396      max=1000
     vus_max..............: 1000    min=1000     max=1000


running (1m01.0s), 0000/1000 VUs, 59550 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  1m0s
```

## `pgfunc` (Postgres with Functions) + `n2-standard-4` (4 vCPU 16GB RAM)

### HTTP, same region, different zone, 100 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 6000      ✗ 0
     data_received..................: 876 kB  15 kB/s
     data_sent......................: 1.8 MB  29 kB/s
     http_req_blocked...............: avg=74.35µs min=1.2µs    med=2.74µs  max=21.38ms p(90)=3.92µs  p(95)=5.26µs
     http_req_connecting............: avg=69.78µs min=0s       med=0s      max=21.23ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=2.68ms  min=748.27µs med=2.25ms  max=36.09ms p(90)=3.45ms  p(95)=4.04ms
       { expected_response:true }...: avg=2.68ms  min=748.27µs med=2.25ms  max=36.09ms p(90)=3.45ms  p(95)=4.04ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 6000
     http_req_receiving.............: avg=42.27µs min=10.64µs  med=28.01µs max=2.18ms  p(90)=64.83µs p(95)=125.37µs
     http_req_sending...............: avg=79.89µs min=5.8µs    med=13.03µs max=18.64ms p(90)=27.99µs p(95)=63.52µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s      max=0s      p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=2.56ms  min=710.64µs med=2.19ms  max=34.79ms p(90)=3.39ms  p(95)=3.99ms
     http_reqs......................: 6000    99.638843/s
     iteration_duration.............: avg=1s      min=1s       med=1s      max=1.04s   p(90)=1s      p(95)=1s
     iterations.....................: 6000    99.638843/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (1m00.2s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### HTTP, same region, different zone, 1000 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 60000     ✗ 0
     data_received..................: 8.8 MB  145 kB/s
     data_sent......................: 18 MB   291 kB/s
     http_req_blocked...............: avg=1.06ms   min=1.13µs   med=2.76µs  max=169.57ms p(90)=3.67µs  p(95)=4.71µs
     http_req_connecting............: avg=1.04ms   min=0s       med=0s      max=169.53ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=4.58ms   min=645.74µs med=1.92ms  max=197.74ms p(90)=5.26ms  p(95)=9.09ms
       { expected_response:true }...: avg=4.58ms   min=645.74µs med=1.92ms  max=197.74ms p(90)=5.26ms  p(95)=9.09ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 60000
     http_req_receiving.............: avg=41.78µs  min=10.43µs  med=28.11µs max=4.71ms   p(90)=65.83µs p(95)=116.63µs
     http_req_sending...............: avg=102.47µs min=5.77µs   med=12.97µs max=24.71ms  p(90)=26.27µs p(95)=59.62µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=4.44ms   min=601.92µs med=1.85ms  max=184.13ms p(90)=5.19ms  p(95)=9.02ms
     http_reqs......................: 60000   990.14732/s
     iteration_duration.............: avg=1s       min=1s       med=1s      max=1.31s    p(90)=1s      p(95)=1s
     iterations.....................: 60000   990.14732/s
     vus............................: 1000    min=1000    max=1000
     vus_max........................: 1000    min=1000    max=1000


running (1m00.6s), 0000/1000 VUs, 60000 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  1m0s
```

### gRPC, same region, different zone, 100 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 6000      ✗ 0
     data_received........: 1.3 MB  22 kB/s
     data_sent............: 1.5 MB  25 kB/s
     grpc_req_duration....: avg=2.61ms min=925.41µs med=2.15ms max=20.95ms p(90)=4.07ms p(95)=5.41ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.04s   p(90)=1s     p(95)=1.01s
     iterations...........: 6000    99.466121/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (1m00.3s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), same region, different zone, 100 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 6000      ✗ 0
     data_received........: 643 kB  11 kB/s
     data_sent............: 688 kB  11 kB/s
     grpc_req_duration....: avg=2.71ms min=879.46µs med=2.4ms max=23.42ms p(90)=3.83ms p(95)=5.16ms
     iteration_duration...: avg=1s     min=1s       med=1s    max=1.04s   p(90)=1s     p(95)=1s
     iterations...........: 6000    99.629454/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (1m00.2s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), same region, different zone, 1000 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 60000      ✗ 0
     data_received........: 6.4 MB  106 kB/s
     data_sent............: 6.9 MB  113 kB/s
     grpc_req_duration....: avg=6.18ms min=855.59µs med=2.31ms max=220.38ms p(90)=7.34ms p(95)=19.37ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.45s    p(90)=1s     p(95)=1.02s
     iterations...........: 60000   986.066907/s
     vus..................: 554     min=554      max=1000
     vus_max..............: 1000    min=1000     max=1000


running (1m00.8s), 0000/1000 VUs, 60000 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  1m0s
```


## `sqlite` + `n2-standard-4` (4 vCPU 16GB RAM)

### HTTP, same region, different zone, 100 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 6000      ✗ 0
     data_received..................: 876 kB  15 kB/s
     data_sent......................: 1.8 MB  29 kB/s
     http_req_blocked...............: avg=236.17µs min=1.18µs   med=2.53µs  max=38.77ms p(90)=4.08µs   p(95)=5.2µs
     http_req_connecting............: avg=231.79µs min=0s       med=0s      max=38.74ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=2.03ms   min=372.72µs med=1.68ms  max=40.79ms p(90)=3.31ms   p(95)=3.91ms
       { expected_response:true }...: avg=2.03ms   min=372.72µs med=1.68ms  max=40.79ms p(90)=3.31ms   p(95)=3.91ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 6000
     http_req_receiving.............: avg=41.55µs  min=9.89µs   med=23.33µs max=9.34ms  p(90)=55.04µs  p(95)=138.5µs
     http_req_sending...............: avg=59.72µs  min=6µs      med=12.35µs max=3.85ms  p(90)=134.08µs p(95)=228.65µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s      p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=1.93ms   min=329.59µs med=1.6ms   max=39.9ms  p(90)=3.19ms   p(95)=3.75ms
     http_reqs......................: 6000    99.683765/s
     iteration_duration.............: avg=1s       min=1s       med=1s      max=1.04s   p(90)=1s       p(95)=1s
     iterations.....................: 6000    99.683765/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (1m00.2s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### HTTP, same region, different zone, 1000 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/admin_nvoss_altostrat_com/http1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 60000      ✗ 0
     data_received..................: 8.8 MB  145 kB/s
     data_sent......................: 18 MB   291 kB/s
     http_req_blocked...............: avg=899.49µs min=1.09µs  med=2.56µs  max=162.15ms p(90)=3.78µs   p(95)=4.91µs
     http_req_connecting............: avg=885.92µs min=0s      med=0s      max=161.95ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=4.63ms   min=360.8µs med=1.68ms  max=126.67ms p(90)=4.32ms   p(95)=10.92ms
       { expected_response:true }...: avg=4.63ms   min=360.8µs med=1.68ms  max=126.67ms p(90)=4.32ms   p(95)=10.92ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 60000
     http_req_receiving.............: avg=39.02µs  min=8.8µs   med=23.39µs max=13.46ms  p(90)=51.39µs  p(95)=129.76µs
     http_req_sending...............: avg=162.98µs min=5.24µs  med=12.55µs max=34.66ms  p(90)=149.23µs p(95)=281.87µs
     http_req_tls_handshaking.......: avg=0s       min=0s      med=0s      max=0s       p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=4.43ms   min=326µs   med=1.61ms  max=111.82ms p(90)=4.19ms   p(95)=10.26ms
     http_reqs......................: 60000   990.040655/s
     iteration_duration.............: avg=1s       min=1s      med=1s      max=1.26s    p(90)=1s       p(95)=1.01s
     iterations.....................: 60000   990.040655/s
     vus............................: 1000    min=1000     max=1000
     vus_max........................: 1000    min=1000     max=1000


running (1m00.6s), 0000/1000 VUs, 60000 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  1m0s
```

### gRPC, same region, different zone, 100 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 6000     ✗ 0
     data_received........: 1.3 MB  22 kB/s
     data_sent............: 1.5 MB  24 kB/s
     grpc_req_duration....: avg=2.18ms min=590.32µs med=1.64ms max=19.57ms p(90)=4.13ms p(95)=5.52ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.04s   p(90)=1s     p(95)=1.01s
     iterations...........: 6000    99.50221/s
     vus..................: 100     min=100    max=100
     vus_max..............: 100     min=100    max=100


running (1m00.3s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), same region, different zone, 100 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 6000      ✗ 0
     data_received........: 643 kB  11 kB/s
     data_sent............: 688 kB  11 kB/s
     grpc_req_duration....: avg=2.27ms min=565.57µs med=1.76ms max=37.28ms p(90)=3.66ms p(95)=5.08ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.05s   p(90)=1s     p(95)=1s
     iterations...........: 6000    99.643061/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (1m00.2s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), same region, different zone, 1000 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 60000     ✗ 0
     data_received........: 6.4 MB  106 kB/s
     data_sent............: 6.9 MB  114 kB/s
     grpc_req_duration....: avg=3.42ms min=442.94µs med=1.8ms max=275.07ms p(90)=7.58ms p(95)=13.11ms
     iteration_duration...: avg=1s     min=1s       med=1s    max=1.35s    p(90)=1s     p(95)=1.01s
     iterations...........: 60000   989.48392/s
     vus..................: 51      min=51      max=1000
     vus_max..............: 1000    min=1000    max=1000


running (1m00.6s), 0000/1000 VUs, 60000 complete and 0 interrupted iterations
```

### HTTP, different region, 100 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 6000      ✗ 0
     data_received..................: 876 kB  15 kB/s
     data_sent......................: 1.8 MB  29 kB/s
     http_req_blocked...............: avg=311.95µs min=1.06µs med=2.18µs  max=31.84ms p(90)=3.4µs    p(95)=4.49µs
     http_req_connecting............: avg=308.26µs min=0s     med=0s      max=31.74ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=8.7ms    min=7.07ms med=8.41ms  max=17.25ms p(90)=9.94ms   p(95)=11.4ms
       { expected_response:true }...: avg=8.7ms    min=7.07ms med=8.41ms  max=17.25ms p(90)=9.94ms   p(95)=11.4ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 6000
     http_req_receiving.............: avg=33.89µs  min=9.29µs med=22.09µs max=3.44ms  p(90)=40.41µs  p(95)=68.65µs
     http_req_sending...............: avg=75.55µs  min=5.47µs med=11.07µs max=5.94ms  p(90)=103.04µs p(95)=185.23µs
     http_req_tls_handshaking.......: avg=0s       min=0s     med=0s      max=0s      p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=8.59ms   min=7.03ms med=8.35ms  max=17.21ms p(90)=9.77ms   p(95)=10.94ms
     http_reqs......................: 6000    99.010378/s
     iteration_duration.............: avg=1s       min=1s     med=1s      max=1.04s   p(90)=1.01s    p(95)=1.01s
     iterations.....................: 6000    99.010378/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (1m00.6s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### HTTP, different region, 1000 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 60000      ✗ 0
     data_received..................: 8.8 MB  144 kB/s
     data_sent......................: 18 MB   289 kB/s
     http_req_blocked...............: avg=990.24µs min=1.1µs  med=2.42µs  max=167.34ms p(90)=3.35µs   p(95)=4.39µs
     http_req_connecting............: avg=974.07µs min=0s     med=0s      max=160.74ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=11.19ms  min=7.01ms med=8.17ms  max=139.6ms  p(90)=10.79ms  p(95)=16.29ms
       { expected_response:true }...: avg=11.19ms  min=7.01ms med=8.17ms  max=139.6ms  p(90)=10.79ms  p(95)=16.29ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 60000
     http_req_receiving.............: avg=33.4µs   min=9.3µs  med=22.98µs max=10.57ms  p(90)=42.16µs  p(95)=66.75µs
     http_req_sending...............: avg=108.07µs min=5.33µs med=11.81µs max=14.18ms  p(90)=115.59µs p(95)=265.88µs
     http_req_tls_handshaking.......: avg=0s       min=0s     med=0s      max=0s       p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=11.05ms  min=6.98ms med=8.11ms  max=127.89ms p(90)=10.68ms  p(95)=16.12ms
     http_reqs......................: 60000   983.932453/s
     iteration_duration.............: avg=1.01s    min=1s     med=1s      max=1.28s    p(90)=1.01s    p(95)=1.01s
     iterations.....................: 60000   983.932453/s
     vus............................: 448     min=448      max=1000
     vus_max........................: 1000    min=1000     max=1000


running (1m01.0s), 0000/1000 VUs, 60000 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  1m0s
```

### gRPC, different region, 100 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 5900      ✗ 0
     data_received........: 1.3 MB  21 kB/s
     data_sent............: 1.5 MB  24 kB/s
     grpc_req_duration....: avg=9.31ms min=7.16ms med=8.5ms max=22.34ms p(90)=12.23ms p(95)=14.43ms
     iteration_duration...: avg=1.02s  min=1.02s  med=1.02s max=1.05s   p(90)=1.03s   p(95)=1.03s
     iterations...........: 5900    97.354455/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (1m00.6s), 000/100 VUs, 5900 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), different region, 100 VUs

```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 6000      ✗ 0
     data_received........: 643 kB  11 kB/s
     data_sent............: 687 kB  11 kB/s
     grpc_req_duration....: avg=8.75ms min=7.19ms med=8.3ms max=21.31ms p(90)=10.25ms p(95)=12.2ms
     iteration_duration...: avg=1s     min=1s     med=1s    max=1.05s   p(90)=1.01s   p(95)=1.01s
     iterations...........: 6000    99.003884/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (1m00.6s), 000/100 VUs, 6000 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  1m0s
```

### gRPC (reusing connections), different region, 1000 VUs
```
          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 1m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 1m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 59720      ✗ 0
     data_received........: 6.4 MB  105 kB/s
     data_sent............: 6.8 MB  112 kB/s
     grpc_req_duration....: avg=11.48ms min=7.19ms med=8.27ms max=180.33ms p(90)=13.1ms p(95)=20.31ms
     iteration_duration...: avg=1.01s   min=1s     med=1s     max=1.39s    p(90)=1.01s  p(95)=1.02s
     iterations...........: 59720   978.866639/s
     vus..................: 720     min=720      max=1000
     vus_max..............: 1000    min=1000     max=1000


running (1m01.0s), 0000/1000 VUs, 59720 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  1m0s
```
