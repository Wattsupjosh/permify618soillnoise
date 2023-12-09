package memory

import (
	"context"
	"fmt"

	memory "github.com/Permify/permify/pkg/database/memory"

	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/hashicorp/go-memdb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BundleReader memory", func() {
	var Schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			SchemaDefinitionsTable: {
				Name: SchemaDefinitionsTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "Name"},
								&memdb.StringFieldIndex{Field: "Version"},
							},
						},
					},
					"version": {
						Name:   "version",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "Version"},
							},
						},
					},
					"tenant": {
						Name:   "tenant",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
							},
						},
					},
				},
			},
			AttributesTable: {
				Name: AttributesTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "EntityID"},
								&memdb.StringFieldIndex{Field: "Attribute"},
							},
						},
					},
					"entity-type-index": {
						Name:   "entity-type-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
							},
						},
					},
					"entity-type-and-attribute-index": {
						Name:   "entity-type-and-attribute-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "Attribute"},
							},
						},
					},
				},
			},
			RelationTuplesTable: {
				Name: RelationTuplesTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "EntityID"},
								&memdb.StringFieldIndex{Field: "Relation"},
								&memdb.StringFieldIndex{Field: "SubjectType"},
								&memdb.StringFieldIndex{Field: "SubjectID"},
								&memdb.StringFieldIndex{Field: "SubjectRelation"},
							},
							AllowMissing: true,
						},
					},
					"entity-index": {
						Name:   "entity-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "EntityID"},
								&memdb.StringFieldIndex{Field: "Relation"},
							},
						},
					},
					"relation-index": {
						Name:   "relation-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "Relation"},
								&memdb.StringFieldIndex{Field: "SubjectType"},
							},
						},
					},
					"entity-type-index": {
						Name:   "entity-type-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
							},
						},
					},
					"entity-type-and-relation-index": {
						Name:   "entity-type-and-relation-index",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "TenantID"},
								&memdb.StringFieldIndex{Field: "EntityType"},
								&memdb.StringFieldIndex{Field: "Relation"},
							},
						},
					},
				},
			},
			TenantsTable: {
				Name: TenantsTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "ID"},
							},
						},
					},
				},
			},
			BundlesTable: {
				Name: BundlesTable,
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.StringFieldIndex{
							Field: "Name",
						},
					},
				},
			},
		},
	}

	var db *memory.Memory
	var bundleWriter *BundleWriter
	var bundleReader *BundleReader

	BeforeEach(func() {
		dbx, err := memory.New(Schema)
		if err != nil {
			fmt.Printf(err.Error(), "FUCK")
			// handle error
		}
		db = dbx
		bundleWriter = NewBundleWriter(db)
		bundleReader = NewBundleReader(db)
	})

	AfterEach(func() {
		err := db.Close()
		Expect(err).ShouldNot(HaveOccurred())
	})
	Context("Read", func() {
		It("should write and read DataBundles with correct relationships and attributes", func() {
			ctx := context.Background()

			bundles := []*base.DataBundle{
				{
					Name: "user_created",
					Arguments: []string{
						"organizationID",
						"userID",
					},
					Operations: []*base.Operation{
						{
							RelationshipsWrite: []string{
								"organization:{{.organizationID}}#member@user:{{.userID}}",
								"organization:{{.organizationID}}#admin@user:{{.userID}}",
							},
							RelationshipsDelete: []string{},
							AttributesWrite: []string{
								"organization:{{.organizationID}}$public|boolean:true",
							},
							AttributesDelete: []string{
								"organization:{{.organizationID}}$balance|integer[]:120,568",
							},
						},
					},
				},
			}

			names, err := bundleWriter.Write(ctx, "t1", bundles)

			Expect(err).ShouldNot(HaveOccurred())
			Expect(names).Should(Equal([]string{"user_created"}))

			bundle, err := bundleReader.Read(ctx, "t1", "user_created")
			Expect(err).ShouldNot(HaveOccurred())

			Expect(bundle.GetName()).Should(Equal("user_created"))
			Expect(bundle.GetArguments()).Should(Equal([]string{
				"organizationID",
				"userID",
			}))

			Expect(bundle.GetOperations()[0].RelationshipsWrite).Should(Equal([]string{
				"organization:{{.organizationID}}#member@user:{{.userID}}",
				"organization:{{.organizationID}}#admin@user:{{.userID}}",
			}))

			Expect(bundle.GetOperations()[0].RelationshipsDelete).Should(BeEmpty())

			Expect(bundle.GetOperations()[0].AttributesWrite).Should(Equal([]string{
				"organization:{{.organizationID}}$public|boolean:true",
			}))

			Expect(bundle.GetOperations()[0].AttributesDelete).Should(Equal([]string{
				"organization:{{.organizationID}}$balance|integer[]:120,568",
			}))
		})

		It("should get error on non-existing bundle", func() {
			ctx := context.Background()

			_, err := bundleReader.Read(ctx, "t1", "user_created")
			Expect(err.Error()).Should(Equal(base.ErrorCode_ERROR_CODE_BUNDLE_NOT_FOUND.String()))
		})
	})
})
