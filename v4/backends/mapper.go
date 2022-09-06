package backends

import (
	"context"

	"github.com/PerformLine/pivot/v4/dal"
	"github.com/PerformLine/pivot/v4/filter"
)

type ResultFunc func(ptrToInstance interface{}, err error) // {}

type Mapper interface {
	GetBackend() Backend
	GetCollection() *dal.Collection
	Migrate(ctx context.Context) error
	Drop(ctx context.Context) error
	Exists(ctx context.Context, id interface{}) bool
	Create(ctx context.Context, from interface{}) error
	Get(ctx context.Context, id interface{}, into interface{}) error
	Update(ctx context.Context, from interface{}) error
	CreateOrUpdate(ctx context.Context, id interface{}, from interface{}) error
	Delete(ctx context.Context, ids ...interface{}) error
	DeleteQuery(ctx context.Context, flt interface{}) error
	Find(ctx context.Context, flt interface{}, into interface{}) error
	FindFunc(ctx context.Context, flt interface{}, destZeroValue interface{}, resultFn ResultFunc) error
	All(ctx context.Context, into interface{}) error
	Each(ctx context.Context, destZeroValue interface{}, resultFn ResultFunc) error
	List(ctx context.Context, fields []string) (map[string][]interface{}, error)
	ListWithFilter(ctx context.Context, fields []string, flt interface{}) (map[string][]interface{}, error)
	Sum(ctx context.Context, field string, flt interface{}) (float64, error)
	Count(ctx context.Context, flt interface{}) (uint64, error)
	Minimum(ctx context.Context, field string, flt interface{}) (float64, error)
	Maximum(ctx context.Context, field string, flt interface{}) (float64, error)
	Average(ctx context.Context, field string, flt interface{}) (float64, error)
	GroupBy(ctx context.Context, fields []string, aggregates []filter.Aggregate, flt interface{}) (*dal.RecordSet, error)
}
