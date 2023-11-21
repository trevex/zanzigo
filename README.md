# `zanzigo`

## Development

### Persistent Postgres

During development it might make sense to persist data created by tests.
You can specify a different database to use, by setting `TEST_POSTGRES_DATABASE_URL` environment variable.

For example start a postgres database in docker and run tests against it as follows:
```bash
docker run --name postgres -e POSTGRES_USER=zanzigo -e POSTGRES_PASSWORD=zanzigo -e POSTGRES_DB=zanzigo -e listen_addresses='*' --net=host -d postgres:15.4
TEST_POSTGRES_DATABASE_URL="postgres://zanzigo:zanzigo@127.0.0.1:5432/zanzigo?sslmode=disable" go test -v ./...
```

If you want to inspect the database it might be helpful to run `pgAdmin4`:
```bash
docker run -d --name pgadmin  -e PGADMIN_DEFAULT_EMAIL='test@test.local' -e PGADMIN_DEFAULT_PASSWORD=secret -e PGADMIN_CONFIG_SERVER_MODE='False' -e PGADMIN_LISTEN_PORT=8080 --net=host dpage/pgadmin4
```
