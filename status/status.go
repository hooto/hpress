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

package status

import (
	"sync"
	"time"

	"github.com/hooto/iam/v2/pkg/iamapi"
	"github.com/hooto/iam/v2/pkg/iamserver"
	"github.com/lessos/lessgo/utilx"

	"github.com/hooto/hpress/config"
)

const (
	IamServiceOK           int = 1200
	IamServiceUnavailable  int = 1503
	IamServiceUnRegistered int = 1501
)

var (
	Uptime           string
	IamServiceStatus int
	Locker           sync.RWMutex
)

func init() {
	Uptime = utilx.TimeNow("atom")
}

func Init() {
	go func() {

		for {

			Refresh()

			if IamServiceStatus == IamServiceUnRegistered ||
				IamServiceStatus == IamServiceOK {
				break
			}

			time.Sleep(3e9)
		}
	}()
}

func Refresh() {

	err := iamserver.AppVerifier.Update(&iamapi.AppInstance{
		ID:          config.Config.IamAuth.AppId,
		Version:     config.Version,
		Permissions: config.Perms,
	})

	if err == nil {
		IamServiceStatus = IamServiceOK
	}
}
