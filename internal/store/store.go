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
	"log/slog"

	"github.com/lynkdb/iomix/connect"
	"github.com/lynkdb/kvgo/v2/pkg/kvapi"
	"github.com/lynkdb/kvgo/v2/pkg/kvrep"
	"github.com/lynkdb/kvgo/v2/pkg/storage"
	"github.com/lynkdb/lynkapi/go/lynktable"
	"github.com/lynkdb/mysqlgo"
	"github.com/lynkdb/pgsqlgo"
)

var (
	err         error
	Data        lynktable.Connector
	DataOptions *connect.ConnOptions
	DataLocal   kvapi.Client
)

// ConnOptsMap converts a connect.ConnOptions into the flat map
// consumed by mysqlgo/pgsqlgo NewConnector.
func ConnOptsMap(opts *connect.ConnOptions) map[string]string {

	m := map[string]string{}
	if opts == nil {
		return m
	}

	for _, v := range opts.Items {
		m[v.Name] = v.Value
	}

	return m
}

func Setup(dbc *storage.Options, cfg connect.MultiConnOptions) error {

	if dbc == nil {
		return errors.New("No hpress_local Config Found")
	}

	if DataLocal, err = kvrep.NewReplica(dbc); err != nil {
		return fmt.Errorf("db open error %s", err.Error())
	}

	opts := cfg.Options("hpress_database")
	if opts == nil {
		slog.Error(err.Error())
		return errors.New("No hpress_database Config.IoConnectors Found")
	}

	switch opts.Driver {

	case "lynkdb/mysqlgo":
		Data, err = mysqlgo.NewConnector(ConnOptsMap(opts))

	case "lynkdb/pgsqlgo":
		Data, err = pgsqlgo.NewConnector(ConnOptsMap(opts))

	default:
		return errors.New("Invalid lynkdb/driver")
	}

	if err != nil {
		slog.Error(fmt.Sprintf("store_init %s", err.Error()))
		return err
	}

	DataOptions = opts

	return nil
}
