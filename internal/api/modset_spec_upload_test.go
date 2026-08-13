// Copyright 2015 Eryx <evorui at gmail dot com>, All rights reserved.
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

package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	stdjson "encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/sysinner/innerstack/v2/pkg/inapi"
	"github.com/ulikunitz/xz"
)

// makeIpk builds an in-memory IPK1 container from a set of tar entries,
// compressing the payload per compress ("xz" | "gzip" | "" = raw tar). The
// header is produced with stdlib json on inapi.Package, matching the real
// innerstack builder so the round-trip exercises the same code path.
func makeIpk(t *testing.T, compress string, files map[string]string) []byte {
	t.Helper()

	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)
	for name, content := range files {
		if strings.HasSuffix(name, "/") {
			if err := tw.WriteHeader(&tar.Header{
				Name:     name,
				Typeflag: tar.TypeDir,
				Mode:     0755,
			}); err != nil {
				t.Fatalf("tar write dir header: %v", err)
			}
			continue
		}
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tar write header: %v", err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatalf("tar write body: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}

	var dataBlock bytes.Buffer
	switch compress {
	case "xz":
		zw, err := xz.NewWriter(&dataBlock)
		if err != nil {
			t.Fatalf("xz writer: %v", err)
		}
		if _, err := zw.Write(tarBuf.Bytes()); err != nil {
			t.Fatalf("xz write: %v", err)
		}
		if err := zw.Close(); err != nil {
			t.Fatalf("xz close: %v", err)
		}
	case "gzip":
		gw := gzip.NewWriter(&dataBlock)
		if _, err := gw.Write(tarBuf.Bytes()); err != nil {
			t.Fatalf("gzip write: %v", err)
		}
		if err := gw.Close(); err != nil {
			t.Fatalf("gzip close: %v", err)
		}
	case "":
		dataBlock.Write(tarBuf.Bytes())
	default:
		t.Fatalf("unsupported compress in fixture: %q", compress)
	}

	headerBytes, err := stdjson.Marshal(&inapi.Package{
		Release: &inapi.PackageRelease{Compress: compress},
	})
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}

	var out bytes.Buffer
	out.WriteString(ipkMagic)
	var lb [4]byte
	binary.LittleEndian.PutUint32(lb[:], uint32(len(headerBytes)))
	out.Write(lb[:])
	out.Write(headerBytes)
	out.Write(dataBlock.Bytes())
	return out.Bytes()
}

// mkIpkRaw assembles an IPK1 with an explicit header JSON and data block,
// for crafting malformed fixtures (bad header json, missing release, etc.).
func mkIpkRaw(header, data []byte) []byte {
	var out bytes.Buffer
	out.WriteString(ipkMagic)
	var lb [4]byte
	binary.LittleEndian.PutUint32(lb[:], uint32(len(header)))
	out.Write(lb[:])
	out.Write(header)
	out.Write(data)
	return out.Bytes()
}

// mkIpkHeaderLen assembles an IPK1 with a hand-set declared header length
// (independent of the actual header bytes), for overflow / out-of-bounds tests.
func mkIpkHeaderLen(headerLen uint32, body []byte) []byte {
	var out bytes.Buffer
	out.WriteString(ipkMagic)
	var lb [4]byte
	binary.LittleEndian.PutUint32(lb[:], headerLen)
	out.Write(lb[:])
	out.Write(body)
	return out.Bytes()
}

func TestParseIpkPackage(t *testing.T) {
	const specJSON = `{"meta":{"name":"ruilog/notebook","version":"0.0.36"}}`

	for _, compress := range []string{"xz", "gzip", ""} {
		label := "compress_" + compress
		if compress == "" {
			label = "compress_none"
		}
		t.Run(label, func(t *testing.T) {
			data := makeIpk(t, compress, map[string]string{"spec.json": specJSON})

			hdr, cpr, err := parseIpkPackage(data)
			if err != nil {
				t.Fatalf("parseIpkPackage: unexpected error: %v", err)
			}
			if hdr == nil || hdr.Release == nil {
				t.Fatalf("expected non-nil header/release")
			}
			if hdr.Release.GetCompress() != compress {
				t.Fatalf("compress = %q, want %q", hdr.Release.GetCompress(), compress)
			}

			tr := tar.NewReader(io.LimitReader(cpr, ipkDecompressCap))
			th, err := tr.Next()
			if err != nil {
				t.Fatalf("tar next: %v", err)
			}
			if th.Name != "spec.json" {
				t.Fatalf("entry = %q, want spec.json", th.Name)
			}
			body, err := io.ReadAll(tr)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if string(body) != specJSON {
				t.Fatalf("body = %q, want %q", body, specJSON)
			}
		})
	}
}

func TestParseIpkPackageErrors(t *testing.T) {
	good := makeIpk(t, "xz", map[string]string{"spec.json": "{}"})

	// header declaring compress=bzip2 but otherwise well-formed.
	bzip2Hdr, _ := stdjson.Marshal(&inapi.Package{
		Release: &inapi.PackageRelease{Compress: "bzip2"},
	})

	cases := []struct {
		name string
		data []byte
	}{
		{"too_small", []byte("IPK")},
		{"bad_magic", append([]byte("XXXXXXXX"), good[8:]...)},
		// headerLen far exceeds the cap: must be rejected, never OOM.
		{"header_len_overflow", mkIpkHeaderLen(0xFFFFFFFF, []byte("{}\nrawdata"))},
		// headerLen claims 100 bytes but only 99 follow: 8+100 > len.
		{"header_len_out_of_bounds", mkIpkHeaderLen(100, bytes.Repeat([]byte("x"), 99))},
		{"bad_header_json", mkIpkRaw([]byte("{"), []byte("datablock"))},
		{"release_nil", mkIpkRaw([]byte("{}"), []byte("datablock"))},
		{"empty_data_block", mkIpkRaw([]byte("{}"), nil)},
		{"unsupported_compress", mkIpkRaw(bzip2Hdr, []byte("datablock"))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := parseIpkPackage(tc.data); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
