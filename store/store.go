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

package store

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hooto/hlog4g/hlog"
	"github.com/lynkdb/iomix/connect"
	"github.com/lynkdb/iomix/rdb"
	"github.com/lynkdb/kvgo/v2/pkg/kvapi"
	"github.com/lynkdb/kvgo/v2/pkg/kvrep"
	"github.com/lynkdb/kvgo/v2/pkg/storage"
	"github.com/lynkdb/mysqlgo"
	"github.com/lynkdb/pgsqlgo"
)

var (
	err         error
	Data        rdb.Connector
	DataOptions *connect.ConnOptions
	DataLocal   kvapi.Client
)

func Setup(dbc *storage.Options, cfg connect.MultiConnOptions) error {

	if dbc == nil {
		return errors.New("No hpress_local Config Found")
	}

	if DataLocal, err = kvrep.NewReplica(dbc); err != nil {
		return fmt.Errorf("db open error %s", err.Error())
	}

	opts := cfg.Options("hpress_database")
	if opts == nil {
		hlog.Print("error", err.Error())
		return errors.New("No hpress_database Config.IoConnectors Found")
	}

	switch opts.Driver {

	case "lynkdb/mysqlgo":
		Data, err = mysqlgo.NewConnector(*opts)

	case "lynkdb/pgsqlgo":
		Data, err = pgsqlgo.NewConnector(*opts)

	default:
		return errors.New("Invalid lynkdb/driver")
	}

	if err != nil {
		hlog.Printf("error", "store_init %s", err.Error())
		return err
	}

	DataOptions = opts

	if err = db_upgrade_0_5(Data); err != nil {
		return err
	}

	if err = db_fix_term_autoincr(Data); err != nil {
		return err
	}

	return nil
}

// db_fix_term_autoincr repairs hpress term tables (hpt_*) whose integer id
// column lost its nextval() default, so newly inserted terms were stored with
// id=0 (the literal column default) instead of an auto-incremented value.
//
// Root cause: the pgsqlgo modeler declares these ids IncrAble (sequence
// seq_<table>__id) and emits ALTER ... SET DEFAULT nextval(), but that default
// did not stick for tables created under older code, and SchemaSync only
// re-runs on a module version bump, so the wrong default (0) persisted.
// rdb.Base.Insert omits the id column, so it fell back to the column default.
// With api.Term.ID being omitempty, id=0 then serializes as missing -> the
// admin UI shows "id == undefined".
//
// This runs at every startup and is idempotent: for each hpt_ table it wires
// the id default to its sequence, reseats the sequence past max(id), and
// reassigns any leftover id=0 row to a fresh id. mysqlgo is unaffected because
// its AUTO_INCREMENT column assigns ids at insert time.
func db_fix_term_autoincr(data rdb.Connector) error {

	if DataOptions.Driver != "lynkdb/pgsqlgo" {
		return nil
	}

	rows, err := data.QueryRaw("SELECT tablename FROM pg_tables WHERE tablename LIKE 'hpt\\_%' ORDER BY tablename")
	if err != nil {
		return fmt.Errorf("term autoincr fix: list tables: %w", err)
	}

	for _, r := range rows {

		t := r.Field("tablename").String()
		if t == "" {
			continue
		}

		seq := "seq_" + t + "__id"
		qt := "\"" + t + "\""

		// ensure the sequence exists (no-op if already present)
		if _, err := data.ExecRaw(fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s", seq)); err != nil {
			hlog.Printf("warn", "term autoincr fix: %s create sequence: %s", t, err.Error())
			continue
		}

		// wire the id default to the sequence
		if _, err := data.ExecRaw(fmt.Sprintf(
			"ALTER TABLE %s ALTER COLUMN id SET DEFAULT nextval('%s')", qt, seq)); err != nil {
			hlog.Printf("warn", "term autoincr fix: %s set default: %s", t, err.Error())
			continue
		}

		// reseat the sequence past the current max(id) so the next insert continues correctly
		maxid := uint64(1)
		if mr, err := data.QueryRaw(fmt.Sprintf("SELECT COALESCE(max(id), 0) AS m FROM %s", qt)); err == nil && len(mr) > 0 {
			if v := mr[0].Field("m").Uint64(); v > maxid {
				maxid = v
			}
		}
		_, _ = data.ExecRaw(fmt.Sprintf("SELECT setval('%s', %d)", seq, maxid))

		// reassign any leftover id=0 row to a fresh id (id is PK, so at most one)
		if ur, err := data.ExecRaw(fmt.Sprintf(
			"UPDATE %s SET id = nextval('%s') WHERE id = 0", qt, seq)); err == nil {
			if n, _ := ur.RowsAffected(); n > 0 {
				hlog.Printf("warn", "term autoincr fix: %s reassigned %d id=0 row(s)", t, n)
			}
		}
	}

	return nil
}

func db_upgrade_0_5(data rdb.Connector) error {

	mdr, _ := data.Modeler()

	tbls, _ := mdr.TableDump()

	for _, tbl := range tbls {

		if strings.HasPrefix(tbl.Name, "nx") ||
			strings.HasPrefix(tbl.Name, "tx") ||
			tbl.Name == "modules" {

			for _, cv := range tbl.Columns {

				if cv.Name != "created" && cv.Name != "updated" {
					continue
				}

				if cv.Type != "datetime" {
					continue
				}

				sqls := []string{}

				hlog.Printf("warn", "store_init upgrade table %s, colume %s, to int",
					tbl.Name, cv.Name)

				switch DataOptions.Driver {

				case "lynkdb/mysqlgo":
					sqls = []string{
						fmt.Sprintf("ALTER TABLE %s ADD time_tmp int", tbl.Name),
						fmt.Sprintf("UPDATE %s SET time_tmp = UNIX_TIMESTAMP(%s)", tbl.Name, cv.Name),
						fmt.Sprintf("ALTER TABLE %s DROP column %s", tbl.Name, cv.Name),
						fmt.Sprintf("ALTER TABLE %s CHANGE time_tmp %s int", tbl.Name, cv.Name),
					}

				case "lynkdb/pgsqlgo":
					sqls = []string{
						fmt.Sprintf("ALTER TABLE %s ADD COLUMN time_tmp bigint", tbl.Name),
						fmt.Sprintf("UPDATE %s SET time_tmp = extract(epoch from %s)", tbl.Name, cv.Name),
						fmt.Sprintf("ALTER TABLE %s DROP column %s", tbl.Name, cv.Name),
						fmt.Sprintf("ALTER TABLE %s RENAME time_tmp TO %s", tbl.Name, cv.Name),
					}
				}

				for _, sql := range sqls {
					if _, err := data.ExecRaw(sql); err != nil {
						return err
					}
				}

				hlog.Printf("warn", "store_init upgrade table %s, colume %s, to int, DONE",
					tbl.Name, cv.Name)
			}
		}

		if strings.HasPrefix(tbl.Name, "nx") ||
			strings.HasPrefix(tbl.Name, "tx") ||
			tbl.Name == "sys_config" ||
			tbl.Name == "modules" {

			tbl_name_new := ""
			if tbl.Name[:2] == "nx" {
				tbl_name_new = "hpn_" + tbl.Name[2:]
			} else if tbl.Name[:2] == "tx" {
				tbl_name_new = "hpt_" + tbl.Name[2:]
			} else {
				tbl_name_new = "hp_" + tbl.Name
			}

			hlog.Printf("warn", "store_init rename table %s to %s", tbl.Name, tbl_name_new)

			sql := ""

			switch DataOptions.Driver {

			case "lynkdb/mysqlgo":
				sql = fmt.Sprintf("RENAME TABLE %s TO %s", tbl.Name, tbl_name_new)

			case "lynkdb/pgsqlgo":
				sql = fmt.Sprintf("ALTER TABLE %s RENAME TO %s", tbl.Name, tbl_name_new)
			}

			if _, err := data.ExecRaw(sql); err != nil {
				return err
			}

			hlog.Printf("warn", "store_init rename table %s to %s, DONE", tbl.Name, tbl_name_new)
		}

	}

	return nil
}
