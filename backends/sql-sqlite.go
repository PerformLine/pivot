package backends

import (
	// "fmt"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/pathutil"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

func (self *SqlBackend) initializeSqlite() (string, string, error) {
	dataset := self.conn.Dataset()
	var dsn string

	switch dataset {
	case `/memory`:
		return `sqlite3`, `:memory:`, nil
	default:
		if strings.HasPrefix(dataset, `/.`) {
			dataset = strings.TrimPrefix(dataset, `/`)
		} else if strings.HasPrefix(dataset, `/~`) {
			dataset = strings.TrimPrefix(dataset, `/`)

			if v, err := pathutil.ExpandUser(dataset); err == nil {
				dataset = v
			} else {
				return ``, ``, err
			}
		}

		dsn = dataset

		opts := make(map[string]interface{})

		if v := self.conn.OptString(`cache`, `shared`); v != `` {
			opts[`cache`] = v
		}

		if v := self.conn.OptString(`mode`, `memory`); v != `` {
			opts[`mode`] = v
		}

		if len(opts) > 0 {
			dsn = dsn + `?` + maputil.Join(opts, `=`, `&`)
		}

		return `sqlite3`, dsn, nil
	}
}
