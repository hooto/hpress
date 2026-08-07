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
	"github.com/lynkdb/iomix/rdb"
	"github.com/lynkdb/mysqlgo"
	"github.com/lynkdb/pgsqlgo"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/store"
)

func utf8_rune_filter(str string) string {
	strs, outs := []rune(str), []rune{}
	for _, v := range strs {
		if utf8.ValidRune(v) && v != 0 {
			outs = append(outs, v)
		}
	}
	return string(outs)
}

func data_sync_pull() error {

	if len(config.Config.ExtUpDatabases) == 0 {
		return nil
	}

	var cfgs types.KvPairs
	if rs := store.DataLocal.NewReader(api.NsSysDataPull()).Exec(); rs.OK() {
		rs.JsonDecode(&cfgs)
	}

	var (
		limit int64 = 50
		src   rdb.Connector
		err   error
		tng   = uint32(time.Now().Unix())
		dtbs  types.ArrayString
	)
	defer func() {
		if src != nil {
			src.Close()
		}
	}()

	dmr, err := store.Data.Modeler()
	if err != nil {
		return err
	}

	if tbs, err := dmr.TableDump(); err != nil {
		return err
	} else {

		for _, vt := range tbs {
			dtbs.Set(vt.Name)
		}
	}

	for _, cv := range config.Config.ExtUpDatabases {

		// fmt.Println("\n\ndb sync", cv.Name)

		if src != nil {
			// TODO
			src.Close()
		}

		switch cv.Driver {
		case "lynkdb/mysqlgo":
			src, err = mysqlgo.NewConnector(*cv)

		case "lynkdb/pgsqlgo":
			src, err = pgsqlgo.NewConnector(*cv)

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

		mr, err := src.Modeler()
		if err != nil {
			return err
		}

		tbs, err := mr.TableDump()
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
				cnew    = 0
				cupd    = 0
				cign    = 0
				q       = src.NewQueryer().From(vt.Name).Order("updated ASC").Limit(limit)
				offset  = int64(0)
				up_name = fmt.Sprintf("sync-time/%s:%s/%s",
					cv.Value("host"), cv.Value("port"), vt.Name)
				up_offset = uint32(0)
			)
			err = nil

			if pv := cfgs.Get(up_name); pv.Uint32() > 0 {
				up_offset = pv.Uint32()
				q.Where().And("updated.ge", up_offset)
				// slog.Warn(fmt.Sprintf("%s updated.ge %d", vt.Name, pv.Uint32()))
			}

			// fmt.Println("\nTABLE", vt.Name, tn, tng)

			for {

				rs, err := src.Query(q)
				if err != nil {
					slog.Warn(fmt.Sprintf("%s query error %s", vt.Name, err.Error()))
					break
				}

				for _, v := range rs {

					tup := v.Field("updated").Uint32()
					if tup < tng && tup > up_offset {
						up_offset = tup
					}

					sets := map[string]interface{}{}
					ext_counter := 0
					for k, f := range v.Fields {
						if k == "ext_access_counter" {
							ext_counter = f.Int()
							continue
						}
						sets[k] = f.String()
					}

					qr := store.Data.NewQueryer().From(vt.Name)
					fr := store.Data.NewFilter().And("id", v.Field("id").String())
					qr.SetFilter(fr)
					rsi, err := store.Data.Fetch(qr)

					if rsi.NotFound() {

						if ext_counter > 0 {
							sets["ext_access_counter"] = ext_counter
						}

						_, err = store.Data.Insert(vt.Name, sets)
						if err != nil {
							if strings.Contains(err.Error(), "invalid byte sequence for encoding") {
								for sk, sv := range sets {
									switch sv.(type) {
									case string:
										sets[sk] = utf8_rune_filter(sv.(string))
									}
								}
								_, err = store.Data.Insert(vt.Name, sets)
							}
						}

						if err != nil {
							slog.Warn(fmt.Sprintf("data sync (%s) ErrInsert %s %s",
								up_name, v.Field("id").String(), err.Error()))
							break

						} else {
							// fmt.Println("  OK INSERT", vt.Name, v.Field("id").String())
							cnew += 1
						}

					} else if err != nil {
						slog.Warn(fmt.Sprintf("data sync (%s), ID: %s, QueryError %s",
							vt.Name, v.Field("id").String(), err.Error()))
						break
					} else {

						var (
							tlc          = rsi.Field("updated").Uint32()
							sync_counter = false
						)

						if ext_counter > 0 {
							if ext_counter > rsi.Field("ext_access_counter").Int() {
								if tup > tlc {
									sets["ext_access_counter"] = ext_counter
								} else {
									sets = map[string]interface{}{
										"ext_access_counter": ext_counter,
									}
								}
								sync_counter = true
							}
						}

						if tup > tlc || sync_counter {

							_, err = store.Data.Update(vt.Name, sets, fr)

							if err != nil {
								if strings.Contains(err.Error(), "invalid byte sequence for encoding") {
									for sk, sv := range sets {
										switch sv.(type) {
										case string:
											sets[sk] = utf8_rune_filter(sv.(string))
										}
									}
									_, err = store.Data.Update(vt.Name, sets, fr)
								}
							}

							if err != nil {
								slog.Warn(fmt.Sprintf("data sync (%s) ErrUpdate %s %s",
									up_name, v.Field("id").String(), err.Error()))
								// fmt.Println("  ER UPDATE", vt.Name, v.Field("id").String())
								break
							} else {
								// fmt.Println("  OK UPDATE", vt.Name, v.Field("id").String())
								cupd += 1
							}
						} else {
							// fmt.Println("  OK IGNORE ", vt.Name, v.Field("id").String())
							cign += 1
						}

						continue
					}

				}

				if err != nil || len(rs) < int(limit) {
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
						up_name, cnew, cupd, cign))
					cfgs.Set(up_name, up_offset)
				}
			} else {
				slog.Warn(fmt.Sprintf("data sync ((%s) error : %s",
					up_name, err.Error()))
			}
		}
	}

	if rs := store.DataLocal.NewWriter(api.NsSysDataPull(), nil).SetJsonValue(cfgs).Exec(); !rs.OK() {
		// fmt.Println("  DATA PULL TAG ERROR")
	}

	return nil
}
