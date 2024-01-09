#!/usr/bin/env bash

set -euxo pipefail

gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-b zanzigo-europe-west1 -- 'sudo apt-get update && sudo apt-get install -y golang-1.21 git make && git clone https://github.com/trevex/zanzigo && GO=/usr/lib/go-1.21/bin/go make build'
