package backends

import (
	"context"

	"github.com/PerformLine/pivot/v4/dal"
	"github.com/PerformLine/pivot/v4/filter"
)

type Aggregator interface {
	AggregatorConnectionString() *dal.ConnectionString
	AggregatorInitialize(Backend) error
	Sum(ctx context.Context, collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
	Count(ctx context.Context, collection *dal.Collection, f ...*filter.Filter) (uint64, error)
	Minimum(ctx context.Context, collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
	Maximum(ctx context.Context, collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
	Average(ctx context.Context, collection *dal.Collection, field string, f ...*filter.Filter) (float64, error)
	GroupBy(ctx context.Context, collection *dal.Collection, fields []string, aggregates []filter.Aggregate, f ...*filter.Filter) (*dal.RecordSet, error)
}
