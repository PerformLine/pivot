package backends

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/PerformLine/go-stockutil/log"
	"github.com/PerformLine/go-stockutil/sliceutil"
	"github.com/PerformLine/pivot/v4/dal"
	"github.com/PerformLine/pivot/v4/filter"
)

var IndexerPageSize int = 100
var MaxFacetCardinality int = 10000
var DefaultCompoundJoiner = `:`

type IndexPage struct {
	Page         int
	TotalPages   int
	Limit        int
	Offset       int
	TotalResults int64
}

type IndexResultFunc func(record *dal.Record, err error, page IndexPage) error // {}

type Indexer interface {
	IndexConnectionString() *dal.ConnectionString
	IndexInitialize(Backend) error
	IndexExists(ctx context.Context, collection *dal.Collection, id interface{}) bool
	IndexRetrieve(ctx context.Context, collection *dal.Collection, id interface{}) (*dal.Record, error)
	IndexRemove(ctx context.Context, collection *dal.Collection, ids []interface{}) error
	Index(ctx context.Context, collection *dal.Collection, records *dal.RecordSet) error
	QueryFunc(ctx context.Context, collection *dal.Collection, filter *filter.Filter, resultFn IndexResultFunc) error
	Query(ctx context.Context, collection *dal.Collection, filter *filter.Filter, resultFns ...IndexResultFunc) (*dal.RecordSet, error)
	ListValues(ctx context.Context, collection *dal.Collection, fields []string, filter *filter.Filter) (map[string][]interface{}, error)
	DeleteQuery(ctx context.Context, collection *dal.Collection, f *filter.Filter) error
	FlushIndex() error
	GetBackend() Backend
}

func MakeIndexer(connection dal.ConnectionString) (Indexer, error) {
	log.Infof("Creating indexer: %v", connection.String())

	switch connection.Backend() {
	case `elasticsearch`:
		return NewElasticsearchIndexer(connection), nil
	default:
		return nil, fmt.Errorf("Unknown indexer type %q", connection.Backend())
	}
}

func PopulateRecordSetPageDetails(recordset *dal.RecordSet, f *filter.Filter, page IndexPage) {
	// result count is whatever we were told it was for this query
	if page.TotalResults >= 0 {
		recordset.KnownSize = true
		recordset.ResultCount = page.TotalResults
	} else {
		recordset.ResultCount = int64(len(recordset.Records))
	}

	if page.TotalPages > 0 {
		recordset.TotalPages = page.TotalPages
	} else if recordset.ResultCount >= 0 && f.Limit > 0 {
		// total pages = ceil(result count / page size)
		recordset.TotalPages = int(math.Ceil(float64(recordset.ResultCount) / float64(f.Limit)))
	} else {
		recordset.TotalPages = 1
	}

	if recordset.RecordsPerPage == 0 {
		recordset.RecordsPerPage = page.Limit
	}

	// page is the last page number set
	if page.Limit > 0 {
		recordset.Page = int(math.Ceil(float64(f.Offset+1) / float64(page.Limit)))
	}
}

func DefaultQueryImplementation(ctx context.Context, indexer Indexer, collection *dal.Collection, f *filter.Filter, resultFns ...IndexResultFunc) (*dal.RecordSet, error) {
	recordset := dal.NewRecordSet()

	if err := indexer.QueryFunc(ctx, collection, f, func(indexRecord *dal.Record, err error, page IndexPage) error {
		defer PopulateRecordSetPageDetails(recordset, f, page)

		parent := indexer.GetBackend()
		var forceIndexRecord bool

		// look for a filter option that specifies that we should explicitly NOT attempt to retrieve the
		// record from the parent by ID, but rather always use the index record as-is.
		if f != nil {
			if vI, ok := f.Options[`ForceIndexRecord`]; ok {
				if v, ok := vI.(bool); ok {
					forceIndexRecord = v
				}
			}
		}

		// index compound field processing
		if parent != nil {
			if len(collection.IndexCompoundFields) > 1 {
				joiner := collection.IndexCompoundFieldJoiner

				if joiner == `` {
					joiner = DefaultCompoundJoiner
				}

				var values []interface{}

				values = sliceutil.Sliceify(
					strings.Split(fmt.Sprintf("%v", indexRecord.ID), joiner),
				)

				if len(values) != len(collection.IndexCompoundFields) {
					// if the index record's ID isn't holding _all_ of the field components,
					// attempt to retrieve the rest of the values from the record itself.
					if diff := len(collection.IndexCompoundFields) - len(values); diff >= 1 {
						for _, cf := range collection.IndexCompoundFields[diff:] {
							if v := indexRecord.Get(cf); v != nil {
								values = append(values, v)
							}
						}
					}

					if len(values) != len(collection.IndexCompoundFields) {
						return fmt.Errorf(
							"Index item %v: expected %d compound field components, got %d",
							indexRecord.ID,
							len(collection.IndexCompoundFields),
							len(values),
						)
					}
				}

				for i, field := range collection.IndexCompoundFields {
					if i == 0 {
						indexRecord.ID = values
					} else {
						indexRecord.Set(field, values[i])
					}
				}
			}
		}

		emptyRecord := dal.NewRecord(indexRecord.ID)
		emptyRecord.Error = err

		if len(resultFns) > 0 {
			resultFn := resultFns[0]

			if f.IdOnly() {
				return resultFn(emptyRecord, err, page)
			} else if parent != nil && !forceIndexRecord {
				if record, err := parent.Retrieve(ctx, collection.Name, indexRecord.ID, f.Fields...); err == nil {
					return resultFn(record, err, page)
				} else {
					return resultFn(emptyRecord, err, page)
				}
			} else {
				return resultFn(indexRecord, err, page)
			}
		} else {
			if f.IdOnly() {
				recordset.Records = append(recordset.Records, dal.NewRecord(indexRecord.ID))

			} else if parent != nil && !forceIndexRecord {
				if record, err := parent.Retrieve(ctx, collection.Name, indexRecord.ID, f.Fields...); err == nil {
					recordset.Records = append(recordset.Records, record)

				} else {
					recordset.Records = append(recordset.Records, dal.NewRecordErr(indexRecord.ID, err))
				}
			} else {
				recordset.Records = append(recordset.Records, indexRecord)
			}

			return nil
		}
	}); err != nil {
		return nil, err
	}

	return recordset, nil
}
