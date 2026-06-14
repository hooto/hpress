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
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/hooto/hcaptcha/captcha4g"
	"github.com/hooto/hflag4g/hflag"
	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/htoml4g/htoml"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/connect"
	"github.com/lynkdb/kvgo/v2/pkg/storage"
	"github.com/sysinner/incore/v2/pkg/inconf"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/store"
)

var (
	Prefix         string
	Config         ConfigCommon
	AppName        = "hooto-press"
	Version        = "0.10.0"
	Release        = "1"
	SysVersionSign = ""
	CaptchaConfig  = captcha4g.DefaultConfig

	User = &user.User{
		Uid:      "2048",
		Gid:      "2048",
		Username: "action",
		HomeDir:  "/home/action",
	}

	SysConfigList          = api.SysConfigList{}
	inited                 = false
	RouterBasepathDefault  = "/"
	RouterBasepathDefaults = []string{}
	Languages              = []*api.LangEntry{}
)

type ConfigCommon struct {
	UrlBasePath    string                   `json:"url_base_path,omitempty" toml:"url_base_path,omitempty"`
	ModuleDir      string                   `json:"module_dir,omitempty" toml:"module_dir,omitempty"`
	InstanceID     string                   `json:"instance_id" toml:"instance_id"`
	IamAuth        *iamserver.AppAuthConfig `json:"iam_auth" toml:"iam_auth"`
	AppTitle       string                   `json:"app_title,omitempty" toml:"app_title,omitempty"`
	HttpPort       uint16                   `json:"http_port" toml:"http_port"`
	IoConnectors   connect.MultiConnOptions `json:"io_connectors" toml:"io_connectors"`
	DataCache      *storage.Options         `json:"data_cache" toml:"data_cache"`
	RunMode        string                   `json:"run_mode,omitempty" toml:"run_mode,omitempty"`
	ExtUpDatabases connect.MultiConnOptions `json:"ext_up_databases,omitempty" toml:"ext_up_databases,omitempty"`
	ExpModuleInits []string                 `json:"exp_module_inits,omitempty" toml:"exp_module_inits,omitempty"`
	ExpGdocPaths   []string                 `json:"exp_gdoc_paths,omitempty" toml:"exp_gdoc_paths,omitempty"`
}

func init() {

	SysConfigList.Kind = "SysConfigList"

	//
	SysConfigList.Insert(api.SysConfig{
		"frontend_header_site_name", "Site Name",
		"Site's Name", "",
	})
	SysConfigList.Insert(api.SysConfig{
		"frontend_header_site_logo_url", "",
		"", "",
	})
	SysConfigList.Insert(api.SysConfig{
		"frontend_header_site_icon_url", "",
		"", "",
	})

	SysConfigList.Insert(api.SysConfig{
		"frontend_footer_copyright", fmt.Sprintf("© 2015~%d hooto.com", time.Now().Year()),
		"", "text",
	})

	//
	SysConfigList.Insert(api.SysConfig{
		"frontend_html_head_subtitle", "HP",
		"Sub Title for HTML Head Title", "",
	})
	SysConfigList.Insert(api.SysConfig{
		"frontend_html_head_meta_keywords", "",
		"Meta Keywords in HTML Head for Search engine optimization", "",
	})
	SysConfigList.Insert(api.SysConfig{
		"frontend_html_head_meta_description", "",
		"Meta Description in HTML Head for Search engine optimization", "",
	})

	SysConfigList.Insert(api.SysConfig{
		"frontend_html_footer", "",
		"Raw HTML Text for custom page footer", "text",
	})

	SysConfigList.Insert(api.SysConfig{
		"frontend_footer_analytics_scripts", "",
		"Embeded analytics scripts, ex. Google Analytics or Piwik ...", "text",
	})

	SysConfigList.Insert(api.SysConfig{
		"frontend_languages", "",
		"Multi languages support list", "",
	})

	SysConfigList.Insert(api.SysConfig{
		"storage_service_endpoint", "/hp/s2/deft",
		"Storage Service Endpoint", "",
	})

	//
	SysConfigList.Insert(api.SysConfig{
		"http_h_ac_allow_origin", "",
		"HTTP Access-Control-Allow-Origin", "",
	})

	SysConfigList.Insert(api.SysConfig{
		"router_basepath_default", "",
		"Default basepath of router", "",
	})

	go func() {
		for {
			time.Sleep(60e9)
			if err := syncInnerStackConfig(); err != nil {
				hlog.Printf("error", "syncInnerStackConfig err: %s", err.Error())
			}
		}
	}()
}

func Setup() error {

	if inited {
		return nil
	}

	var err error

	prefix := hflag.Value("prefix").String()

	if prefix == "" {
		if prefix, err = filepath.Abs(filepath.Dir(os.Args[0]) + "/.."); err != nil {
			prefix = "/opt/hooto/press"
		}
	}

	Prefix = filepath.Clean(prefix)
	Config.ModuleDir = Prefix + "/modules"

	file := Prefix + "/etc/config.toml"
	if err := htoml.DecodeFromFile(file, &Config); err != nil {

		if !os.IsNotExist(err) {
			return err
		}

		if err := json.DecodeFile(Prefix+"/etc/config.json", &Config); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}

		Save()
	}

	{
		if Config.HttpPort == 0 {
			Config.HttpPort = 9533
		}
	}

	if Config.InstanceID != "" {
		SysVersionSign = idhash.HashToHexString([]byte(fmt.Sprintf("%s-%s-%s", Version, Release, Config.InstanceID)), 16)
	} else {
		SysVersionSign = "unreg"
	}

	if Config.AppTitle == "" {
		Config.AppTitle = "Hooto Press"
	}

	if err := syncInnerStackConfig(); err != nil {
		return err
	}

	// Default User
	if User, err = user.Current(); err != nil {
		return err
	}

	if err := storeSetup(); err != nil {
		return err
	}

	if Config.IamAuth == nil {
		Config.IamAuth = &iamserver.AppAuthConfig{}
	}

	if err := iamserver.AppVerifier.Setup(Config.IamAuth); err != nil {
		return err
	}

	// Setting CAPTCHA
	captcha4g.DataConnector = store.DataLocal
	CaptchaConfig.DataDir = Prefix + "/var/hcaptchadb"
	if err := captcha4g.Config(CaptchaConfig); err != nil {
		return err
	}

	//
	{
		rs, err := store.Data.Query(store.Data.NewQueryer().From("hp_sys_config").Limit(1000))
		if err != nil {
			hlog.Print("error", err.Error())
			return err
		}

		for _, v := range rs {

			item := api.SysConfig{
				Key:   v.Field("key").String(),
				Value: v.Field("value").String(),
			}

			if item.Key == "router_basepath_default" {
				item.Value = filepath.Clean("/" + strings.TrimSpace(item.Value))
				if item.Value == "" || item.Value == "." || item.Value == "/" {
					item.Value = "/"
					RouterBasepathDefaults = []string{}
				} else {
					RouterBasepathDefaults = strings.Split(strings.Trim(item.Value, "/"), "/")
				}
				RouterBasepathDefault = item.Value
			}

			if item.Key == "frontend_languages" {
				if langs := api.LangsStringFilterArray(item.Value); len(langs) > 0 {
					Languages = []*api.LangEntry{}
					for _, lv := range langs {
						for _, lv2 := range api.LangArray {
							if lv == lv2.Id {
								Languages = append(Languages, lv2)
							}
						}
					}
				}
			}

			SysConfigList.Insert(item)
		}
	}

	if err := module_init(); err != nil {
		return err
	}

	inited = true

	hlog.Printf("info", "hooto-press inited, version %s, release %s",
		Version, Release)

	return nil
}

func syncInnerStackConfig() error {

	if Config.RunMode == "local-dev" {
		return nil
	}

	conf, err := inconf.NewAppConfigHelper()
	if err != nil {
		return err
	}

	var (
		dbConnOpts = Config.IoConnectors.Options(types.NameIdentifier("hpress_database"))
		chg        = false
	)

	if Config.IamAuth == nil {
		Config.IamAuth = &iamserver.AppAuthConfig{}
	}

	if v, ok := conf.ConfigValueOK("iam_auth_endpoint_url"); ok {
		if v != Config.IamAuth.BaseURL {
			Config.IamAuth.BaseURL = v
			chg = true
		}
	}

	// Find database dependency from App Deploy Depends
	var (
		dbDriver = types.NameIdentifier("lynkdb/pgsqlgo")
		dbDep    = conf.Depend("postgresql-v18")
	)

	if dbDep == nil {
		return errors.New("No Database Connection Config Found")
	}

	dbReplica, dbService := dbDep.Service("postgresql")
	if dbService == nil {
		return errors.New("No Database Connection Service Found")
	}

	dbDriver = types.NameIdentifier("lynkdb/pgsqlgo")

	if dbConnOpts == nil {
		dbConnOpts = &connect.ConnOptions{
			Name:      types.NameIdentifier("hpress_database"),
			Connector: "iomix/rdb/connector",
		}
	}
	if dbConnOpts.Driver != dbDriver {
		dbConnOpts.Driver = dbDriver
	}

	if v := dbDep.ConfigValue("db_name"); v != "" {
		if dbConnOpts.Value("dbname") != v {
			dbConnOpts.SetValue("dbname", v)
			chg = true
		}
	}

	if v := dbDep.ConfigValue("db_user"); v != "" {
		if dbConnOpts.Value("user") != v {
			dbConnOpts.SetValue("user", v)
			chg = true
		}
	}

	if v := dbDep.ConfigValue("db_auth"); v != "" {
		if dbConnOpts.Value("pass") != v {
			dbConnOpts.SetValue("pass", v)
			chg = true
		}
	}

	dbHost := dbReplica.HostIpv4

	if dbReplica.VpcIpv4 != "" {
		dbHost = dbReplica.VpcIpv4
	} else if dbReplica.HostIpv4 != "" {
		dbHost = dbReplica.HostIpv4
	}

	if dbHost != "" && dbConnOpts.Value("host") != dbHost {
		dbConnOpts.SetValue("host", dbHost)
		chg = true
	}

	if dbReplica.VpcIpv4 != "" {
		if p := fmt.Sprintf("%d", dbService.Port); p != dbConnOpts.Value("port") {
			dbConnOpts.SetValue("port", p)
			chg = true
		}
	} else if dbService.HostPort > 0 {
		if p := fmt.Sprintf("%d", dbService.HostPort); p != dbConnOpts.Value("port") {
			dbConnOpts.SetValue("port", p)
			dbConnOpts.SetValue("host", dbReplica.HostIpv4)
			chg = true
		}
	}

	Config.IoConnectors.SetOptions(*dbConnOpts)

	if chg {
		Save()
		hlog.Printf("warn", "sysinner configs synced")
	}

	return nil
}

func storeSetup() error {

	if Config.DataCache == nil {

		Config.DataCache = &storage.Options{
			DataDirectory:   Prefix + "/var/hpress_cache",
			WriteBufferSize: 2,
			BlockCacheSize:  8,
		}

		Save()
	}

	{
		io_name := types.NewNameIdentifier("hpress_database")
		opts := Config.IoConnectors.Options(io_name)

		if opts == nil {
			if Config.RunMode != "local-dev" {
				return errors.New("iomix/rdb/connector " + io_name.String() + " Not Found")
			}
			opts = &connect.ConnOptions{
				Name:      io_name,
				Connector: "iomix/rdb/connector",
				Driver:    types.NewNameIdentifier("lynkdb/pgsqlgo"),
			}
		}

		if opts.Value("host") == "" {
			opts.SetValue("host", "localhost")
		}

		if opts.Value("port") == "" {
			opts.SetValue("port", "5432")
		}

		Config.IoConnectors.SetOptions(*opts)
	}

	if err := store.Setup(Config.DataCache, Config.IoConnectors); err != nil {
		hlog.Printf("error", "storeSetup %s", err.Error())
		return err
	}

	dm, err := store.Data.Modeler()
	if err != nil {
		hlog.Printf("error", "storeSetup %s", err.Error())
		return err
	}

	err = dm.SchemaSyncByJson(dsBase)
	if err != nil {
		hlog.Printf("error", "storeSetup %s", err.Error())
	}

	Save()

	return err
}

func Save() error {
	return htoml.EncodeToFile(Config, Prefix+"/etc/config.toml", nil)
}
