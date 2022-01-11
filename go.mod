module github.com/PerformLine/pivot/v3

replace github.com/ghetzel/pivot/v3 v3.4.0 => github.com/PerformLine/pivot/v3 v3.4.0

require (
	cloud.google.com/go v0.56.0 // indirect
	github.com/alecthomas/chroma v0.7.3 // indirect
	github.com/alexcesaro/statsd v2.0.0+incompatible
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/aws/aws-sdk-go v1.34.13
	github.com/blevesearch/bleve v0.7.0
	github.com/deckarep/golang-set v0.0.0-20171013212420-1d4478f51bed
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/fatih/structs v1.0.0
	github.com/ghetzel/cli v1.17.0
	github.com/ghetzel/diecast v1.18.0
	github.com/ghetzel/go-stockutil v1.9.5
	github.com/ghetzel/pivot/v3 v3.4.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-shiori/go-readability v0.0.0-20200413080041-05caea5f6592 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/hashicorp/golang-lru v0.5.1
	github.com/husobee/vestigo v1.1.0
	github.com/jbenet/go-base58 v0.0.0-20150317085156-6237cf65f3a6
	github.com/jdxcode/netrc v0.0.0-20201119100258-050cafb6dbe6
	github.com/lib/pq v1.2.0
	github.com/mattn/go-sqlite3 v1.14.4
	github.com/microcosm-cc/bluemonday v1.0.1 // indirect
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20180319144342-a05df785d2dc
	github.com/ory/dockertest v3.3.2+incompatible
	github.com/stretchr/testify v1.6.1
	github.com/urfave/negroni v1.0.1-0.20191011213438-f4316798d5d3
	github.com/yuin/gopher-lua v0.0.0-20210529063254-f4c35e4016d9 // indirect
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/mgo.v2 v2.0.0-20160818020120-3f83fa500528
)

go 1.13

//  replace github.com/ghetzel/go-stockutil v1.8.62 => ../go-stockutil
