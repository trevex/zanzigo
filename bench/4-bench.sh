#!/usr/bin/env bash

set -euxo pipefail


REGION=${1-europe-west1}

time gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone ${REGION}-c bench-${REGION} -- '
set -euxo pipefail
cat <<EOT > ~/http100.js
import http from "k6/http";
import { sleep, check } from "k6";
export const options = {
  vus: 100,
  duration: "60s",
};
export default function () {
  const i = Math.floor(Math.random() * 10000)
  const url = "http://10.0.0.5:4000/zanzigo.v1.ZanzigoService/Check";
  const payload = JSON.stringify({ tuple: {
    object_type: "doc",
    object_id: "mydoc" + i,
    object_relation: "viewer",
    subject_type: "user",
    subject_id: "myuser" + i
  }});

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(url, payload, params);
  check(res, {
    "is status 200": (r) => r.status === 200,
  });

  sleep(1);
}
EOT
k6 run ~/http100.js
'

time gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone ${REGION}-c bench-${REGION} -- '
set -euxo pipefail
cat <<EOT > ~/http1000.js
import http from "k6/http";
import { sleep, check } from "k6";
export const options = {
  vus: 1000,
  duration: "60s",
};
export default function () {
  const i = Math.floor(Math.random() * 10000)
  const url = "http://10.0.0.5:4000/zanzigo.v1.ZanzigoService/Check";
  const payload = JSON.stringify({ tuple: {
    object_type: "doc",
    object_id: "mydoc" + i,
    object_relation: "viewer",
    subject_type: "user",
    subject_id: "myuser" + i
  }});

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(url, payload, params);
  check(res, {
    "is status 200": (r) => r.status === 200,
  });

  sleep(1);
}
EOT
k6 run ~/http1000.js
'

time gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone ${REGION}-c bench-${REGION} -- '
set -euxo pipefail
cat <<EOT > ~/grpc100.js
import http from "k6/http";
import grpc from "k6/net/grpc";
import { sleep, check } from "k6";

const client = new grpc.Client();
client.load([], "zanzigo.proto");

export const options = {
  vus: 100,
  duration: "60s",
};
export default function () {
  client.connect("10.0.0.5:4000", {
    plaintext: true
  });

  const i = Math.floor(Math.random() * 10000)
  const payload = { tuple: {
    object_type: "doc",
    object_id: "mydoc" + i,
    object_relation: "viewer",
    subject_type: "user",
    subject_id: "myuser" + i,
    subject_relation: ""
  }};

  const res = client.invoke("zanzigo.v1.ZanzigoService/Check", payload);

  check(res, {
    "status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  client.close();
  sleep(1);
}
EOT
k6 run ~/grpc100.js
'

time gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone ${REGION}-c bench-${REGION} -- '
set -euxo pipefail
cat <<EOT > ~/grpc100reuse.js
import http from "k6/http";
import grpc from "k6/net/grpc";
import { sleep, check } from "k6";

const client = new grpc.Client();
client.load([], "zanzigo.proto");

export const options = {
  vus: 100,
  duration: "60s",
};
export default function () {
  if (__ITER == 0) {
    client.connect("10.0.0.5:4000", {
      plaintext: true
    });
  }

  const i = Math.floor(Math.random() * 10000)
  const payload = { tuple: {
    object_type: "doc",
    object_id: "mydoc" + i,
    object_relation: "viewer",
    subject_type: "user",
    subject_id: "myuser" + i,
    subject_relation: ""
  }};

  const res = client.invoke("zanzigo.v1.ZanzigoService/Check", payload);

  check(res, {
    "status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  sleep(1);
}
EOT
k6 run ~/grpc100reuse.js
'

time gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone ${REGION}-c bench-${REGION} -- '
set -euxo pipefail
cat <<EOT > ~/grpc1000reuse.js
import http from "k6/http";
import grpc from "k6/net/grpc";
import { sleep, check } from "k6";

const client = new grpc.Client();
client.load([], "zanzigo.proto");

export const options = {
  vus: 1000,
  duration: "60s",
};
export default function () {
  if (__ITER == 0) {
    client.connect("10.0.0.5:4000", {
      plaintext: true
    });
  }

  const i = Math.floor(Math.random() * 10000)
  const payload = { tuple: {
    object_type: "doc",
    object_id: "mydoc" + i,
    object_relation: "viewer",
    subject_type: "user",
    subject_id: "myuser" + i,
    subject_relation: ""
  }};

  const res = client.invoke("zanzigo.v1.ZanzigoService/Check", payload);

  check(res, {
    "status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  sleep(1);
}
EOT
k6 run ~/grpc1000reuse.js
'
