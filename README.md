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

To traverse the authorization-model and check a permission, you need a resolver,
such as [`SequentialResolver`](#resolver):

```go
resolver, err := zanzigo.NewSequentialResolver(model, storage, 16)
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

## Resolver

### Sequential

The `SequentialResolver` traverse the tree each depth at a time as long as the storage-implementation supports it.
The name might be slightly misleading as no parallelism is required with current storage-implementations.
As the storage-implementation takes care of concurrency, while the resolver collects all results of a given depth.

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
