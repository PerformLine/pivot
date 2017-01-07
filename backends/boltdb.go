package backends

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/pivot/dal"
	"gopkg.in/mgo.v2/bson"
	"os"
	"path"
)

var BoltDatabaseMode = os.FileMode(0644)
var BoltDatabaseOptions *bolt.Options = nil
var BoltDefaultSearchIndexer = `bleve:///`
var BoltDatabaseIndexSubdirectory = `indexes`

// BoltDB ->
//
type BoltBackend struct {
	Backend
	conn    dal.ConnectionString
	db      *bolt.DB
	indexer Indexer
	options ConnectOptions
}

func NewBoltBackend(connection dal.ConnectionString) *BoltBackend {
	return &BoltBackend{
		conn: connection,
	}
}

func (self *BoltBackend) GetConnectionString() *dal.ConnectionString {
	return &self.conn
}

func (self *BoltBackend) SetOptions(options ConnectOptions) {
	self.options = options
}

func (self *BoltBackend) Initialize() error {
	dbBaseDir := self.conn.Dataset()
	dbFileName := `data.boltdb`

	if db, err := bolt.Open(path.Join(dbBaseDir, dbFileName), BoltDatabaseMode, BoltDatabaseOptions); err == nil {
		self.db = db
	} else {
		return err
	}

	if indexConnString := sliceutil.OrString(
		self.options.Indexer,
		BoltDefaultSearchIndexer,
	); indexConnString != `` {
		if ics, err := dal.ParseConnectionString(indexConnString); err == nil {
			// an empty path denotes using the same parent directory as the DB we're indexing
			if ics.Dataset() == `/` {
				ics.URI.Path = path.Join(dbBaseDir, BoltDatabaseIndexSubdirectory)
			}

			if indexer, err := MakeIndexer(ics); err == nil {
				if err := indexer.IndexInitialize(self); err == nil {
					self.indexer = indexer
					log.Debugf("Search indexing enabled for %T backend at %q", self, self.indexer.IndexConnectionString())
				} else {
					return err
				}
			} else {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (self *BoltBackend) Insert(collection string, recordset *dal.RecordSet) error {
	return self.upsertRecords(collection, recordset, true)
}

func (self *BoltBackend) Exists(collection string, id interface{}) bool {
	exists := false

	self.db.View(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket([]byte(collection[:])); bucket != nil {
			idS := fmt.Sprintf("%v", id)

			if data := bucket.Get([]byte(idS[:])); data != nil {
				exists = true
			}
		}

		return nil
	})

	return exists
}

func (self *BoltBackend) Retrieve(collection string, id interface{}, fields ...string) (*dal.Record, error) {
	record := dal.NewRecord(id)

	err := self.db.View(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket([]byte(collection[:])); bucket != nil {
			idS := fmt.Sprintf("%v", id)

			if data := bucket.Get([]byte(idS[:])); data != nil {
				return bson.Unmarshal(data, record)
			} else {
				return fmt.Errorf("Record %q does not exist", id)
			}
		} else {
			return fmt.Errorf("Failed to retrieve bucket %q", collection)
		}

		return nil
	})

	if err == nil {
		return record, nil
	} else {
		return nil, err
	}
}

func (self *BoltBackend) Update(collection string, recordset *dal.RecordSet, target ...string) error {
	if len(target) > 0 {
		return fmt.Errorf("Multi-target updates are not supported on this backend")
	}

	return self.upsertRecords(collection, recordset, false)
}

func (self *BoltBackend) Delete(collection string, ids ...interface{}) error {
	switch len(ids) {
	case 0:
		// TODO: implementent delete `all`
		return fmt.Errorf("Delete requires at least one criterion")

	case 1:
		return self.db.Update(func(tx *bolt.Tx) error {
			if bucket := tx.Bucket([]byte(collection[:])); bucket != nil {
				// perform deletes
				for _, id := range ids {
					idStr := fmt.Sprintf("%v", id)

					if err := bucket.Delete([]byte(idStr[:])); err != nil {
						return err
					}
				}

				// if we have a search index, remove the corresponding documents from it
				if search := self.WithSearch(); search != nil {
					if err := search.IndexRemove(collection, ids); err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("Failed to retrieve bucket %q", collection)
			}

			return nil
		})
	default:
		return fmt.Errorf("Delete by query not supported on this backend")
	}
}

func (self *BoltBackend) WithSearch() Indexer {
	return self.indexer
}

func (self *BoltBackend) ListCollections() ([]string, error) {
	collectionNames := make([]string, 0)

	err := self.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			collectionNames = append(collectionNames, string(name[:]))
			return nil
		})
	})

	return collectionNames, err
}

func (self *BoltBackend) CreateCollection(definition dal.Collection) error {
	return self.db.Update(func(tx *bolt.Tx) error {
		bucketName := []byte(definition.Name[:])
		_, err := tx.CreateBucket(bucketName)
		return err
	})
}

func (self *BoltBackend) DeleteCollection(collection string) error {
	return self.db.Update(func(tx *bolt.Tx) error {
		bucketName := []byte(collection[:])
		return tx.DeleteBucket(bucketName)
	})
}

func (self *BoltBackend) GetCollection(name string) (*dal.Collection, error) {
	collection := dal.NewCollection(name)

	err := self.db.View(func(tx *bolt.Tx) error {
		if bucket := tx.Bucket([]byte(name[:])); bucket != nil {
			collection.Name = name
		} else {
			return fmt.Errorf("No such collection %q", name)
		}

		return nil
	})

	return collection, err
}

func (self *BoltBackend) upsertRecords(collection string, recordset *dal.RecordSet, autocreateBucket bool) error {
	defer stats.NewTiming().Send(`pivot.backends.boltdb.upsert_time`)

	return self.db.Update(func(tx *bolt.Tx) error {
		stats.Increment(`pivot.backends.boltdb.upsert`)

		bucketName := []byte(collection[:])
		bucket := tx.Bucket(bucketName)

		if bucket == nil {
			if autocreateBucket {
				stats.Increment(`pivot.backends.boltdb.create_collection`)

				if b, err := tx.CreateBucket(bucketName); err == nil {
					bucket = b
				} else {
					return fmt.Errorf("Failed to create bucket %q: %v", collection, err)
				}
			} else {
				return fmt.Errorf("Failed to retrieve bucket %q", collection)
			}
		}

		for _, record := range recordset.Records {
			tm := stats.NewTiming()

			if data, err := bson.Marshal(record); err == nil {
				tm.Send(`pivot.backends.boltdb.serialize_record`)
				tm = stats.NewTiming()

				idS := fmt.Sprintf("%v", record.ID)

				if err := bucket.Put([]byte(idS[:]), data); err == nil {
					tm.Send(`pivot.backends.boltdb.commit_record`)
				} else {
					return err
				}
			} else {
				return err
			}
		}

		if search := self.WithSearch(); search != nil {
			tm := stats.NewTiming()

			if err := search.Index(collection, recordset); err == nil {
				tm.Send(`pivot.backends.boltdb.index_record`)
			} else {
				return err
			}
		}

		return nil
	})
}
