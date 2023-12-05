package main

import (
	"context"
	"log"
	"os"

	"github.com/trevex/zanzigo"
	"github.com/trevex/zanzigo/storage/postgres"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")

	// Let's make sure to run migrations
	err := postgres.RunMigrations(databaseURL)
	if err != nil {
		log.Fatalln(err)
	}
	// And create our storage backend
	storage, err := postgres.NewPostgresStorage(databaseURL, postgres.UseFunctions())
	if err != nil {
		log.Fatalln(err)
	}
	defer storage.Close()

	// Our authorization model:
	// - Users can belong to groups.
	// - Documents can be nested into folders.
	// - Permissions are inherited accordingly...
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
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	// We add user 'myuser' to the group 'mygroup'
	err = storage.Write(ctx, zanzigo.TupleString("group:mygroup#member@user:myuser"))
	if err != nil {
		log.Fatalln(err)
	}
	// The document 'mydoc' is in folder 'myfolder'
	err = storage.Write(ctx, zanzigo.TupleString("doc:mydoc#parent@folder:myfolder"))
	if err != nil {
		log.Fatalln(err)
	}
	// Members of group 'mygroup' are viewers of folder 'myfolder'
	err = storage.Write(ctx, zanzigo.TupleString("folder:myfolder#viewer@group:mygroup#member"))
	if err != nil {
		log.Fatalln(err)
	}

	// Let's create the resolver and check some permissions
	resolver, err := zanzigo.NewResolver(model, storage, 16)
	if err != nil {
		log.Fatalln(err)
	}
	// Based on the indirect permission through the group's permissions on the folder,
	// the following would return 'true':
	result, err := resolver.Check(context.Background(), zanzigo.TupleString("doc:mydoc#viewer@user:myuser"))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("The user 'myuser' is viewer of doc 'mydoc': %v", result)

	// The following should be 'false':
	result, err = resolver.Check(context.Background(), zanzigo.TupleString("doc:mydoc#editor@user:myuser"))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("The user 'myuser' is editor of doc 'mydoc': %v", result)
}
