package backends

import (
	"context"
	"sync"
	"time"

	"github.com/PerformLine/go-stockutil/typeutil"
	"github.com/PerformLine/pivot/v4/dal"
	"github.com/PerformLine/pivot/v4/filter"
)

type CachingBackend struct {
	backend Backend
	cache   sync.Map
}

func NewCachingBackend(parent Backend) *CachingBackend {
	return &CachingBackend{
		backend: parent,
	}
}

func (self *CachingBackend) ResetCache() {
	self.cache = sync.Map{}
}

func (self *CachingBackend) Retrieve(ctx context.Context, collection string, id interface{}, fields ...string) (*dal.Record, error) {
	cacheset := make(map[interface{}]interface{})

	if c, ok := self.cache.LoadOrStore(collection, cacheset); ok {
		cacheset = c.(map[interface{}]interface{})
	}

	if typeutil.IsScalar(id) {
		if recordI, ok := cacheset[id]; ok {
			return recordI.(*dal.Record), nil
		}
	}

	if record, err := self.backend.Retrieve(ctx, collection, id, fields...); err == nil {
		cacheset[id] = record
		return record, nil
	} else {
		return nil, err
	}
}

// passthrough the remaining functions to fulfill the Backend interface
// -------------------------------------------------------------------------------------------------
func (self *CachingBackend) Exists(ctx context.Context, collection string, id interface{}) bool {
	return self.backend.Exists(ctx, collection, id)
}

func (self *CachingBackend) Initialize() error {
	return self.backend.Initialize()
}

func (self *CachingBackend) SetIndexer(cs dal.ConnectionString) error {
	return self.backend.SetIndexer(cs)
}

func (self *CachingBackend) RegisterCollection(c *dal.Collection) {
	self.backend.RegisterCollection(c)
}

func (self *CachingBackend) GetConnectionString() *dal.ConnectionString {
	return self.backend.GetConnectionString()
}

func (self *CachingBackend) Insert(ctx context.Context, collection string, records *dal.RecordSet) error {
	return self.backend.Insert(ctx, collection, records)
}

func (self *CachingBackend) Update(ctx context.Context, collection string, records *dal.RecordSet, target ...string) error {
	return self.backend.Update(ctx, collection, records, target...)
}

func (self *CachingBackend) Delete(ctx context.Context, collection string, ids ...interface{}) error {
	return self.backend.Delete(ctx, collection, ids...)
}

func (self *CachingBackend) CreateCollection(ctx context.Context, definition *dal.Collection) error {
	return self.backend.CreateCollection(ctx, definition)
}

func (self *CachingBackend) DeleteCollection(ctx context.Context, collection string) error {
	return self.backend.DeleteCollection(ctx, collection)
}

func (self *CachingBackend) ListCollections() ([]string, error) {
	return self.backend.ListCollections()
}

func (self *CachingBackend) GetCollection(collection string) (*dal.Collection, error) {
	return self.backend.GetCollection(collection)
}

func (self *CachingBackend) WithSearch(collection *dal.Collection, filters ...*filter.Filter) Indexer {
	return self.backend.WithSearch(collection, filters...)
}

func (self *CachingBackend) WithAggregator(collection *dal.Collection) Aggregator {
	return self.backend.WithAggregator(collection)
}

func (self *CachingBackend) Flush() error {
	return self.backend.Flush()
}

func (self *CachingBackend) Ping(ctx context.Context, d time.Duration) error {
	return self.backend.Ping(ctx, d)
}

func (self *CachingBackend) String() string {
	return self.backend.String()
}

func (self *CachingBackend) Supports(feature ...BackendFeature) bool {
	return self.backend.Supports(feature...)
}
