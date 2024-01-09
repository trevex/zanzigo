#!/usr/bin/env bash

set -euxo pipefail

time gcloud compute ssh --project nvoss-test --tunnel-through-iap --zone europe-west1-b zanzigo-europe-west1 -- '
set -euxo pipefail
COUNT=10000
for i in $(seq 1 $COUNT); do
  curl --header "Content-Type: application/json" \
      --data "{\"tuple\": {\"object_type\": \"group\", \"object_id\":\"mygroup$i\", \"object_relation\": \"member\", \"subject_type\": \"user\", \"subject_id\": \"myuser$i\"}}" \
      http://localhost:4000/zanzigo.v1.ZanzigoService/Write
  curl --header "Content-Type: application/json" \
      --data "{\"tuple\": {\"object_type\": \"doc\", \"object_id\":\"mydoc$i\", \"object_relation\": \"parent\", \"subject_type\": \"folder\", \"subject_id\": \"myfolder$i\"}}" \
      http://localhost:4000/zanzigo.v1.ZanzigoService/Write
  curl --header "Content-Type: application/json" \
      --data "{\"tuple\": {\"object_type\": \"folder\", \"object_id\":\"myfolder$i\", \"object_relation\": \"viewer\", \"subject_type\": \"group\", \"subject_id\": \"mygroup$i\", \"subject_relation\": \"member\"}}" \
      http://localhost:4000/zanzigo.v1.ZanzigoService/Write
done
'





