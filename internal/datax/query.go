// Copyright 2015 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package datax

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"

	"github.com/lynkdb/lynkapi/go/lynktable"

	"github.com/hooto/hpress/internal/store"
)

type QuerySet struct {
	ModName string
	ModelID string
	cols    string
	Table   string
	order   string
	limit   int64
	offset  int64
	filter  lynktable.Filter
	Pager   bool
}

func NewQuery(modname, table string) *QuerySet {
	return &QuerySet{
		ModName: modname,
		ModelID: table,
		Table:   table,
		cols:    "*",
		limit:   1,
		offset:  0,
	}
}

func (q *QuerySet) Hash() string {

	sql, params := q.Where().Parse()

	ps := []string{}
	for _, p := range params {
		ps = append(ps, fmt.Sprintf("%v", p))
	}

	str := fmt.Sprintf("%s.%s.%s.%s.%d.%d %s,%s",
		q.ModName, q.Table, q.cols, q.order, q.limit, q.offset, sql, strings.Join(ps, ","))

	h := md5.New()
	io.WriteString(h, str)

	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

// Query executes the QuerySet and returns a cursor-style Result.
// The Result is single-pass; iterate with Valid/Next and read via Field.
func (q *QuerySet) Query() lynktable.Result {

	qs := store.Data.NewQueryer().
		Select(q.cols).
		From(q.Table).
		Order(q.order).
		Limit(q.limit).
		Offset(q.offset)

	qs.SetFilter(q.filter)

	return store.Data.Query(qs)
}

func (q *QuerySet) QueryEntry() lynktable.Result {

	q.limit = 1

	return q.Query()
}

func (q *QuerySet) Select(s string) string {
	q.cols = s
	return ""
}

func (q *QuerySet) From(s string) string {
	q.Table = s
	return ""
}

func (q *QuerySet) Order(s string) string {
	q.order = s
	return ""
}

func (q *QuerySet) Limit(num int64) string {
	q.limit = num
	return ""
}

func (q *QuerySet) Offset(num int64) string {
	q.offset = num
	return ""
}

func (q *QuerySet) Where() lynktable.Filter {
	if q.filter == nil {
		q.filter = store.Data.NewFilter()
	}
	return q.filter
}

func (q *QuerySet) Filter(expr string, args ...interface{}) string {
	q.Where().And(expr, args...)
	return ""
}

func (q *QuerySet) FilterOr(expr string, args ...interface{}) string {
	q.Where().Or(expr, args...)
	return ""
}
