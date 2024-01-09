# Benchmark

## Setup

The benchmark runs on several VMs on Google Cloud, so a project with a billing account has to be created first.

Replace the project and regions in `setup/bench.auto.tfvars` as desired. You can also setup remote state in `setup/main.tf` if desired.

Run the following commands, when in `bench/` to create the test environment:
```bash
terraform -chdir=setup init
terraform -chdir=setup apply
```

It will create a network with two subnets in two different regions. The first region is the primary region containing a Postgres database and a VM.
The second region is a region to test cross-region latency of Zanzigo.



