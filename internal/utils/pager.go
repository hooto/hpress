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

package utils

type Paginator struct {
	ItemCount         uint64
	CountPerPage      uint64
	PageCount         uint64
	CurrentPageNumber uint64
	FirstPageNumber   uint64
	PrevPageNumber    uint64
	NextPageNumber    uint64
	LastPageNumber    uint64
	RangeLen          uint64
	RangeStartNumber  uint64
	RangeEndNumber    uint64
	RangePages        []uint64
}

// Pager build a page collection
//
// @param uint64 current  Number of Current page
// @param uint64 count    Total number of items available
// @param uint64 perpage  Number of items per page
// @param uint64 rangelen Number of discrete page numbers that will be displayed
func NewPager(current, count, perpage, rangelen uint64) Paginator {

	pager := Paginator{
		ItemCount:         count,
		CountPerPage:      perpage,
		CurrentPageNumber: current,
		RangeLen:          rangelen,
	}

	//
	if pager.ItemCount < 0 {
		pager.ItemCount = 0
	}

	if pager.CountPerPage < 1 {
		pager.CountPerPage = 10
	} else if pager.CountPerPage > 100 {
		pager.CountPerPage = 100
	}

	if pager.RangeLen < 1 {
		pager.RangeLen = 1
	}

	pager.PageCount = pager.ItemCount / pager.CountPerPage
	if (pager.ItemCount % pager.CountPerPage) > 0 {
		pager.PageCount++
	}

	// fixed current page number
	if pager.CurrentPageNumber < 1 {
		pager.CurrentPageNumber = 1
	} else if pager.CurrentPageNumber > pager.PageCount {
		pager.CurrentPageNumber = pager.PageCount
	}

	//
	pager.RangeStartNumber = 1
	if pager.CurrentPageNumber > rangelen/2 {
		pager.RangeStartNumber = pager.CurrentPageNumber - (rangelen / 2)
	}

	pager.RangeEndNumber = pager.PageCount
	if (pager.RangeStartNumber + rangelen) < pager.PageCount {
		pager.RangeEndNumber = pager.RangeStartNumber + rangelen - 1
	}

	// taking previous page
	if pager.CurrentPageNumber > 1 {
		pager.PrevPageNumber = pager.CurrentPageNumber - 1
	}

	// taking next page
	if pager.CurrentPageNumber < pager.PageCount {
		pager.NextPageNumber = pager.CurrentPageNumber + 1
	}

	// taking pages list
	for i := pager.RangeStartNumber; i <= pager.RangeEndNumber; i++ {
		pager.RangePages = append(pager.RangePages, i)
	}

	if pager.RangeStartNumber > 1 {
		pager.FirstPageNumber = 1
	}

	if pager.RangeEndNumber < pager.PageCount {
		pager.LastPageNumber = pager.PageCount
	}

	return pager
}
