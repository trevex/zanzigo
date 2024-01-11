#!/usr/bin/env bash

set -euxo pipefail

gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-b zanzigo-europe-west1 -- 'sudo apt-get update && sudo apt-get install -y golang-1.21 git make && git clone https://github.com/trevex/zanzigo && GO=/usr/lib/go-1.21/bin/go make build'

gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-c bench-europe-west1 -- 'sudo gpg -k && sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69 && echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list && sudo apt-get update && sudo apt-get install k6'
gcloud compute scp --project nvoss-test --tunnel-through-iap --zone europe-west1-c ../api/zanzigo/v1/zanzigo.proto bench-europe-west1:~/zanzigo.proto

gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west4-c bench-europe-west4 -- 'sudo gpg -k && sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69 && echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list && sudo apt-get update && sudo apt-get install k6'
gcloud compute scp --project nvoss-test --tunnel-through-iap --zone europe-west4-c ../api/zanzigo/v1/zanzigo.proto bench-europe-west4:~/zanzigo.proto
