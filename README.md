[![Go Reference](https://pkg.go.dev/badge/github.com/trevex/zanzigo.svg)](https://pkg.go.dev/github.com/trevex/zanzigo)

# `zanzigo`

The `zanzigo`-library provides building blocks for creating your own [Zanzibar](https://research.google/pubs/pub48190/)-esque authorization service.
If you are unfamiliar with Google's Zanzibar, check out [zanzibar.academy](https://zanzibar.academy/) by Auth0.

## Install

```bash
go get -u github.com/trevex/zanzigo
```

## Getting started

First you will need an authorization-model, which defines the ruleset of your relation-based access control.
The structure of `ObjectMap` is inspired by [warrant](https://docs.warrant.dev/concepts/object-types/).

```go
model, err := zanzigo.NewModel(zanzigo.ObjectMap{
    "user": zanzigo.RelationMap{},
	"group": zanzigo.RelationMap{
		"member": zanzigo.Rule{},
	},
	"folder": zanzigo.RelationMap{
		"owner": zanzigo.Rule{},
		"editor": zanzigo.Rule{
			InheritIf: "owner",
		},
		"viewer": zanzigo.Rule{
			InheritIf: "editor",
		},
	},
	"doc": zanzigo.RelationMap{
		"parent": zanzigo.Rule{},
		"owner": zanzigo.Rule{
			InheritIf:    "owner",
			OfType:       "folder",
			WithRelation: "parent",
		},
		"editor": zanzigo.AnyOf(
			zanzigo.Rule{InheritIf: "owner"},
			zanzigo.Rule{
				InheritIf:    "editor",
				OfType:       "folder",
				WithRelation: "parent",
			},
		),
		"viewer": zanzigo.AnyOf(
			zanzigo.Rule{InheritIf: "editor"},
			zanzigo.Rule{
				InheritIf:    "viewer",
				OfType:       "folder",
				WithRelation: "parent",
			},
		),
	},
})
```

Next, you will need a storage-implementation, check out the [Storage](#storage)-section of this document for details.
For simplicity, let's use Postgres and assume `databaseURL` is defined:

```go
if err := postgres.RunMigrations(databaseURL); err != nil {
    // ...
}

storage, err := postgres.NewPostgresStorage(databaseURL)
if err != nil {
    // ...
}
```

To traverse the authorization-model and check a permission, you need a resolver:

```go
resolver, err := zanzigo.NewResolver(model, storage, 16)
if err != nil {
    // ...
}

// Alternatively construct zanzigo.Tuple directly instead of using the string-format from the paper.
result, err := resolver.Check(context.Background(), zanzigo.TupleString("doc:mydoc#viewer@user:myuser"))
```

That is it!

For more thorough examples, check out the `examples/`-folder in the repository.
Details regarding the storage- and resolver-implementation can be found below or in the [generated documentation](https://pkg.go.dev/github.com/trevex/zanzigo).


## Storage

### Postgres

Make sure the database migrations ran before creating the storage-backend:
```go
if err := postgres.RunMigrations(databaseURL); err != nil {
    log.Fatalf("Could not migrate db: %s", err)
}
```

The Postgres implementation comes in two flavors. One is using queries:

```go
storage, err := postgres.NewPostgresStorage(databaseURL)
```

The queries are prepared and executed at the same time using `UNION ALL`, so no parallelism of the resolver is required.
The database will traverse all checks of a certain depth at the same time for us.

The second flavor is using stored Postgres-functions:

```go
storage, err := postgres.NewPostgresStorage(databaseURL, postgres.UseFunctions())
```

This storage-implementation prepares Postgres-functions, which will traverse the authorization-model.
This means only a single query is issues calling a particular function and directly return the result of the check.

Both flavors have advantages and disadvantages, but are compatible, so swapping is possible at any time.

### SQLite3

Alternatively SQLite3 can be used as follow:
```go
dbfile := "./sqlite.db" # URL parameters from mattn/go-sqlite3 can be used
if err := sqlite3.RunMigrations(dbfile); err != nil {
    log.Fatalf("Could not migrate db: %s", err)
}
storage, err := sqlite3.NewSQLiteStorage(dbfile)
```

### Which storage implementation to use?

This really depends on which underlying database will fulfill your needs, so familiarize yourself with their trade-offs using the upstream documentation.

You might also want to consider the following on day 2:
1. You can use [Litestream](https://github.com/benbjohnson/litestream) or [LiteFS](https://github.com/superfly/litefs) to scale beyond a single replica, e.g. multiple read-replicas of SQLite3.
2. Both function and query-based flavors of the Postgres implementation should work with [Neon](https://github.com/neondatabase/neon), while only query-based approach is expected to be compatible with [CockroachDB](https://github.com/cockroachdb/cockroach).

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
