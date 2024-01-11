#!/usr/bin/env bash

set -euxo pipefail

FLAVOR=${1-sqlite}

if [ "$FLAVOR" = "sqlite" ]; then
  gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-b zanzigo-europe-west1 -- 'cd zanzigo/ && ./build/zanzigo server examples/model.json'
elif [ "$FLAVOR" = "pgfunc" ]; then
  gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-b zanzigo-europe-west1 -- 'cd zanzigo/ && ./build/zanzigo server --postgres-url="postgres://zanzigo:zanzigo@10.196.0.3:5432/zanzigo" --use-functions examples/model.json'
elif [ "$FLAVOR" = "pgquery" ]; then
  gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-b zanzigo-europe-west1 -- 'cd zanzigo/ && ./build/zanzigo server --postgres-url="postgres://zanzigo:zanzigo@10.196.0.3:5432/zanzigo" examples/model.json'
fi
