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

package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/htoml4g/htoml"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lynkdb/iomix/rdb/modeler"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/store"
	"github.com/hooto/hpress/websrv/web"
)

var (
	locker    sync.Mutex
	Modules   = map[string]*api.Spec{}
	mtmu      sync.RWMutex
	modThemes = map[string]map[string]string{}
)

func ModTheme(name string) map[string]string {
	mtmu.Lock()
	defer mtmu.Unlock()
	mt, ok := modThemes[name]
	if ok {
		return mt
	}
	return map[string]string{}
}

func ThemeConfigFetchString(name, key string, args ...string) string {
	tc := ModTheme(name)
	if v, ok := tc[key]; ok && len(v) > 0 {
		return v
	}
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

func SpecSet(spec *api.Spec) {

	locker.Lock()
	defer locker.Unlock()

	if strings.Contains(spec.SrvName, "/") {
		spec.SrvName, _ = api.SrvNameFilter(spec.SrvName)
	}

	Modules[spec.SrvName] = spec

}

func SpecGet(modname string) *api.Spec {
	for _, mod := range Modules {

		if mod.Meta.Name == modname {
			if mod.Status == 1 {
				return mod
			}
			break
		}
	}
	return nil
}

func SpecNodeModel(modname, modelName string) (*api.NodeModel, error) {

	for _, mod := range Modules {

		if mod.Meta.Name != modname {
			continue
		}

		for _, nodeModel := range mod.NodeModels {

			if modelName == nodeModel.Meta.Name {
				return nodeModel, nil
			}
		}
	}

	return &api.NodeModel{}, errors.New("Spec Not Found")
}

func SpecTermModel(modname, modelName string) (*api.TermModel, error) {

	for _, mod := range Modules {

		if mod.Meta.Name != modname {
			continue
		}

		for _, termModel := range mod.TermModels {

			if modelName == termModel.Meta.Name {
				return &termModel, nil
			}
		}
	}

	return &api.TermModel{}, errors.New("Spec Not Found")
}

func module_init() error {

	timenow := uint32(time.Now().Unix())

	if store.Data == nil {
		return errors.New("No RDB Connector Found")
	}

	//
	q := store.Data.NewQueryer().From("hp_modules").Limit(200)
	// q.Where.And("status", 1)
	if rs, err := store.Data.Query(q); err == nil {

		for _, v := range rs {

			var mod api.Spec

			if err := v.Field("body").JsonDecode(&mod); err == nil && mod.Meta.Name != "" {
				if mod.SrvName == "" || strings.Contains(mod.SrvName, "/") {
					mod.SrvName, _ = api.SrvNameFilter(v.Field("srvname").String())
				}

				// upgrade
				sync := false
				for j, v2 := range mod.NodeModels {

					v2.ModName = mod.Meta.Name
					v2.SrvName = mod.SrvName

					if ft := v2.Field("title"); ft == nil {
						v2.Fields = append([]api.FieldModel{
							{
								Name:   "title",
								Type:   "string",
								Length: "100",
								Title:  "Title",
							},
						}, v2.Fields...)
						mod.NodeModels[j] = v2
						sync = true

						if err := _instance_schema_sync(&mod); err != nil {
							hlog.Printf("error", err.Error())
							return err
						}
					}

					//
					table := api.NodeTableName(mod.Meta.Name, v2.Meta.Name)
					qs := store.Data.NewQueryer().
						Select("id,title,field_title").
						From(table).
						Limit(10000)

					if rs, err := store.Data.Query(qs); err == nil && len(rs) > 0 {

						num := 0

						for _, v3 := range rs {
							if len(v3.Field("field_title").String()) < 2 {

								// fmt.Println("id", v3.Field("id").String(), v3.Field("title").String())

								fr := store.Data.NewFilter()
								fr.And("id", v3.Field("id").String())

								set := map[string]interface{}{
									"field_title": v3.Field("title").String(),
								}
								store.Data.Update(table, set, fr)
								num++
							}
						}

						hlog.Printf("warn", "upgrade %s/%s %d/%d", mod.Meta.Name, table, num, len(rs))
					}
				}

				if sync {
					js, _ := json.Encode(mod, "  ")

					fr := store.Data.NewFilter()
					fr.And("name", mod.Meta.Name)

					set := map[string]interface{}{
						"body": string(js),
					}
					store.Data.Update("hp_modules", set, fr)
				}

				// if v.Field("status").String() != "1" {
				// 	continue
				// }

				Modules[mod.SrvName] = &mod
			} else {
				hlog.Printf("error", "Module.Init(%s) Failed", v.Field("name").String())
			}
		}
	}

	for _, v := range Config.ExpModuleInits {
		coreModules = append(coreModules, v)
	}

	//
	for _, modname := range coreModules {

		var spec api.Spec
		err := json.DecodeFile(fmt.Sprintf("%s/modules/%s/spec.json", Prefix, modname), &spec)
		if err != nil {
			return err
		}

		if !api.NewSpecVersion(spec.Meta.Version).Valid() {
			return fmt.Errorf("Invalid Version of %s", modname)
		}

		// upgrade
		sync := false
		for j, v2 := range spec.NodeModels {
			v2.ModName = modname
			v2.SrvName = spec.SrvName
			if ft := v2.Field("title"); ft == nil {
				v2.Fields = append([]api.FieldModel{
					{
						Name:   "title",
						Type:   "string",
						Length: "100",
						Title:  "Title",
					},
				}, v2.Fields...)
				spec.NodeModels[j] = v2
				sync = true
			}
		}
		if sync {
			spec.Meta.Version = api.NewSpecVersion(spec.Meta.Version).Add(0, 0, 1).String()
		}

		var instResVersion api.SpecVersion
		for _, mod := range Modules {
			if mod.Meta.Name == modname {
				instResVersion = api.SpecVersion(mod.Meta.Version)
				break
			}
		}

		if api.NewSpecVersion(spec.Meta.Version).Compare(&instResVersion) <= 0 {
			continue
		}

		if spec.SrvName == "" || strings.Contains(spec.SrvName, "/") {
			spec.SrvName, err = api.SrvNameFilter(spec.Meta.Name)
			if err != nil {
				return err
			}
		}

		//
		jsb, _ := json.Encode(spec, "  ")
		set := map[string]interface{}{
			"title":   spec.Title,
			"version": spec.Meta.Version,
			"updated": timenow,
			"body":    string(jsb),
		}

		q = store.Data.NewQueryer().From("hp_modules")
		q.Where().And("name", spec.Meta.Name)

		if _, err := store.Data.Fetch(q); err == nil {

			fr := store.Data.NewFilter()
			fr.And("name", spec.Meta.Name)

			store.Data.Update("hp_modules", set, fr)

		} else {

			set["name"] = spec.Meta.Name
			set["srvname"] = spec.SrvName
			set["created"] = timenow
			set["status"] = 1

			store.Data.Insert("hp_modules", set)
		}

		Modules[spec.SrvName] = &spec
	}

	//
	for _, mod := range Modules {
		if err := _instance_schema_sync(mod); err != nil {
			return err
		}
		SpecSrvRefresh(mod.SrvName)

		// upgrade
		json.EncodeToFile(mod, fmt.Sprintf("%s/modules/%s/spec.json", Prefix, mod.Meta.Name), "  ")
	}

	// Migrate legacy hash-keyed tables for modules that were not yet in
	// hp_modules at store.Setup time (e.g. exp modules loaded from disk or
	// exp_module_inits that are not persisted). Uses the now-complete in-memory
	// module list. Idempotent; a no-op on an already-migrated database.
	moduleNames := make([]string, 0, len(Modules))
	for _, mod := range Modules {
		if mod.Meta.Name != "" {
			moduleNames = append(moduleNames, mod.Meta.Name)
		}
	}
	if err := store.UpgradeModuleTableNaming(moduleNames); err != nil {
		return err
	}

	return nil
}

func SpecRefresh(modname string) {

	for srvname, spec := range Modules {
		if spec.Meta.Name == modname {
			SpecSrvRefresh(srvname)
			break
		}
	}
}

func SpecSrvRefresh(srvname string) {

	if strings.Contains(srvname, "/") {
		srvname, _ = api.SrvNameFilter(srvname)
	}

	spec, ok := Modules[srvname]
	if !ok {
		return
	}

	var theme map[string]string
	if err := htoml.Decode([]byte(spec.ThemeConfig), &theme); err == nil {
		mtmu.Lock()
		modThemes[spec.SrvName] = theme
		mtmu.Unlock()
	}

	for i, v := range spec.Router.Routes {
		spec.Router.Routes[i].Tree = strings.Split(strings.Trim(filepath.Clean(v.Path), "/"), "/")
	}

	web.Templates.Clean(spec.Meta.Name)
	web.Templates.Set(spec.Meta.Name,
		[]string{fmt.Sprintf("%s/modules/%s/views", Prefix, spec.Meta.Name)})
}

func _instance_schema_sync(spec *api.Spec) error {

	if store.Data == nil {
		return errors.New("No RDB Connector Found")
	}

	ds := &modeler.Schema{}

	// nodes
	for _, nodeModel := range spec.NodeModels {

		var tbl modeler.Table

		if err := json.Decode([]byte(dsTplNodeModels), &tbl); err != nil {
			continue
		}

		tbl.Name = api.NodeTableName(spec.Meta.Name, nodeModel.Meta.Name)

		if nodeModel.Extensions.AccessCounter {
			tbl.AddColumn(&modeler.Column{
				Name: "ext_access_counter",
				Type: "uint32",
			})
		}

		if nodeModel.Extensions.CommentPerEntry {
			tbl.AddColumn(&modeler.Column{
				Name:    "ext_comment_perentry",
				Type:    "uint8",
				Default: "1",
			})
		}

		if nodeModel.Extensions.Permalink != "" &&
			nodeModel.Extensions.Permalink != "off" {
			tbl.AddColumn(&modeler.Column{
				Name:   "ext_permalink_name",
				Type:   "string",
				Length: "100",
			})
			tbl.AddColumn(&modeler.Column{
				Name:   "ext_permalink_idx",
				Type:   "string",
				Length: "12",
			})
			tbl.AddIndex(&modeler.Index{
				Name: "ext_permalink_idx",
				Type: modeler.IndexTypeIndex,
				Cols: []string{"ext_permalink_idx"},
			})
		}

		if nodeModel.Extensions.NodeRefer != "" {
			tbl.AddColumn(&modeler.Column{
				Name:   "ext_node_refer",
				Type:   "string",
				Length: "16",
			})
			tbl.AddIndex(&modeler.Index{
				Name: "ext_node_refer",
				Type: modeler.IndexTypeIndex,
				Cols: []string{"ext_node_refer"},
			})
		}

		for _, field := range nodeModel.Fields {

			switch field.Type {

			case "string":

				if field.Name == "title" {
					field.Length = "100"
				}

				tbl.AddColumn(&modeler.Column{
					Name:   "field_" + field.Name,
					Type:   "string",
					Length: field.Length,
				})

				if attr := field.Attrs.Get("langs"); attr != nil && len(attr.String()) > 3 {
					tbl.AddColumn(&modeler.Column{
						Name: "field_" + field.Name + "_langs",
						Type: "string-text",
					})
				}

				switch field.IndexType {
				case modeler.IndexTypeUnique, modeler.IndexTypeIndex:
					tbl.AddIndex(&modeler.Index{
						Name: "field_" + field.Name,
						Type: field.IndexType,
						Cols: []string{"field_" + field.Name},
					})
				}

			case "text":

				tbl.AddColumn(&modeler.Column{
					Name: "field_" + field.Name,
					Type: "string-text",
				})

				tbl.AddColumn(&modeler.Column{
					Name:   "field_" + field.Name + "_attrs",
					Type:   "string",
					Length: "200",
				})

				if attr := field.Attrs.Get("langs"); attr != nil && len(attr.String()) > 3 {
					tbl.AddColumn(&modeler.Column{
						Name: "field_" + field.Name + "_langs",
						Type: "string-text",
					})
				}

			case "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":

				tbl.AddColumn(&modeler.Column{
					Name: "field_" + field.Name,
					Type: field.Type,
				})

			}
		}

		for _, term := range nodeModel.Terms {

			switch term.Type {

			case api.TermTag:

				tbl.AddColumn(&modeler.Column{
					Name:   "term_" + term.Meta.Name,
					Type:   "string",
					Length: "200",
				})

				// tbl.AddColumn(&modeler.Column{
				// 	Name: "term_" + term.Meta.Name + "_body",
				// 	Type: "string-text",
				// })

				tbl.AddColumn(&modeler.Column{
					Name:   "term_" + term.Meta.Name + "_idx",
					Type:   "string",
					Length: "100",
				})

				tbl.AddIndex(&modeler.Index{
					Name: "term_" + term.Meta.Name + "_idx",
					Type: modeler.IndexTypeIndex,
					Cols: []string{"term_" + term.Meta.Name + "_idx"},
				})

			case api.TermTaxonomy:

				tbl.AddColumn(&modeler.Column{
					Name: "term_" + term.Meta.Name,
					Type: "uint32",
				})

				tbl.AddIndex(&modeler.Index{
					Name: "term_" + term.Meta.Name,
					Type: modeler.IndexTypeIndex,
					Cols: []string{"term_" + term.Meta.Name},
				})
			}

		}

		ds.Tables = append(ds.Tables, &tbl)
	}

	// terms
	for _, termModel := range spec.TermModels {

		var tbl modeler.Table

		if err := json.Decode([]byte(dsTplTermModels), &tbl); err != nil {
			continue
		}

		tbl.Name = api.TermTableName(spec.Meta.Name, termModel.Meta.Name)

		switch termModel.Type {

		case api.TermTag:

			tbl.AddColumn(&modeler.Column{
				Name:   "uid",
				Type:   "string",
				Length: "16",
			})

			tbl.AddIndex(&modeler.Index{
				Name: "uid",
				Type: modeler.IndexTypeUnique,
				Cols: []string{"uid"},
			})

		case api.TermTaxonomy:

			tbl.AddColumn(&modeler.Column{
				Name: "pid",
				Type: "uint32",
			})

			tbl.AddIndex(&modeler.Index{
				Name: "pid",
				Type: modeler.IndexTypeIndex,
				Cols: []string{"pid"},
			})

			tbl.AddColumn(&modeler.Column{
				Name: "weight",
				Type: "int16",
			})

			tbl.AddIndex(&modeler.Index{
				Name: "weight",
				Type: modeler.IndexTypeIndex,
				Cols: []string{"weight"},
			})

		default:
			continue
		}

		ds.Tables = append(ds.Tables, &tbl)
	}

	// sync
	dm, err := store.Data.Modeler()
	if err != nil {
		return err
	}

	// SchemaSync is intentionally disabled during the module table-naming
	// migration stabilization window: hpn_/hpt_ table structures must not be
	// created or altered while the rename upgrade is settling. Tables are
	// expected to already exist (from the rename upgrade or a prior run).
	// Re-enable the call below once the upgrade is stable.
	//
	// err = dm.SchemaSync(ds)
	// if err != nil {
	// 	return err
	// }
	_ = dm
	_ = ds

	for _, termModel := range spec.TermModels {

		switch termModel.Type {

		case api.TermTaxonomy:

			tblName := api.TermTableName(spec.Meta.Name, termModel.Meta.Name)
			rs, _ := store.Data.Fetch(store.Data.NewQueryer().From(tblName))
			if rs.NotFound() {
				store.Data.Insert(tblName, map[string]interface{}{
					"pid":     0,
					"title":   "Default",
					"status":  1,
					"weight":  0,
					"created": time.Now().Unix(),
					"userid":  "",
				})
			}
		}
	}

	return nil
}
