// The zanzigo-package provides building blocks for creating your own [Zanzibar]-esque
// authorization service.
//
// Your start by defining an authorization model:
//
//	model, err := zanzigo.NewModel(zanzigo.ObjectMap{
//		"user": zanzigo.RelationMap{},
//		"group": zanzigo.RelationMap{
//			"member": zanzigo.Rule{},
//		},
//		"folder": zanzigo.RelationMap{
//			"owner": zanzigo.Rule{},
//			"editor": zanzigo.Rule{
//				InheritIf: "owner",
//			},
//			"viewer": zanzigo.Rule{
//				InheritIf: "editor",
//			},
//		},
//		"doc": zanzigo.RelationMap{
//			"parent": zanzigo.Rule{},
//			"owner": zanzigo.Rule{
//				InheritIf:    "owner",
//				OfType:       "folder",
//				WithRelation: "parent",
//			},
//			"editor": zanzigo.AnyOf(
//				zanzigo.Rule{InheritIf: "owner"},
//				zanzigo.Rule{
//					InheritIf:    "editor",
//					OfType:       "folder",
//					WithRelation: "parent",
//				},
//			),
//			"viewer": zanzigo.AnyOf(
//				zanzigo.Rule{InheritIf: "editor"},
//				zanzigo.Rule{
//					InheritIf:    "viewer",
//					OfType:       "folder",
//					WithRelation: "parent",
//				},
//			),
//		},
//	})
//
// With a storage-implementation available, tuples can be inserted (check [whitepaper] for notation or altenatively construct [Tuple] directly):
//
//	// We add user 'myuser' to the group 'mygroup'
//	 = storage.Write(ctx, zanzigo.TupleString("group:mygroup#member@user:myuser"))
//	// The document 'mydoc' is in folder 'myfolder'
//	_ = storage.Write(ctx, zanzigo.TupleString("doc:mydoc#parent@folder:myfolder"))
//	// Members of group 'mygroup' are viewers of folder 'myfolder'
//	_ = storage.Write(ctx, zanzigo.TupleString("folder:myfolder#viewer@group:mygroup#member"))
//
// Using a [Resolver] such as [SequentialResolver] permissions can be checked by traversing the tuples using the rules of the authorization-model:
//
//	resolver, _ := zanzigo.NewSequentialResolver(model, storage, 16)
//	// Based on the indirect permission through the group's permissions on the folder,
//	// the following would return 'true':
//	result, _ := resolver.Check(context.Background(), zanzigo.TupleString("doc:mydoc#viewer@user:myuser"))
//
// For more examples, check the repository.
// You may find additional information in the README.
//
// [Zanzibar]: https://research.google/pubs/pub48190/
// [whitepaper]: https://research.google/pubs/pub48190/
package zanzigo
