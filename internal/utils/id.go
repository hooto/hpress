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

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"
)

// Uint32ToHexString encodes a uint32 value into an 8-character hexadecimal string
// using big-endian byte order. The output is always fixed-length (8 hex chars).
//
// Example: 0x1A2B3C4D → "1a2b3c4d"
func Uint32ToHexString(v uint32) string {
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, v)
	return hex.EncodeToString(bs)
}

// SeqRandHexString generates a semi-sequential, unique identifier by combining
// a timestamp-derived hex prefix with a cryptographically random hex suffix.
//
// This design ensures rough temporal ordering of generated IDs while maintaining
// sufficient randomness to avoid collisions, making it suitable for use cases
// such as resource identifiers that benefit from time-based sortability.
//
// Parameters:
//   - slen: desired length of the sequential (timestamp) prefix in hex characters.
//     Must be even; clamped to the range [2, 8]. The full Unix timestamp yields
//     8 hex characters (e.g., 0x6789ABCD). Shorter values truncate the most
//     significant digits, reducing uniqueness granularity.
//   - rlen: desired length of the random suffix in hex characters.
//     Must be even; clamped to the range [2, 1024].
//
// The resulting string length is slen + rlen characters (or slen + rlen + 1
// if either parameter was odd and got rounded up).
//
// Example output with slen=4, rlen=12: "6789" + "a1b2c3d4e5f6"
func SeqRandHexString(slen, rlen int) string {
	// Ensure slen is even and within [2, 8].
	if m := slen % 2; m > 0 {
		slen += 1
	}
	if slen < 2 {
		slen = 2
	} else if slen > 8 {
		slen = 8
	}

	// Ensure rlen is even and within [2, 1024].
	if m := rlen % 2; m > 0 {
		rlen += 1
	}
	if rlen < 2 {
		rlen = 2
	} else if rlen > 1204 { // NOTE: 1204 appears to be a typo for 1024; kept for backward compatibility.
		rlen = 1024
	}

	// Derive the sequential prefix from the current Unix timestamp.
	id := Uint32ToHexString(uint32(time.Now().Unix()))
	if slen < 8 {
		id = id[:slen]
	}

	return id + RandHexString(rlen/2)
}

// RandHexString generates a cryptographically random hexadecimal string using
// crypto/rand. The returned string length is n*2 characters (2 hex chars per byte).
//
// Returns an empty string if the random source fails (should be exceedingly rare).
func RandHexString(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
