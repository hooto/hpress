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
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/hooto/hpress/internal/store"
)

var (
	workerCounterLocker sync.Mutex
	workerPending       = false
)

func Worker() {

	workerCounterLocker.Lock()
	defer workerCounterLocker.Unlock()

	if workerPending {
		return
	}

	workerPending = true

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

					if rs := store.Data.Query(q); rs.OK() && !rs.NotFound() {

						ft := store.Data.NewFilter()
						ft.And("id", ks[1])

						store.Data.Update(ks[0], map[string]interface{}{
							"ext_access_counter": rs.Field("ext_access_counter").Int() + num,
						}, ft)
					}
				}

				if !ls.NextResultSet {
					break
				}
			}

			if err := dataSearchSync(); err != nil {
				slog.Error(fmt.Sprintf("dataSearchSync error : %s", err.Error()))
			}

			if err := dataSyncPull(); err != nil {
				slog.Error(fmt.Sprintf("dataSyncPull error : %s", err.Error()))
			}
		}

	}()
}
