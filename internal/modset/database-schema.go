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
            "type": "pri",
            "columns": ["id"]
        },
        {
            "type": "idx",
            "columns": ["pid"]
        },
        {
            "type": "idx",
            "columns": ["status"]
        },
        {
            "type": "idx",
            "columns": ["userid"]
        },
        {
            "type": "idx",
            "columns": ["created"]
        },
        {
            "type": "idx",
            "columns": ["updated"]
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
            "type": "pri",
            "columns": ["id"]
        },
        {
            "type": "idx",
            "columns": ["status"]
        },
        {
            "type": "idx",
            "columns": ["userid"]
        },
        {
            "type": "idx",
            "columns": ["created"]
        },
        {
            "type": "idx",
            "columns": ["updated"]
        }
    ]
}
`
)
