package generators

import (
	"fmt"

	"github.com/PerformLine/go-stockutil/typeutil"
	"github.com/PerformLine/pivot/v4/filter"
)

var ElasticsearchExactMatchQueryType = `term`
var ElasticsearchFulltextDefaultConjunctionAnd = true

func esCriterionOperatorIs(gen *Elasticsearch, criterion filter.Criterion) (map[string]interface{}, error) {
	c := make(map[string]interface{})

	if len(criterion.Values) == 1 && criterion.Values[0] == `null` {
		c[`missing`] = map[string]interface{}{
			`field`:      criterion.Field,
			`existence`:  true,
			`null_value`: true,
		}

		gen.values = append(gen.values, nil)
	} else {
		or_terms := make([]map[string]interface{}, 0)

		for _, value := range criterion.Values {
			gen.values = append(gen.values, value)

			or_terms = append(or_terms, map[string]interface{}{
				ElasticsearchExactMatchQueryType: map[string]interface{}{
					criterion.Field: value,
				},
			})

			if v, ok := gen.options[`multifield`]; ok {
				if vS, ok := v.(string); ok {
					or_terms = append(or_terms, map[string]interface{}{
						ElasticsearchExactMatchQueryType: map[string]interface{}{
							(criterion.Field + `.` + vS): value,
						},
					})
				}
			}

			c[`bool`] = map[string]interface{}{
				`should`: or_terms,
			}
		}
	}

	return c, nil
}

func esCriterionOperatorNot(gen *Elasticsearch, criterion filter.Criterion) (map[string]interface{}, error) {
	c := make(map[string]interface{})

	if len(criterion.Values) == 0 {
		return c, fmt.Errorf("The not criterion must have at least one value")

	} else if len(criterion.Values) == 1 && criterion.Values[0] == `null` {
		gen.values = append(gen.values, nil)

		c[`bool`] = map[string]interface{}{
			`must_not`: map[string]interface{}{
				`missing`: map[string]interface{}{
					`field`:      criterion.Field,
					`existence`:  true,
					`null_value`: true,
				},
			},
		}
	} else {
		and_not := make([]map[string]interface{}, 0)

		for _, value := range criterion.Values {
			gen.values = append(gen.values, value)

			//  strings get treated as regular expressions
			if criterion.Type == `str` {
				and_not = append(and_not, map[string]interface{}{
					`bool`: map[string]interface{}{
						`must_not`: map[string]interface{}{
							`regexp`: map[string]interface{}{
								criterion.Field: map[string]interface{}{
									`value`: value,
									`flags`: `ALL`,
								},
							},
						},
					},
				})
			} else {
				//  all other types are a simple term match
				and_not = append(and_not, map[string]interface{}{
					`bool`: map[string]interface{}{
						`must_not`: map[string]interface{}{
							ElasticsearchExactMatchQueryType: map[string]interface{}{
								criterion.Field: value,
							},
						},
					},
				})
			}
		}

		c[`bool`] = map[string]interface{}{
			`must`: and_not,
		}
	}

	return c, nil
}

func esCriterionOperatorPattern(gen *Elasticsearch, opname string, criterion filter.Criterion) (map[string]interface{}, error) {
	var c = make(map[string]interface{})

	if len(criterion.Values) == 0 {
		return nil, fmt.Errorf("The not criterion must have at least one value")
	} else {
		var or_regexp = make([]map[string]interface{}, 0)

		for _, value := range criterion.Values {
			gen.values = append(gen.values, value)
			var valueClause string

			switch opname {
			case `contains`:
				valueClause = fmt.Sprintf(".*%s.*", value)
			case `prefix`:
				valueClause = fmt.Sprintf("%s.*", value)
			case `suffix`:
				valueClause = fmt.Sprintf(".*%s", value)
			default:
				return nil, fmt.Errorf("Unsupported pattern operator %q", opname)
			}

			or_regexp = append(or_regexp, map[string]interface{}{
				`regexp`: map[string]interface{}{
					criterion.Field: map[string]interface{}{
						`value`: valueClause,
						`flags`: `ALL`,
					},
				},
			})

			if v, ok := gen.options[`multifield`]; ok {
				if vS, ok := v.(string); ok {
					or_regexp = append(or_regexp, map[string]interface{}{
						`regexp`: map[string]interface{}{
							(criterion.Field + `.` + vS): map[string]interface{}{
								`value`: valueClause,
								`flags`: `ALL`,
							},
						},
					})
				}
			}
		}

		c[`bool`] = map[string]interface{}{
			`should`: or_regexp,
		}
	}

	return c, nil
}

func esCriterionOperatorRange(gen *Elasticsearch, criterion filter.Criterion, operator string) (map[string]interface{}, error) {
	var c = make(map[string]interface{})

	if l := len(criterion.Values); l == 1 {
		gen.values = append(gen.values, criterion.Values[0])

		c[`range`] = map[string]interface{}{
			criterion.Field: map[string]interface{}{
				operator: criterion.Values[0],
			},
		}
	} else {
		return c, fmt.Errorf("Ranging criteria can only accept one value, %d given", l)
	}

	return c, nil
}

func esCriterionOperatorPivotRange(gen *Elasticsearch, criterion filter.Criterion) (map[string]interface{}, error) {
	var c = make(map[string]interface{})

	if l := len(criterion.Values); l == 2 {
		gen.values = append(gen.values, criterion.Values...)

		c[`range`] = map[string]interface{}{
			criterion.Field: map[string]interface{}{
				"gte": criterion.Values[0],
				"lt":  criterion.Values[1],
			},
		}
	} else {
		return c, fmt.Errorf("Ranging criteria can only accept two values, %d given", l)
	}

	return c, nil
}

func esCriterionOperatorFulltext(gen *Elasticsearch, criterion filter.Criterion) (map[string]interface{}, error) {
	var c = make(map[string]interface{})

	var or_queries = make([]map[string]interface{}, 0)
	var defop string

	if ElasticsearchFulltextDefaultConjunctionAnd {
		defop = `AND`
	} else {
		defop = `OR`
	}

	for _, value := range criterion.Values {
		gen.values = append(gen.values, value)

		or_queries = append(or_queries, map[string]interface{}{
			`query_string`: map[string]interface{}{
				`query`:            typeutil.String(value),
				`default_field`:    criterion.Field,
				`default_operator`: defop,
				`lenient`:          true,
			},
		})
	}

	c[`bool`] = map[string]interface{}{
		`should`: or_queries,
	}

	return c, nil
}
