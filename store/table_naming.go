// Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/hooto/hpress/api"
	"github.com/lynkdb/iomix/rdb"
)

// legacyModuleTableRe matches the legacy per-module table scheme, whose middle
// segment was the first 12 hex chars of md5(moduleName):
//
//	hpn_<12hex>_<model>
//	hpt_<12hex>_<model>
//
// Tables already on the new scheme (hpn_core_blog_post) do not match, because
// the module-key segment is not 12 lowercase hex chars, so the upgrade is
// naturally idempotent.
var legacyModuleTableRe = regexp.MustCompile(`^(hpn|hpt)_([0-9a-f]{12})_(.+)$`)

// legacyModuleHash returns the legacy 12-hex-char md5 prefix a module name was
// keyed by. It is equivalent to both utils.StringEncode16(name, 12) and
// idhash.HashToHexString([]byte(name), 12) used across the older code base.
func legacyModuleHash(modName string) string {
	sum := md5.Sum([]byte(modName))
	return hex.EncodeToString(sum[:6])
}

// moduleHashToKey builds a reverse lookup from the legacy 12-hex hash to the
// new human-readable module key for every supplied module name.
func moduleHashToKey(moduleNames []string) map[string]string {
	m := make(map[string]string, len(moduleNames))
	for _, name := range moduleNames {
		if name == "" {
			continue
		}
		m[legacyModuleHash(name)] = api.ModuleTableKey(name)
	}
	return m
}

// parseLegacyModuleTable inspects a physical table name and, if it follows the
// legacy hash-keyed scheme whose hash is present in hashToKey, returns the
// pieces needed to build the new name: the table prefix (hpn/hpt), the matched
// module key, and the trailing model name. ok is false for non-legacy tables,
// already-renamed tables, or hashes with no known module.
func parseLegacyModuleTable(tblName string, hashToKey map[string]string) (prefix, key, model string, ok bool) {
	m := legacyModuleTableRe.FindStringSubmatch(tblName)
	if m == nil {
		return "", "", "", false
	}
	key, found := hashToKey[m[2]]
	if !found {
		return "", "", "", false
	}
	return m[1], key, m[3], true
}

// db_upgrade_module_table_naming is the startup entry point: it reads the
// installed module names from hp_modules and runs the rename pass. Extra
// modules known only in memory (exp modules not yet persisted) are covered by
// UpgradeModuleTableNaming, called from config after module init.
func db_upgrade_module_table_naming(data rdb.Connector) error {

	// On a fresh database hp_modules does not exist yet, but there are no
	// legacy tables either, so a query error is harmless.
	names, err := data.QueryRaw("SELECT name FROM hp_modules")
	if err != nil {
		slog.Warn(fmt.Sprintf("table naming upgrade: skip (cannot read hp_modules: %s)", err.Error()))
		return nil
	}

	moduleNames := make([]string, 0, len(names))
	for _, r := range names {
		if n := r.Field("name").String(); n != "" {
			moduleNames = append(moduleNames, n)
		}
	}

	return renameModuleTables(data, moduleNames)
}

// UpgradeModuleTableNaming renames legacy hash-keyed tables for the given
// module names. Callers that hold a more complete module list than hp_modules
// (config.Modules includes exp modules loaded from disk/exp_module_inits that
// may not be persisted yet) run this idempotent pass after startup wiring so
// those modules' legacy tables are migrated too.
func UpgradeModuleTableNaming(moduleNames []string) error {
	return renameModuleTables(Data, moduleNames)
}

// renameModuleTables is the core of the upgrade. It is deliberately a
// RENAME-ONLY migration: for every legacy hpn_/hpt_<hash>_<model> table whose
// hash maps to one of moduleNames it renames the table to the new module-name
// scheme (hpn_<ns>_<mod>_<model>).
//
// It performs no DROP and no row copy/INSERT. If a target table already exists
// the legacy table is left untouched and logged, so the pass can be run any
// number of times without any risk of altering or losing data. md5 is not
// reversible, so tables whose hash matches no supplied module are skipped too.
//
// Idempotent: once a legacy table is renamed it no longer matches the legacy
// pattern, so subsequent runs are no-ops.
func renameModuleTables(data rdb.Connector, moduleNames []string) error {

	hashToKey := moduleHashToKey(moduleNames)
	if len(hashToKey) == 0 {
		return nil
	}

	mdr, err := data.Modeler()
	if err != nil {
		return fmt.Errorf("table naming upgrade: modeler: %w", err)
	}

	tbls, err := mdr.TableDump()
	if err != nil {
		return fmt.Errorf("table naming upgrade: table dump: %w", err)
	}

	for _, tbl := range tbls {

		prefix, key, model, ok := parseLegacyModuleTable(tbl.Name, hashToKey)
		if !ok {
			continue
		}

		newName := prefix + "_" + key + "_" + model

		// Rename-only: never touch an existing target. If one is already
		// present the legacy table is left in place for manual resolution, so
		// no data can be lost or modified regardless of how often this runs.
		//
		// NOTE: do not use modeler.TableExist here. Its pgsqlgo driver runs
		// "SELECT count(*) ..." and then returns len(rows) > 0, which is always
		// true because count(*) always yields exactly one row -- so every name
		// reports as existing and the whole rename pass would silently skip.
		// See tableExists below for a correct check.
		exists, existErr := tableExists(data, string(DataOptions.Driver), newName)
		if existErr != nil {
			return fmt.Errorf("table naming upgrade: exist check %s: %w", newName, existErr)
		}
		if exists {
			slog.Warn(fmt.Sprintf("table naming upgrade: target %s already exists, skip %s (rename-only)",
				newName, tbl.Name))
			continue
		}

		if err := renameTable(data, string(DataOptions.Driver), tbl.Name, newName); err != nil {
			return fmt.Errorf("table naming upgrade: rename %s -> %s: %w", tbl.Name, newName, err)
		}

		slog.Warn(fmt.Sprintf("table naming upgrade: renamed %s -> %s", tbl.Name, newName))
	}

	return nil
}

// tableExists reports whether a base table named tblName exists in the
// connected database, for either supported driver.
//
// It exists because modeler.TableExist is broken on pgsqlgo: it issues
// "SELECT count(*) FROM information_schema.tables WHERE table_name = ?" and then
// tests len(rows) > 0. count(*) is an aggregate without GROUP BY, so it returns
// exactly one row even when the count is 0 -- hence TableExist is true for any
// name. Querying for an actual matching row (LIMIT 1) instead makes the result
// correct: a row is returned only when the table really exists.
func tableExists(data rdb.Connector, driver, tblName string) (bool, error) {
	var q string
	switch driver {
	case "lynkdb/mysqlgo":
		q = "SELECT 1 FROM INFORMATION_SCHEMA.tables WHERE table_schema = DATABASE() AND table_name = ? LIMIT 1"
	default: // lynkdb/pgsqlgo
		q = "SELECT 1 FROM INFORMATION_SCHEMA.tables WHERE table_schema = 'public' AND table_name = ? LIMIT 1"
	}
	rs, err := data.QueryRaw(q, tblName)
	if err != nil {
		return false, err
	}
	return len(rs) > 0, nil
}

// renameTable renames old to new using the driver-specific statement.
func renameTable(data rdb.Connector, driver, oldName, newName string) error {
	var sql string
	switch driver {
	case "lynkdb/mysqlgo":
		sql = fmt.Sprintf("RENAME TABLE %s TO %s", oldName, newName)
	default: // lynkdb/pgsqlgo
		sql = fmt.Sprintf("ALTER TABLE %s RENAME TO %s", oldName, newName)
	}
	_, err := data.ExecRaw(sql)
	return err
}
