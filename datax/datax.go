// Copyright 2018 Eryx <evorui аt gmаil dοt cοm>, All rights reserved.
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
	"strings"
	"sync"
	"time"

	"github.com/hooto/hlog4g/hlog"

	"github.com/hooto/hpress/store"
)

var (
	worker_counter_locker sync.Mutex
	worker_pending        = false
)

func Worker() {

	worker_counter_locker.Lock()
	defer worker_counter_locker.Unlock()

	if worker_pending {
		return
	}

	worker_pending = true

	go func() {

		limit := 1000

		for {

			time.Sleep(1e9)
			if store.DataLocal == nil {
				continue
			}

			gdocRefresh()

			for {

				ls := store.DataLocal.NewRanger(
					[]byte("access_counter"), []byte("access_counter")).
					SetLimit(int64(limit)).Exec()

				imap := map[string]int{}

				for _, v := range ls.Items {

					s := strings.Split(string(v.Key), "/")

					if len(s) == 4 {

						key := s[1] + "/" + s[3]
						if _, ok := imap[key]; ok {
							imap[key]++
						} else {
							imap[key] = 1
						}
					}

					store.DataLocal.NewDeleter(v.Key).Exec()
				}

				for key, num := range imap {

					ks := strings.Split(key, "/")

					if len(ks) != 2 {
						continue
					}

					q := store.Data.NewQueryer().From(ks[0]).Limit(1)
					q.Where().And("id", ks[1])

					if rs, err := store.Data.Query(q); err == nil && len(rs) > 0 {

						ft := store.Data.NewFilter()
						ft.And("id", ks[1])

						store.Data.Update(ks[0], map[string]interface{}{
							"ext_access_counter": rs[0].Field("ext_access_counter").Int() + num,
						}, ft)
					}
				}

				if !ls.NextResultSet {
					break
				}
			}

			if err := data_search_sync(); err != nil {
				hlog.Printf("error", "data_search_sync error : %s", err.Error())
			}

			if err := data_sync_pull(); err != nil {
				hlog.Printf("error", "data_sync_pull error : %s", err.Error())
			}
		}

	}()
}
