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
	"github.com/hooto/iam/iamapi"
	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/encoding/json"
	"github.com/lessos/lessgo/types"
	"github.com/lynkdb/iomix/connect"
	"github.com/lynkdb/kvgo/v2/pkg/storage"
	"github.com/sysinner/incore/inconf"

	"github.com/hooto/hpress/api"
	"github.com/hooto/hpress/store"
)

var (
	Prefix         string
	Config         ConfigCommon
	AppName        = "hooto-press"
	Version        = "0.9.0.dev"
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
	UrlBasePath           string                   `json:"url_base_path,omitempty" toml:"url_base_path,omitempty"`
	ModuleDir             string                   `json:"module_dir,omitempty" toml:"module_dir,omitempty"`
	InstanceID            string                   `json:"instance_id" toml:"instance_id"`
	AppInstance           iamapi.AppInstance       `json:"app_instance" toml:"app_instance"`
	AppTitle              string                   `json:"app_title,omitempty" toml:"app_title,omitempty"`
	HttpPort              uint16                   `json:"http_port" toml:"http_port"`
	IamServiceUrl         string                   `json:"iam_service_url" toml:"iam_service_url"`
	IamServiceUrlFrontend string                   `json:"iam_service_url_frontend" toml:"iam_service_url_frontend"`
	IoConnectors          connect.MultiConnOptions `json:"io_connectors" toml:"io_connectors"`
	DataCache             *storage.Options         `json:"data_cache" toml:"data_cache"`
	RunMode               string                   `json:"run_mode,omitempty" toml:"run_mode,omitempty"`
	ExtUpDatabases        connect.MultiConnOptions `json:"ext_up_databases,omitempty" toml:"ext_up_databases,omitempty"`
	ExpModuleInits        []string                 `json:"exp_module_inits,omitempty" toml:"exp_module_inits,omitempty"`
	ExpGdocPaths          []string                 `json:"exp_gdoc_paths,omitempty" toml:"exp_gdoc_paths,omitempty"`
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
			if err := syncSysinnerConfig(); err != nil {
				hlog.Printf("error", "syncSysinnerConfig err: %s", err.Error())
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

	if err := syncSysinnerConfig(); err != nil {
		return err
	}

	// Default User
	if User, err = user.Current(); err != nil {
		return err
	}

	if err := storeSetup(); err != nil {
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

func syncSysinnerConfig() error {

	if Config.RunMode == "local-dev" {
		return nil
	}

	conf, err := inconf.NewAppConfigurator("hooto-press*")
	if err != nil {
		return err
	}

	opt := conf.AppConfig("cfg/hooto-press")
	if opt == nil {
		return errors.New("No Configure Found")
	}

	var (
		dbConnOpts = Config.IoConnectors.Options(types.NameIdentifier("hpress_database"))
		chg        = false
	)

	if v, ok := opt.ValueOK("iam_service_url"); ok {
		if v.String() != Config.IamServiceUrl {
			Config.IamServiceUrl, chg = v.String(), true
		}
	} else {
		return errors.New("No Config.IamServiceUrl Found")
	}

	if v, ok := opt.ValueOK("iam_service_url_frontend"); ok {
		if v.String() != Config.IamServiceUrlFrontend {
			Config.IamServiceUrlFrontend, chg = v.String(), true
		}
	} else {
		Config.IamServiceUrlFrontend = Config.IamServiceUrl
	}

	var (
		dbService = conf.AppServiceQuery("spec=sysinner-pgsql-*")
		dbCfg     = conf.AppConfigQuery("cfg/sysinner-pgsql")
		dbDriver  = types.NameIdentifier("lynkdb/pgsqlgo")
	)

	if dbService == nil {
		dbService = conf.AppServiceQuery("spec=sysinner-mysql-*")
		dbCfg = conf.AppConfigQuery("cfg/sysinner-mysql")
		dbDriver = "lynkdb/mysqlgo"
	}

	if dbService == nil {
		return errors.New("No Database Connection Service Found")
	}

	if dbCfg == nil {
		return errors.New("No Database Connection Config Found")
	}

	if dbConnOpts == nil {
		dbConnOpts = &connect.ConnOptions{
			Name:      types.NameIdentifier("hpress_database"),
			Connector: "iomix/rdb/connector",
		}
	}
	if dbConnOpts.Driver != dbDriver {
		dbConnOpts.Driver = dbDriver
	}

	if v, ok := dbCfg.ValueOK("db_name"); ok {
		if dbConnOpts.Value("dbname") != v.String() {
			dbConnOpts.SetValue("dbname", v.String())
			chg = true
		}
	}

	if v, ok := dbCfg.ValueOK("db_user"); ok {
		if dbConnOpts.Value("user") != v.String() {
			dbConnOpts.SetValue("user", v.String())
			chg = true
		}
	}

	if v, ok := dbCfg.ValueOK("db_auth"); ok {
		if dbConnOpts.Value("pass") != v.String() {
			dbConnOpts.SetValue("pass", v.String())
			chg = true
		}
	}

	if dbConnOpts.Value("host") != dbService.Endpoints[0].Ip {
		dbConnOpts.SetValue("host", dbService.Endpoints[0].Ip)
		chg = true
	}
	if p := fmt.Sprintf("%d", dbService.Endpoints[0].Port); p != dbConnOpts.Value("port") {
		dbConnOpts.SetValue("port", p)
		chg = true
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
