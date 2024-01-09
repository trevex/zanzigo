# Benchmark

Ok, with this benchmark I wanted to see how SQLite fares against Postgres for this use-case.
The assumption is that SQLite might have some significant benefits due to data locality.

The benchmark scenario is currently hardcoded to my environment, but all the scripts and code used is here to reproduce the results, if required.

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
