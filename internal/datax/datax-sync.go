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
	"time"
	"unicode/utf8"

	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/lynkapi/go/lynktable"
	"github.com/lynkdb/mysqlgo"
	"github.com/lynkdb/pgsqlgo"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
)

func utf8RuneFilter(str string) string {
	strs, outs := []rune(str), []rune{}
	for _, v := range strs {
		if utf8.ValidRune(v) && v != 0 {
			outs = append(outs, v)
		}
	}
	return string(outs)
}

func dataSyncPull() error {

	if len(config.Config.ExtUpDatabases) == 0 {
		return nil
	}

	var cfgs types.KvPairs
	if rs := store.DataLocal.NewReader(hpapi.NsSysDataPull()).Exec(); rs.OK() {
		rs.JsonDecode(&cfgs)
	}

	var (
		limit int64 = 50
		src   lynktable.Connector
		err   error
		tng   = uint32(time.Now().Unix())
		dtbs  types.ArrayString
	)
	defer func() {
		if src != nil {
			src.Close()
		}
	}()

	dmr := store.Data.Modeler()

	tbs, err := dmr.TableList()
	if err != nil {
		return err
	}

	for _, vt := range tbs {
		dtbs.Set(vt.Name)
	}

	for _, cv := range config.Config.ExtUpDatabases {

		// fmt.Println("\n\ndb sync", cv.Name)

		if src != nil {
			// TODO
			src.Close()
		}

		switch cv.Driver {
		case "lynkdb/mysqlgo":
			src, err = mysqlgo.NewConnector(store.ConnOptsMap(cv))

		case "lynkdb/pgsqlgo":
			src, err = pgsqlgo.NewConnector(store.ConnOptsMap(cv))

		default:
			continue
		}

		if err != nil {
			slog.Warn(fmt.Sprintf("data connect ((%s) error : %s",
				cv.Name, err.Error()))
			continue
		}

		if src == nil {
			continue
		}

		mr := src.Modeler()

		tbs, err := mr.TableList()
		if err != nil {
			return err
		}

		for _, vt := range tbs {

			if !strings.HasPrefix(vt.Name, "hpt_") &&
				!strings.HasPrefix(vt.Name, "hpn_") {
				continue
			}

			if !dtbs.Has(vt.Name) {
				continue
			}

			var (
				cnew   = 0
				cupd   = 0
				cign   = 0
				q      = src.NewQueryer().From(vt.Name).Order("updated ASC").Limit(limit)
				offset = int64(0)
				upName = fmt.Sprintf("sync-time/%s:%s/%s",
					cv.Value("host"), cv.Value("port"), vt.Name)
				upOffset = uint32(0)
			)
			err = nil

			if pv := cfgs.Get(upName); pv.Uint32() > 0 {
				upOffset = pv.Uint32()
				q.Where().And("updated.ge", upOffset)
				// slog.Warn(fmt.Sprintf("%s updated.ge %d", vt.Name, pv.Uint32()))
			}

			// fmt.Println("\nTABLE", vt.Name, tn, tng)

			for {

				rs := src.Query(q)
				if rs.Err() != nil {
					slog.Warn(fmt.Sprintf("%s query error %s", vt.Name, rs.Err().Error()))
					break
				}

				for rs.Valid() {

					tup := rs.Field("updated").Uint32()
					if tup < tng && tup > upOffset {
						upOffset = tup
					}

					sets := map[string]interface{}{}
					extCounter := 0
					for _, k := range rs.Columns() {
						if k == "ext_access_counter" {
							extCounter = rs.Field(k).Int()
							continue
						}
						sets[k] = rs.Field(k).String()
					}

					qr := store.Data.NewQueryer().From(vt.Name)
					fr := store.Data.NewFilter().And("id", rs.Field("id").String())
					qr.SetFilter(fr)
					rsi := store.Data.Query(qr)

					if rsi.NotFound() {

						if extCounter > 0 {
							sets["ext_access_counter"] = extCounter
						}

						err = store.Data.Insert(vt.Name, sets).Err()
						if err != nil {
							if strings.Contains(err.Error(), "invalid byte sequence for encoding") {
								for sk, sv := range sets {
									switch sv.(type) {
									case string:
										sets[sk] = utf8RuneFilter(sv.(string))
									}
								}
								err = store.Data.Insert(vt.Name, sets).Err()
							}
						}

						if err != nil {
							slog.Warn(fmt.Sprintf("data sync (%s) ErrInsert %s %s",
								upName, rs.Field("id").String(), err.Error()))
							break

						} else {
							// fmt.Println("  OK INSERT", vt.Name, rs.Field("id").String())
							cnew += 1
						}

					} else if rsi.Err() != nil {
						slog.Warn(fmt.Sprintf("data sync (%s), ID: %s, QueryError %s",
							vt.Name, rs.Field("id").String(), rsi.Err().Error()))
						break
					} else {

						var (
							tlc         = rsi.Field("updated").Uint32()
							syncCounter = false
						)

						if extCounter > 0 {
							if extCounter > rsi.Field("ext_access_counter").Int() {
								if tup > tlc {
									sets["ext_access_counter"] = extCounter
								} else {
									sets = map[string]interface{}{
										"ext_access_counter": extCounter,
									}
								}
								syncCounter = true
							}
						}

						if tup > tlc || syncCounter {

							err = store.Data.Update(vt.Name, sets, fr).Err()

							if err != nil {
								if strings.Contains(err.Error(), "invalid byte sequence for encoding") {
									for sk, sv := range sets {
										switch sv.(type) {
										case string:
											sets[sk] = utf8RuneFilter(sv.(string))
										}
									}
									err = store.Data.Update(vt.Name, sets, fr).Err()
								}
							}

							if err != nil {
								slog.Warn(fmt.Sprintf("data sync (%s) ErrUpdate %s %s",
									upName, rs.Field("id").String(), err.Error()))
								// fmt.Println("  ER UPDATE", vt.Name, rs.Field("id").String())
								break
							} else {
								// fmt.Println("  OK UPDATE", vt.Name, rs.Field("id").String())
								cupd += 1
							}
						} else {
							// fmt.Println("  OK IGNORE ", vt.Name, rs.Field("id").String())
							cign += 1
						}
					}

					rs.Next()
				}

				if n, _ := rs.RowsAffected(); err != nil || n < limit {
					// fmt.Printf("  DONE INSERT/IGNORE %d, UPDATE %d, ALL %d\n",
					// 	cnew, cupd, int(offset)+len(rs))
					break
				}

				offset += limit
				q.Offset(offset)
			}

			if err == nil {
				if cnew > 0 || cupd > 0 {
					slog.Info(fmt.Sprintf("data sync (%s) INSERT %d, UPDATE %d, IGNORE %d",
						upName, cnew, cupd, cign))
					cfgs.Set(upName, upOffset)
				}
			} else {
				slog.Warn(fmt.Sprintf("data sync ((%s) error : %s",
					upName, err.Error()))
			}
		}
	}

	if rs := store.DataLocal.NewWriter(hpapi.NsSysDataPull(), nil).SetJsonValue(cfgs).Exec(); !rs.OK() {
		// fmt.Println("  DATA PULL TAG ERROR")
	}

	return nil
}
