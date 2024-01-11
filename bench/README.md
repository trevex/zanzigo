# Benchmark

The goal of this benchmark was to estimate the cost of having the authorization logic a network hop away or potentially in a different region.
A secondary objective was to see how the different storage-implementations fare against each other.

The benchmark was run once for each storage configuration with 100 VUs and 1000 VUs for 10 minutes both using gRPC and HTTP and once more using SQLite to test cross-region latency cost.

Google Cloud was used to run the benchmark and the exact zones used were:
* Postgres DB: `europe-west1-d`
* Zanzigo Instance: `europe-west1-b`
* Bench VM 1: `europe-west1-c`
* Bench VM 2: `europe-west4-c`

All machines used 4 vCPU and 16GB RAM (`n2-standard-4`) respectively.

The code is currently mostly hard-coded but left here to study or reproduce results, if desired.

Before the benchmark runs we ingest the following tuples with 10000 different IDs:
```
group:mygroup{ID}#member@user:myuser{ID}"))
doc:mydoc{ID}#parent@folder:myfolder{ID}"))
folder:myfolder{ID}#viewer@group:mygroup{ID}#member
```
A `myuser{ID}` has therefore an indirect `viewer` relation to document `mydoc{ID}`.

The benchmark will randomly run a Check of `doc:mydoc{ID}#viewer@user:myuser{ID}`, which requires tree traversal and represents a sub-optimal case for an authorization model with a layer of groups and folders.

## 0. Setup the infrastructure

Study and update the code and run the following commands from within `bench/`:
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
# keep running to ingest data and bench
./2-server.sh sqlite # pgfunc | pgquery
```

## 3. Ingest the test data
```bash
./3-data.sh
```

## 4. Run the benchmarks
```bash
./4-bench.sh [region] | tee -a [name].txt
```

# Results

The the scenario naming schema of the following table is `{pgquery|pgfunc|sqlite}-{http|grpc}-{vus}[-dr]`.
It indicates which storage configuration was used, which protocol and number of VUs.
If load was generated from a different region `-dr` is appended.

The following table shows request duration measurements

| Scenario | Avg | Min | Med | Max | p(90) | p(95) |
| --- | --- | --- | --- | --- | --- | --- |
| `pgquery-http-100` | 3.27ms | 971.54µs | 2.89ms | 77.45ms | 4.55ms | 5.35ms |
| `pgquery-http-1000` | 3.55ms | 567.23µs | 2.63ms | 290.59ms | 4.46ms | 5.78ms |
| `pgquery-grpc-100` | 3.57ms | 617.86µs | 3.23ms | 35.75ms | 5.29ms | 6.19ms |
| `pgquery-grpc-1000` | 4.04ms | 670.63µs | 2.8ms | 329.71ms | 5.16ms | 7.18ms | 
| `pgfunc-http-100` | 2.35ms | 711.38µs | 2.16ms | 37.79ms | 3.4ms  | 3.9ms |
| `pgfunc-http-1000` | 2.58ms | 522.61µs | 1.86ms | 174.75ms | 3.35ms | 4.46ms |
| `pgfunc-grpc-100` | 2.69ms | 837.79µs | 2.47ms | 40.5ms | 3.88ms | 4.53ms |
| `pgfunc-grpc-1000` | 2.94ms | 745.91µs | 2.07ms | 200.19ms | 3.63ms| 5ms |
| `sqlite-http-100` | 2.1ms | 358.8µs | 1.84ms | 30.93ms | 3.46ms | 4.09ms |
| `sqlite-http-1000` | 2.3ms | 305.23µs | 1.75ms | 208.33ms | 3.58ms | 4.99ms |
| `sqlite-grpc-100` | 2.02ms | 509.4µs | 1.7ms | 24.59ms | 3.4ms| 4.36ms |
| `sqlite-grpc-1000` | 2.68ms | 439.52µs | 1.81ms | 250.7ms | 3.59ms | 6.03ms |
| `sqlite-http-100-dr` | 8.3ms | 7.03ms | 8.09ms | 19.58ms | 9.43ms | 10.06ms |
| `sqlite-http-1000-dr` | 8.73ms | 7.01ms | 8.21ms | 167.71ms | 9.81ms | 10.79ms |
| `sqlite-grpc-100-dr` | 8.54ms | 7.13ms | 8.28ms | 21.85ms | 9.62ms | 10.52ms |
| `sqlite-grpc-1000-dr` | 8.92ms | 7.14ms | 8.08ms | 183.73ms | 9.56ms | 11.82ms |

# Raw Results


## `pgquery` (Postgres with Queries)

```
cat bench/pgquery.txt
+ cat
+ k6 run /home/nvoss/http100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 59800     ✗ 0
     data_received..................: 8.7 MB  15 kB/s
     data_sent......................: 18 MB   29 kB/s
     http_req_blocked...............: avg=32.93µs min=1.12µs   med=2.54µs  max=33.39ms p(90)=3.64µs  p(95)=4.35µs
     http_req_connecting............: avg=29.88µs min=0s       med=0s      max=33.3ms  p(90)=0s      p(95)=0s
     http_req_duration..............: avg=3.27ms  min=971.54µs med=2.89ms  max=77.45ms p(90)=4.55ms  p(95)=5.35ms
       { expected_response:true }...: avg=3.27ms  min=971.54µs med=2.89ms  max=77.45ms p(90)=4.55ms  p(95)=5.35ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 59800
     http_req_receiving.............: avg=34.77µs min=9.65µs   med=26.92µs max=2.79ms  p(90)=50.28µs p(95)=79.66µs
     http_req_sending...............: avg=18.86µs min=5.63µs   med=12.51µs max=8.46ms  p(90)=23.49µs p(95)=36.54µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s      max=0s      p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=3.21ms  min=857.45µs med=2.83ms  max=71.8ms  p(90)=4.49ms  p(95)=5.29ms
     http_reqs......................: 59800   99.603093/s
     iteration_duration.............: avg=1s      min=1s       med=1s      max=1.1s    p(90)=1s      p(95)=1s
     iterations.....................: 59800   99.603093/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (10m00.4s), 000/100 VUs, 59800 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/http1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 597974     ✗ 0
     data_received..................: 87 MB   145 kB/s
     data_sent......................: 176 MB  292 kB/s
     http_req_blocked...............: avg=150.47µs min=1.06µs   med=2.67µs  max=242.7ms  p(90)=3.69µs  p(95)=4.32µs
     http_req_connecting............: avg=145.97µs min=0s       med=0s      max=242.63ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=3.55ms   min=567.23µs med=2.63ms  max=290.59ms p(90)=4.46ms  p(95)=5.78ms
       { expected_response:true }...: avg=3.55ms   min=567.23µs med=2.63ms  max=290.59ms p(90)=4.46ms  p(95)=5.78ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 597974
     http_req_receiving.............: avg=35.29µs  min=9.25µs   med=28.08µs max=13.05ms  p(90)=50.67µs p(95)=77.38µs
     http_req_sending...............: avg=34.74µs  min=5.29µs   med=12.92µs max=89.39ms  p(90)=23.43µs p(95)=35.81µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=3.48ms   min=524.85µs med=2.58ms  max=242.56ms p(90)=4.41ms  p(95)=5.73ms
     http_reqs......................: 597974  994.960067/s
     iteration_duration.............: avg=1s       min=1s       med=1s      max=1.45s    p(90)=1s      p(95)=1s
     iterations.....................: 597974  994.960067/s
     vus............................: 160     min=160      max=1000
     vus_max........................: 1000    min=1000     max=1000


running (10m01.0s), 0000/1000 VUs, 597974 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 59800     ✗ 0
     data_received........: 6.4 MB  11 kB/s
     data_sent............: 6.4 MB  11 kB/s
     grpc_req_duration....: avg=3.57ms min=617.86µs med=3.23ms max=35.75ms p(90)=5.29ms p(95)=6.19ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.06s   p(90)=1s     p(95)=1s
     iterations...........: 59800   99.593338/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (10m00.4s), 000/100 VUs, 59800 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 597653     ✗ 0
     data_received........: 64 MB   106 kB/s
     data_sent............: 64 MB   107 kB/s
     grpc_req_duration....: avg=4.04ms min=670.63µs med=2.8ms max=329.71ms p(90)=5.16ms p(95)=7.18ms
     iteration_duration...: avg=1s     min=1s       med=1s    max=1.52s    p(90)=1s     p(95)=1s
     iterations...........: 597653  994.425755/s
     vus..................: 519     min=519      max=1000
     vus_max..............: 1000    min=1000     max=1000


running (10m01.0s), 0000/1000 VUs, 597653 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
```

## `pgfunc` (Postgres with Functions)

```
cat bench/pgfunc.txt
+ cat
+ k6 run /home/nvoss/http100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 59900     ✗ 0
     data_received..................: 8.7 MB  15 kB/s
     data_sent......................: 18 MB   29 kB/s
     http_req_blocked...............: avg=6.42µs  min=1.04µs   med=2.53µs  max=24.25ms p(90)=3.43µs  p(95)=4.05µs
     http_req_connecting............: avg=3.52µs  min=0s       med=0s      max=24.21ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=2.35ms  min=711.38µs med=2.16ms  max=37.79ms p(90)=3.4ms   p(95)=3.9ms
       { expected_response:true }...: avg=2.35ms  min=711.38µs med=2.16ms  max=37.79ms p(90)=3.4ms   p(95)=3.9ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 59900
     http_req_receiving.............: avg=37.69µs min=9.28µs   med=26.3µs  max=4.14ms  p(90)=58.82µs p(95)=105.48µs
     http_req_sending...............: avg=19.07µs min=5.6µs    med=12.33µs max=2.65ms  p(90)=22.85µs p(95)=38.8µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s      max=0s      p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=2.3ms   min=658.6µs  med=2.1ms   max=37.74ms p(90)=3.34ms  p(95)=3.85ms
     http_reqs......................: 59900   99.724269/s
     iteration_duration.............: avg=1s      min=1s       med=1s      max=1.03s   p(90)=1s      p(95)=1s
     iterations.....................: 59900   99.724269/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (10m00.7s), 000/100 VUs, 59900 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/http1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 598633     ✗ 0
     data_received..................: 87 MB   145 kB/s
     data_sent......................: 176 MB  293 kB/s
     http_req_blocked...............: avg=114.31µs min=1.02µs   med=2.54µs  max=178.66ms p(90)=3.34µs  p(95)=3.86µs
     http_req_connecting............: avg=109.94µs min=0s       med=0s      max=178.58ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=2.58ms   min=522.61µs med=1.86ms  max=174.75ms p(90)=3.35ms  p(95)=4.46ms
       { expected_response:true }...: avg=2.58ms   min=522.61µs med=1.86ms  max=174.75ms p(90)=3.35ms  p(95)=4.46ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 598633
     http_req_receiving.............: avg=37.15µs  min=9.12µs   med=26.59µs max=5.92ms   p(90)=57.89µs p(95)=97.51µs
     http_req_sending...............: avg=28.79µs  min=5.1µs    med=12.43µs max=17.44ms  p(90)=21.94µs p(95)=35.97µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=2.52ms   min=482.95µs med=1.8ms   max=174.19ms p(90)=3.29ms  p(95)=4.39ms
     http_reqs......................: 598633  996.056941/s
     iteration_duration.............: avg=1s       min=1s       med=1s      max=1.33s    p(90)=1s      p(95)=1s
     iterations.....................: 598633  996.056941/s
     vus............................: 247     min=247      max=1000
     vus_max........................: 1000    min=1000     max=1000


running (10m01.0s), 0000/1000 VUs, 598633 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 59900     ✗ 0
     data_received........: 6.4 MB  11 kB/s
     data_sent............: 6.4 MB  11 kB/s
     grpc_req_duration....: avg=2.69ms min=837.79µs med=2.47ms max=40.5ms p(90)=3.88ms p(95)=4.53ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.05s  p(90)=1s     p(95)=1s
     iterations...........: 59900   99.692536/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (10m00.8s), 000/100 VUs, 59900 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 598008     ✗ 0
     data_received........: 64 MB   106 kB/s
     data_sent............: 64 MB   107 kB/s
     grpc_req_duration....: avg=2.94ms min=745.91µs med=2.07ms max=200.19ms p(90)=3.63ms p(95)=5ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.44s    p(90)=1s     p(95)=1s
     iterations...........: 598008  995.016973/s
     vus..................: 8       min=8        max=1000
     vus_max..............: 1000    min=1000     max=1000


running (10m01.0s), 0000/1000 VUs, 598008 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
```

## `sqlite`

```
cat sqlite.txt
+ cat
+ k6 run /home/nvoss/http100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 59900     ✗ 0
     data_received..................: 8.7 MB  15 kB/s
     data_sent......................: 18 MB   29 kB/s
     http_req_blocked...............: avg=5.38µs  min=1µs      med=2.26µs  max=19.11ms p(90)=3.5µs    p(95)=4.29µs
     http_req_connecting............: avg=2.63µs  min=0s       med=0s      max=19.07ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=2.1ms   min=358.8µs  med=1.84ms  max=30.93ms p(90)=3.46ms   p(95)=4.09ms
       { expected_response:true }...: avg=2.1ms   min=358.8µs  med=1.84ms  max=30.93ms p(90)=3.46ms   p(95)=4.09ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 59900
     http_req_receiving.............: avg=37.22µs min=9.28µs   med=21.99µs max=7.17ms  p(90)=51.87µs  p(95)=127.5µs
     http_req_sending...............: avg=41.08µs min=4.99µs   med=11.69µs max=5.3ms   p(90)=111.41µs p(95)=188.96µs
     http_req_tls_handshaking.......: avg=0s      min=0s       med=0s      max=0s      p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=2.02ms  min=327.73µs med=1.77ms  max=30.3ms  p(90)=3.36ms   p(95)=3.98ms
     http_reqs......................: 59900   99.739335/s
     iteration_duration.............: avg=1s      min=1s       med=1s      max=1.03s   p(90)=1s       p(95)=1s
     iterations.....................: 59900   99.739335/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (10m00.6s), 000/100 VUs, 59900 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/http1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 598933     ✗ 0
     data_received..................: 87 MB   146 kB/s
     data_sent......................: 176 MB  293 kB/s
     http_req_blocked...............: avg=121.36µs min=1.04µs   med=2.36µs  max=194.68ms p(90)=3.4µs    p(95)=4.05µs
     http_req_connecting............: avg=116.99µs min=0s       med=0s      max=194.63ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=2.3ms    min=305.23µs med=1.75ms  max=208.33ms p(90)=3.58ms   p(95)=4.99ms
       { expected_response:true }...: avg=2.3ms    min=305.23µs med=1.75ms  max=208.33ms p(90)=3.58ms   p(95)=4.99ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 598933
     http_req_receiving.............: avg=38.1µs   min=8.65µs   med=22.29µs max=28.81ms  p(90)=49.88µs  p(95)=122.55µs
     http_req_sending...............: avg=63.23µs  min=5.14µs   med=11.88µs max=38.89ms  p(90)=117.32µs p(95)=196.23µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s      max=0s       p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=2.2ms    min=276.47µs med=1.68ms  max=208.25ms p(90)=3.47ms   p(95)=4.84ms
     http_reqs......................: 598933  996.556681/s
     iteration_duration.............: avg=1s       min=1s       med=1s      max=1.27s    p(90)=1s       p(95)=1s
     iterations.....................: 598933  996.556681/s
     vus............................: 381     min=381      max=1000
     vus_max........................: 1000    min=1000     max=1000


running (10m01.0s), 0000/1000 VUs, 598933 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 59900     ✗ 0
     data_received........: 6.4 MB  11 kB/s
     data_sent............: 6.5 MB  11 kB/s
     grpc_req_duration....: avg=2.02ms min=509.4µs med=1.7ms max=24.59ms p(90)=3.4ms p(95)=4.36ms
     iteration_duration...: avg=1s     min=1s      med=1s    max=1.04s   p(90)=1s    p(95)=1s
     iterations...........: 59900   99.740332/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (10m00.6s), 000/100 VUs, 59900 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 598322     ✗ 0
     data_received........: 64 MB   106 kB/s
     data_sent............: 64 MB   107 kB/s
     grpc_req_duration....: avg=2.68ms min=439.52µs med=1.81ms max=250.7ms p(90)=3.59ms p(95)=6.03ms
     iteration_duration...: avg=1s     min=1s       med=1s     max=1.48s   p(90)=1s     p(95)=1s
     iterations...........: 598322  995.540561/s
     vus..................: 322     min=322      max=1000
     vus_max..............: 1000    min=1000     max=1000


running (10m01.0s), 0000/1000 VUs, 598322 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
```

## `sqlite-dr` (different region)

```
cat sqlite-dr.txt
+ cat
+ k6 run /home/nvoss/http100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 59500     ✗ 0
     data_received..................: 8.7 MB  15 kB/s
     data_sent......................: 18 MB   29 kB/s
     http_req_blocked...............: avg=32.54µs min=1.05µs med=2.21µs  max=30.42ms p(90)=3.34µs  p(95)=4.06µs
     http_req_connecting............: avg=29.77µs min=0s     med=0s      max=30.39ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=8.3ms   min=7.03ms med=8.09ms  max=19.58ms p(90)=9.43ms  p(95)=10.06ms
       { expected_response:true }...: avg=8.3ms   min=7.03ms med=8.09ms  max=19.58ms p(90)=9.43ms  p(95)=10.06ms
     http_req_failed................: 0.00%   ✓ 0         ✗ 59500
     http_req_receiving.............: avg=31.49µs min=9.1µs  med=22.26µs max=8.59ms  p(90)=41.71µs p(95)=59.38µs
     http_req_sending...............: avg=31.27µs min=5.38µs med=10.98µs max=4.43ms  p(90)=54.26µs p(95)=125.73µs
     http_req_tls_handshaking.......: avg=0s      min=0s     med=0s      max=0s      p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=8.24ms  min=6.99ms med=8.03ms  max=16.49ms p(90)=9.35ms  p(95)=9.96ms
     http_reqs......................: 59500   99.114668/s
     iteration_duration.............: avg=1s      min=1s     med=1s      max=1.04s   p(90)=1.01s   p(95)=1.01s
     iterations.....................: 59500   99.114668/s
     vus............................: 100     min=100     max=100
     vus_max........................: 100     min=100     max=100


running (10m00.3s), 000/100 VUs, 59500 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/http1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/http1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ is status 200

     checks.........................: 100.00% ✓ 595000     ✗ 0
     data_received..................: 87 MB   145 kB/s
     data_sent......................: 175 MB  291 kB/s
     http_req_blocked...............: avg=100.77µs min=1µs    med=2.28µs  max=170.45ms p(90)=3.15µs  p(95)=3.64µs
     http_req_connecting............: avg=97.14µs  min=0s     med=0s      max=170.37ms p(90)=0s      p(95)=0s
     http_req_duration..............: avg=8.73ms   min=7.01ms med=8.21ms  max=167.71ms p(90)=9.81ms  p(95)=10.79ms
       { expected_response:true }...: avg=8.73ms   min=7.01ms med=8.21ms  max=167.71ms p(90)=9.81ms  p(95)=10.79ms
     http_req_failed................: 0.00%   ✓ 0          ✗ 595000
     http_req_receiving.............: avg=39.59µs  min=8.82µs med=22.73µs max=26.11ms  p(90)=46.16µs p(95)=110.64µs
     http_req_sending...............: avg=49.6µs   min=4.88µs med=11.36µs max=21.48ms  p(90)=100.6µs p(95)=173.3µs
     http_req_tls_handshaking.......: avg=0s       min=0s     med=0s      max=0s       p(90)=0s      p(95)=0s
     http_req_waiting...............: avg=8.64ms   min=6.93ms med=8.14ms  max=163.8ms  p(90)=9.7ms   p(95)=10.62ms
     http_reqs......................: 595000  990.311373/s
     iteration_duration.............: avg=1s       min=1s     med=1s      max=1.33s    p(90)=1.01s   p(95)=1.01s
     iterations.....................: 595000  990.311373/s
     vus............................: 1000    min=1000     max=1000
     vus_max........................: 1000    min=1000     max=1000


running (10m00.8s), 0000/1000 VUs, 595000 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc100.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc100.js
     output: -

  scenarios: (100.00%) 1 scenario, 100 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 100 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 59500     ✗ 0
     data_received........: 6.3 MB  11 kB/s
     data_sent............: 6.4 MB  11 kB/s
     grpc_req_duration....: avg=8.54ms min=7.13ms med=8.28ms max=21.85ms p(90)=9.62ms p(95)=10.52ms
     iteration_duration...: avg=1s     min=1s     med=1s     max=1.06s   p(90)=1.01s  p(95)=1.01s
     iterations...........: 59500   99.103304/s
     vus..................: 100     min=100     max=100
     vus_max..............: 100     min=100     max=100


running (10m00.4s), 000/100 VUs, 59500 complete and 0 interrupted iterations
default ✓ [======================================] 100 VUs  10m0s
+ cat
+ k6 run /home/nvoss/grpc1000.js

          /\      |‾‾| /‾‾/   /‾‾/
     /\  /  \     |  |/  /   /  /
    /  \/    \    |     (   /   ‾‾\
   /          \   |  |\  \ |  (‾)  |
  / __________ \  |__| \__\ \_____/ .io

  execution: local
     script: /home/nvoss/grpc1000.js
     output: -

  scenarios: (100.00%) 1 scenario, 1000 max VUs, 10m30s max duration (incl. graceful stop):
           * default: 1000 looping VUs for 10m0s (gracefulStop: 30s)


     ✓ status is OK

     checks...............: 100.00% ✓ 595000     ✗ 0
     data_received........: 63 MB   105 kB/s
     data_sent............: 64 MB   106 kB/s
     grpc_req_duration....: avg=8.92ms min=7.14ms med=8.08ms max=183.73ms p(90)=9.56ms p(95)=11.82ms
     iteration_duration...: avg=1s     min=1s     med=1s     max=1.39s    p(90)=1s     p(95)=1.01s
     iterations...........: 595000  990.148856/s
     vus..................: 883     min=883      max=1000
     vus_max..............: 1000    min=1000     max=1000


running (10m00.9s), 0000/1000 VUs, 595000 complete and 0 interrupted iterations
default ✓ [======================================] 1000 VUs  10m0s
```
