#!/usr/bin/env bash

set -euxo pipefail

FLAVOR=${1-sqlite}

if [ "$FLAVOR" = "sqlite" ]; then
  gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-b zanzigo-europe-west1 -- 'cd zanzigo/ && ./build/zanzigo server examples/model.json'
elif [ "$FLAVOR" = "pgfunc" ]; then
  echo ""
elif [ "$FLAVOR" = "pgquery" ]; then
  echo ""
fi
