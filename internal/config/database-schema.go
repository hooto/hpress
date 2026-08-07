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

const dsBase = `
{
    "engine": "MyISAM",
    "charset": "utf8",
    "tables": [
    	{
            "name": "hp_modules",
            "columns": [
                {
                    "name": "name",
                    "type": "string",
                    "length": "30"
                },
                {
                    "name": "srvname",
                    "type": "string",
                    "length": "50"
                },
                {
                    "name": "status",
                    "type": "int16",
                    "default": "1"
                },
                {
                    "name": "version",
                    "type": "string",
                    "length": "10",
                    "default": "0"
                },
                {
                    "name": "title",
                    "type": "string",
                    "length": "100"
                },
                {
                    "name": "body",
                    "type": "string-text"
                },
                {
                    "name": "created",
                    "type": "uint32"
                },
                {
                    "name": "updated",
                    "type": "uint32"
                }
            ],
            "indexes": [
                {
                    "name": "PRIMARY",
                    "type": 3,
                    "cols": ["name"]
                },
                {
                    "name": "srvname",
                    "type": 2,
                    "cols": ["srvname"]
                },
                {
                    "name": "status",
                    "type": 1,
                    "cols": ["status"]
                },
                {
                    "name": "created",
                    "type": 1,
                    "cols": ["created"]
                },
                {
                    "name": "updated",
                    "type": 1,
                    "cols": ["updated"]
                }
            ]
        },
        {
            "name": "hp_sys_config",
            "columns": [
                {
                    "name": "key",
                    "type": "string",
                    "length": "50"
                },
                {
                    "name": "value",
                    "type": "string-text"
                }
            ],
            "indexes": [
                {
                    "name": "PRIMARY",
                    "type": 3,
                    "cols": ["key"]
                }
            ]
        }
    ]
}
`

const (
	dsTplNodeModels = `
{
    "columns": [
        {
            "name": "id",
            "type": "string",
            "length": "16"
        },
        {
            "name": "pid",
            "type": "string",
            "length": "16"
        },
        {
            "name": "status",
            "type": "int16"
        },
        {
            "name": "userid",
            "type": "string",
            "length": "10"
        },
        {
            "name": "title",
            "type": "string",
            "length": "100"
        },
        {
            "name": "created",
            "type": "uint32"
        },
        {
            "name": "updated",
            "type": "uint32"
        }
    ],
    "indexes": [
        {
            "name": "PRIMARY",
            "type": 3,
            "cols": ["id"]
        },
        {
            "name": "pid",
            "type": 1,
            "cols": ["pid"]
        },
        {
            "name": "status",
            "type": 1,
            "cols": ["status"]
        },
        {
            "name": "userid",
            "type": 1,
            "cols": ["userid"]
        },
        {
            "name": "created",
            "type": 1,
            "cols": ["created"]
        },
        {
            "name": "updated",
            "type": 1,
            "cols": ["updated"]
        }
    ]
}
`
	dsTplTermModels = `
{
    "name": "template",
    "columns": [
        {
            "name": "id",
            "type": "uint32",
            "incr_able": true
        },
        {
            "name": "status",
            "type": "int16"
        },
        {
            "name": "userid",
            "type": "string",
            "length": "10"
        },
        {
            "name": "title",
            "type": "string",
            "length": "100"
        },
        {
            "name": "created",
            "type": "uint32"
        },
        {
            "name": "updated",
            "type": "uint32"
        }
    ],
    "indexes": [
        {
            "name": "PRIMARY",
            "type": 3,
            "cols": ["id"]
        },
        {
            "name": "status",
            "type": 1,
            "cols": ["status"]
        },
        {
            "name": "userid",
            "type": 1,
            "cols": ["userid"]
        },
        {
            "name": "created",
            "type": 1,
            "cols": ["created"]
        },
        {
            "name": "updated",
            "type": 1,
            "cols": ["updated"]
        }
    ]
}
`
)
