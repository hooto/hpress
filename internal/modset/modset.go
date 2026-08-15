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

package modset

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lessos/lessgo/utilx"
	"github.com/lynkdb/iomix/rdb/modeler"

	"github.com/hooto/hpress/internal/config"
	"github.com/hooto/hpress/internal/hpapi"
	"github.com/hooto/hpress/internal/store"
)

var (
	modNamePattern        = regexp.MustCompile("^[0-9a-z/]{3,30}$")
	modelNamePattern      = regexp.MustCompile("^[a-z]{1,1}[0-9a-z_]{1,20}$")
	nodeFieldNamePattern  = regexp.MustCompile("^[a-z]{1,1}[0-9a-z_]{1,20}$")
	routePathPattern      = regexp.MustCompile("^[0-9a-zA-Z_/\\-:]{1,50}$")
	routeParamNamePattern = regexp.MustCompile("^[a-z]{1,1}[0-9a-zA-Z_]{0,29}$")
)

func ModNameFilter(name string) (string, error) {

	name = strings.Trim(filepath.Clean(strings.ToLower(name)), "/")

	if mat := modNamePattern.MatchString(name); !mat {
		return "", fmt.Errorf("Invalid Module Name (%s)", name)
	}

	return name, nil
}

func ModelNameFilter(name string) (string, error) {

	name = strings.TrimSpace(strings.ToLower(name))

	if mat := modelNamePattern.MatchString(name); !mat {
		return "", fmt.Errorf("Invalid Model Name (%s)", name)
	}

	return name, nil
}

func RoutePathFilter(name string) (string, error) {

	name = strings.TrimSpace(name)

	if mat := routePathPattern.MatchString(name); !mat {
		return "", fmt.Errorf("Invalid Route Path (%s)", name)
	}

	return name, nil
}

func SpecFetch(modname string) (hpapi.Spec, error) {
	var entry hpapi.Spec
	file := fmt.Sprintf("%s/modules/%s/spec.json", config.Prefix, modname)
	err := json.DecodeFile(file, &entry)
	return entry, err
}

func SpecInfoSet(entry hpapi.Spec) error {

	if entry.Meta.Name == "" {
		return errors.New("Name Not Found")
	}

	if entry.Title == "" {
		return errors.New("Title Not Found")
	}

	prev, err := SpecFetch(entry.Meta.Name)
	if err != nil {
		return err
	}

	if entry.Meta.Name == "core/general" {
		entry.Status = 1
	} else if entry.Status != 1 {
		entry.Status = 0
	}

	if prev.Title != entry.Title ||
		prev.SrvName != entry.SrvName ||
		prev.Status != entry.Status ||
		prev.ThemeConfig != entry.ThemeConfig {

		prev.Meta.Version = hpapi.NewSpecVersion(prev.Meta.Version).Add(0, 0, 1).String()
		prev.Title = entry.Title
		prev.SrvName, err = hpapi.SrvNameFilter(entry.SrvName)
		prev.Status = entry.Status
		prev.Meta.Updated = types.MetaTimeNow()
		prev.ThemeConfig = entry.ThemeConfig

		if err := specConfigFileSync(prev); err != nil {
			return err
		}
	}

	return err
}

func SpecTermSet(modname string, entry hpapi.TermModel) error {

	if modname == "" {
		return errors.New("modname Not Found")
	}

	prev, err := SpecFetch(modname)
	if err != nil {
		return err
	}

	sync, found := false, false
	for i, termModel := range prev.TermModels {

		if termModel.Meta.Name == entry.Meta.Name {

			found = true

			if prev.Title != entry.Title {
				prev.TermModels[i].Title = entry.Title
				sync = true
			}
		}
	}

	if !found {
		entry.ModName = ""
		prev.TermModels = append(prev.TermModels, entry)
		sync = true
	}

	if sync {

		prev.Meta.Version = hpapi.NewSpecVersion(prev.Meta.Version).Add(0, 0, 1).String()
		prev.Meta.Updated = types.MetaTimeNow()

		if err := specConfigFileSync(prev); err != nil {
			return err
		}
	}

	return err
}

func termListEqual(ls1, ls2 []hpapi.TermModel) bool {

	if len(ls1) != len(ls2) {
		return false
	}

	for _, kv1 := range ls1 {

		found := false

		for _, kv2 := range ls2 {

			if kv1.Meta.Name == kv2.Meta.Name {

				if kv1.Type != kv2.Type ||
					kv1.Title != kv2.Title {
					return false
				}

				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func SpecNodeSet(modname string, entry *hpapi.NodeModel) error {

	if modname == "" {
		return errors.New("modname Not Found")
	}

	for i, field := range entry.Fields {

		if mat := nodeFieldNamePattern.MatchString(field.Name); !mat {
			return fmt.Errorf("Invalid Field Name (%s)", field.Name)
		}

		if field.Title == "" {
			entry.Fields[i].Title = field.Name
		}

		if !utilx.ArrayContain(field.Type, hpapi.NodeFieldTypes) {
			return fmt.Errorf("Invalid Field Type (%s:%s)", field.Name, field.Type)
		}

		if field.IndexType != 0 && field.IndexType != 1 && field.IndexType != 2 {
			return fmt.Errorf("Invalid Field Index Type (%s:%s)", field.Name, field.IndexType)
		}

		if field.Type == "string" {

			length, _ := strconv.Atoi(field.Length)

			if length < 1 {
				entry.Fields[i].Length = "10"
			} else if length > 200 {
				entry.Fields[i].Length = "200"
			}
		}

		attrDels := []string{}

		for j, attr := range field.Attrs {

			if mat := nodeFieldNamePattern.MatchString(attr.Key); !mat {
				return fmt.Errorf("Invalid Field Attribute Key (%s)", attr.Key)
			}

			if attr.Key == "langs" {
				if field.Type == "string" || field.Type == "text" {
					entry.Fields[i].Attrs[j].Value = hpapi.LangsStringFilter(attr.Value)
				} else {
					entry.Fields[i].Attrs[j].Value = ""
				}
			}

			if entry.Fields[i].Attrs[j].Value == "" {
				attrDels = append(attrDels, attr.Key)
			}
		}

		for _, v := range attrDels {
			entry.Fields[i].Attrs.Del(v)
		}
	}

	prev, err := SpecFetch(modname)
	if err != nil {
		return err
	}

	sync, found := false, false
	for i, nodeModel := range prev.NodeModels {

		if nodeModel.Meta.Name == entry.Meta.Name {

			found = true

			if nodeModel.Title != entry.Title {
				prev.NodeModels[i].Title = entry.Title
				sync = true
			}

			if nodeModel.Extensions.AccessCounter != entry.Extensions.AccessCounter {
				prev.NodeModels[i].Extensions.AccessCounter = entry.Extensions.AccessCounter
				sync = true
			}

			if nodeModel.Extensions.CommentEnable != entry.Extensions.CommentEnable {
				prev.NodeModels[i].Extensions.CommentEnable = entry.Extensions.CommentEnable
				sync = true
			}

			if nodeModel.Extensions.CommentPerEntry != entry.Extensions.CommentPerEntry {
				prev.NodeModels[i].Extensions.CommentPerEntry = entry.Extensions.CommentPerEntry
				sync = true
			}

			if nodeModel.Extensions.Permalink != entry.Extensions.Permalink {
				prev.NodeModels[i].Extensions.Permalink = entry.Extensions.Permalink
				sync = true
			}

			if nodeModel.Extensions.NodeRefer != entry.Extensions.NodeRefer &&
				entry.Extensions.NodeRefer != "" {
				if entry.Extensions.NodeRefer == nodeModel.Meta.Name {
					return errors.New("Invalid Node Refer Value")
				}
				if referNodeModel := prev.NodeModelGet(entry.Extensions.NodeRefer); referNodeModel == nil {
					return errors.New("Node Refer Not Found")
				}

				prev.NodeModels[i].Extensions.NodeRefer = entry.Extensions.NodeRefer
				sync = true
			}

			if nodeModel.Extensions.TextSearch != entry.Extensions.TextSearch {
				prev.NodeModels[i].Extensions.TextSearch = entry.Extensions.TextSearch
				sync = true
			}

			if len(nodeModel.Fields) != len(entry.Fields) && len(entry.Fields) > 0 {

				prev.NodeModels[i].Fields = entry.Fields
				sync = true

			} else {

				for _, prevField := range nodeModel.Fields {

					fieldSync := true

					for _, curField := range entry.Fields {

						if curField.Name == prevField.Name {

							if curField.Title == prevField.Title &&
								curField.Type == prevField.Type &&
								curField.IndexType == prevField.IndexType &&
								curField.Length == prevField.Length &&
								curField.Attrs.Equal(prevField.Attrs) {

								fieldSync = false
							}

							break
						}
					}

					if fieldSync && len(entry.Fields) > 0 {

						sync = true
						prev.NodeModels[i].Fields = entry.Fields

						break
					}
				}
			}

			if !termListEqual(nodeModel.Terms, entry.Terms) {

				prev.NodeModels[i].Terms = entry.Terms
				sync = true

				for _, sterm := range entry.Terms {

					prevTermFound := false

					for i, pterm := range prev.TermModels {

						if pterm.Meta.Name == sterm.Meta.Name {

							prevTermFound = true
							prev.TermModels[i] = sterm
							break
						}
					}

					if !prevTermFound {
						prev.TermModels = append(prev.TermModels, sterm)
					}
				}
			}
		}
	}

	if !found {
		entry.ModName = ""
		prev.NodeModels = append(prev.NodeModels, entry)

		sync = true
	}

	for i, v1 := range prev.NodeModels {

		nodeSubRefer := ""

		if v1.Extensions.NodeRefer == "" {
			for _, v2 := range prev.NodeModels {
				if v2.Meta.Name != v1.Meta.Name &&
					v2.Extensions.NodeRefer == v1.Meta.Name {
					nodeSubRefer = v2.Meta.Name
					break
				}
			}
		}

		if nodeSubRefer != v1.Extensions.NodeSubRefer {
			prev.NodeModels[i].Extensions.NodeSubRefer = nodeSubRefer
			sync = true
		}
	}

	if sync {

		prev.Meta.Version = hpapi.NewSpecVersion(prev.Meta.Version).Add(0, 0, 1).String()
		prev.Meta.Updated = types.MetaTimeNow()

		if err := specConfigFileSync(prev); err != nil {
			return err
		}
	}

	return err
}

func SpecActionSet(modname string, entry hpapi.Action) error {

	if modname == "" {
		return errors.New("modname Not Found")
	}

	if mat := modelNamePattern.MatchString(entry.Name); !mat {
		return fmt.Errorf("Invalid Action Name (%s)", entry.Name)
	}

	prev, err := SpecFetch(modname)
	if err != nil {
		return err
	}

	for i, dentry := range entry.Datax {

		if mat := modelNamePattern.MatchString(dentry.Name); !mat {
			return fmt.Errorf("Invalid Datax Name (%s)", dentry.Name)
		}

		types := strings.Split(dentry.Type, ".")
		if len(types) != 2 {
			return fmt.Errorf("Invalid Datax Type (%s:%s)", dentry.Name, dentry.Type)
		}

		if !utilx.ArrayContain(types[1], []string{"list", "entry"}) {
			return fmt.Errorf("Invalid Datax Type (%s:%s)", dentry.Name, dentry.Type)
		}

		if dentry.CacheTTL > (86400 * 30000) {
			entry.Datax[i].CacheTTL = 86400 * 30000
		}

		switch types[0] {

		case "node":

			if dentry.Query.Limit < 1 {
				entry.Datax[i].Query.Limit = 1
			} else if dentry.Query.Limit > 10000 {
				entry.Datax[i].Query.Limit = 10000
			}

			tableFound := false
			for _, nodeModel := range prev.NodeModels {

				if nodeModel.Meta.Name == dentry.Query.Table {
					tableFound = true
					break
				}
			}

			if !tableFound {
				return fmt.Errorf("Query Table Not Found (%s)", dentry.Query.Table)
			}

		case "term":

			tableFound := false
			for _, termModel := range prev.TermModels {

				if termModel.Meta.Name == dentry.Query.Table {
					tableFound = true
					break
				}
			}

			if !tableFound {
				return fmt.Errorf("Query Table Not Found (%s)", dentry.Query.Table)
			}

		default:
			return fmt.Errorf("Invalid Datax Type (%s:%s)", dentry.Name, dentry.Type)
		}
	}

	sync, found := false, false
	for i, action := range prev.Actions {

		if action.Name == entry.Name {

			found = true

			if len(action.Datax) != len(entry.Datax) && len(entry.Datax) > 0 {

				prev.Actions[i].Datax = entry.Datax
				sync = true

			} else {

				for _, prevDatax := range action.Datax {

					dataxSync := true

					for _, curField := range entry.Datax {

						if curField.Name == prevDatax.Name {

							if curField.Type == prevDatax.Type &&
								curField.Pager == prevDatax.Pager &&
								curField.CacheTTL == prevDatax.CacheTTL &&
								curField.Query.Table == prevDatax.Query.Table &&
								curField.Query.Limit == prevDatax.Query.Limit &&
								curField.Query.Order == prevDatax.Query.Order {

								dataxSync = false
							}

							break
						}
					}

					if dataxSync && len(entry.Datax) > 0 {

						sync = true
						prev.Actions[i].Datax = entry.Datax

						break
					}
				}
			}

		}
	}

	if !found {
		entry.ModName = ""
		prev.Actions = append(prev.Actions, entry)

		sync = true
	}

	if sync {

		prev.Meta.Version = hpapi.NewSpecVersion(prev.Meta.Version).Add(0, 0, 1).String()
		prev.Meta.Updated = types.MetaTimeNow()

		if err := specConfigFileSync(prev); err != nil {
			return err
		}
	}

	return err
}

func SpecActionDel(modname string, entry hpapi.Action) error {

	if modname == "" {
		return errors.New("modname Not Found")
	}

	if mat := modelNamePattern.MatchString(entry.Name); !mat {
		return fmt.Errorf("Invalid Action Name (%s)", entry.Name)
	}

	prev, err := SpecFetch(modname)
	if err != nil {
		return err
	}

	for i, action := range prev.Actions {

		if action.Name != entry.Name {
			continue
		}

		prev.Actions = append(prev.Actions[:i], prev.Actions[i+1:]...)
		prev.Meta.Version = hpapi.NewSpecVersion(prev.Meta.Version).Add(0, 0, 1).String()
		prev.Meta.Updated = types.MetaTimeNow()

		if err := specConfigFileSync(prev); err != nil {
			return err
		}
	}

	return nil
}

func routeParamsEqual(a1, a2 map[string]string) bool {

	if len(a1) != len(a2) {
		return false
	}

	for k, v := range a1 {

		found := false

		for k2, v2 := range a2 {

			if k == k2 {

				if v != v2 {
					return false
				}

				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func SpecRouteSet(modname string, entry hpapi.Route) error {

	if modname == "" {
		return errors.New("modname Not Found")
	}

	var err error

	if entry.Path, err = RoutePathFilter(entry.Path); err != nil {
		return fmt.Errorf("Invalid Action Path (%s)", entry.Path)
	}

	for k := range entry.Params {

		if mat := routeParamNamePattern.MatchString(k); !mat {
			return fmt.Errorf("Invalid Param Name (%s)", k)
		}
	}

	prev, err := SpecFetch(modname)
	if err != nil {
		return err
	}

	sync, found, def := true, false, false
	for i, prevRoute := range prev.Router.Routes {

		if prevRoute.Path == entry.Path {

			found = true

			if entry.DataAction == prevRoute.DataAction &&
				entry.Template == prevRoute.Template &&
				entry.Default == prevRoute.Default &&
				routeParamsEqual(entry.Params, prevRoute.Params) {

				sync = false
			} else {
				entry.ModName = ""
				prev.Router.Routes[i] = entry
			}

			if entry.Default {
				def = true
			}

			break
		}
	}

	if !found {
		entry.ModName = ""
		prev.Router.Routes = append(prev.Router.Routes, entry)

		sync = true
	}

	if def {
		for i, prevRoute := range prev.Router.Routes {
			if prevRoute.Default && prevRoute.Path != entry.Path {
				prev.Router.Routes[i].Default = false
				sync = true
			}
		}
	}

	sort.Slice(prev.Router.Routes, func(i, j int) bool {
		if strings.Compare(prev.Router.Routes[i].Path, prev.Router.Routes[j].Path) < 0 {
			return false
		}
		return true
	})

	if sync {

		prev.Meta.Version = hpapi.NewSpecVersion(prev.Meta.Version).Add(0, 0, 1).String()
		prev.Meta.Updated = types.MetaTimeNow()

		if err := specConfigFileSync(prev); err != nil {
			return err
		}
	}

	return err
}

func SpecRouteDel(modname string, entry hpapi.Route) error {

	if modname == "" {
		return errors.New("modname Not Found")
	}

	var err error

	if entry.Path, err = RoutePathFilter(entry.Path); err != nil {
		return fmt.Errorf("Invalid Action Path (%s)", entry.Path)
	}

	prev, err := SpecFetch(modname)
	if err != nil {
		return err
	}

	for i, prevRoute := range prev.Router.Routes {

		if prevRoute.Path != entry.Path {
			continue
		}

		prev.Router.Routes = append(prev.Router.Routes[:i], prev.Router.Routes[i+1:]...)

		prev.Meta.Version = hpapi.NewSpecVersion(prev.Meta.Version).Add(0, 0, 1).String()
		prev.Meta.Updated = types.MetaTimeNow()

		if err := specConfigFileSync(prev); err != nil {
			return err
		}
	}

	return nil
}

func specConfigFileSync(entry hpapi.Spec) error {

	entry.Meta.Created = 0
	entry.Meta.Updated = 0

	//
	file := fmt.Sprintf("%s/modules/%s/spec.json", config.Prefix, entry.Meta.Name)

	return json.EncodeToFile(entry, file, "  ")
}

func SpecSchemaSync(spec hpapi.Spec) error {

	var (
		ds      = &modeler.Schema{}
		timenow = uint32(time.Now().Unix())
		err     error
	)

	//
	if spec.SrvName == "" || strings.Contains(spec.SrvName, "/") {
		spec.SrvName, err = hpapi.SrvNameFilter(spec.Meta.Name)
		if err != nil {
			return err
		}
	}

	// TODO
	config.SpecSet(&spec)
	config.SpecSrvRefresh(spec.SrvName)

	jsb, _ := json.Encode(spec, "  ")
	set := map[string]interface{}{
		"srvname": spec.SrvName,
		"status":  1,
		"title":   spec.Title,
		"version": spec.Meta.Version,
		"updated": timenow,
		"body":    string(jsb),
	}

	q := store.Data.NewQueryer().From("hp_modules")
	q.Where().And("name", spec.Meta.Name)

	if _, err := store.Data.Fetch(q); err == nil {

		fr := store.Data.NewFilter()
		fr.And("name", spec.Meta.Name)

		_, err = store.Data.Update("hp_modules", set, fr)

	} else {

		set["name"] = spec.Meta.Name
		set["created"] = timenow

		_, err = store.Data.Insert("hp_modules", set)
	}

	//
	for _, nodeModel := range spec.NodeModels {

		var tbl modeler.Table

		if err := json.Decode([]byte(dsTplNodeModels), &tbl); err != nil {
			continue
		}

		tbl.Name = hpapi.NodeTableName(spec.Meta.Name, nodeModel.Meta.Name)

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

			case hpapi.TermTag:

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

			case hpapi.TermTaxonomy:

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

	for _, termModel := range spec.TermModels {

		var tbl modeler.Table

		if err := json.Decode([]byte(dsTplTermModels), &tbl); err != nil {
			continue
		}

		tbl.Name = hpapi.TermTableName(spec.Meta.Name, termModel.Meta.Name)

		switch termModel.Type {

		case hpapi.TermTag:

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

		case hpapi.TermTaxonomy:

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

	//
	ms, err := store.Data.Modeler()
	if err != nil {
		return err
	}

	// SchemaSync is intentionally disabled during the module table-naming
	// migration stabilization window: hpn_/hpt_ table structures must not be
	// created or altered while the rename upgrade is settling. Tables are
	// expected to already exist (from the rename upgrade or a prior run).
	// Re-enable the call below once the upgrade is stable.
	//
	// err = ms.SchemaSync(ds)
	// if err != nil {
	// 	return err
	// }
	_ = ms
	_ = ds

	for _, termModel := range spec.TermModels {

		switch termModel.Type {

		case hpapi.TermTaxonomy:

			tblName := hpapi.TermTableName(spec.Meta.Name, termModel.Meta.Name)
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
