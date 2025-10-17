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

package frontend

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/httpsrv"
	"github.com/lessos/lessgo/crypto/idhash"
	"github.com/lessos/lessgo/sync"

	// "code.ivysaur.me/imagequant/v2"
	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"

	"github.com/hooto/hpress/config"
	"github.com/hooto/hpress/store"
)

const (
	s2BucketDeft  = "/deft/"
	s3UrlPrefix   = "/hp/s2"
	objCacheTTL   = int64(86400 * 1000) // ms
	iplStep       = 64
	iplSizeMin    = iplStep
	iplSizeMax    = 2048
	pngQualityMin = 10
	pngQualityMax = 40
)

var (
	rezlocker  = sync.NewPermitPool(runtime.NumCPU())
	s2PathRE   = regexp.MustCompile("^[0-9a-zA-Z_\\-\\.\\/]{1,100}$")
	pngCmpPath = ""
)

func init() {

	pngCmpPath = config.Prefix + "/bin/pngquant"
	if _, err := os.Stat(pngCmpPath); err != nil {
		pngCmpPath, err = exec.LookPath("pngquant")
		if err != nil {
			pngCmpPath = ""
		}
	}
}

type S2 struct {
	*httpsrv.Controller
}

// resize
func (c S2) IndexAction() {
	s2Server(c.Controller, "", "")
}

func s2Server(c *httpsrv.Controller, objPath, absPath string) {

	c.AutoRender = false
	var err error

	if objPath == "" {
		objPath, err = pathFilter(strings.TrimPrefix(c.Request.UrlPath(), s3UrlPrefix))
		if err != nil {
			c.RenderError(404, "Object Not Found #1")
			return
		}
	}

	if absPath == "" {
		absPath = config.Prefix + "/var/storage/" + objPath
	}

	var (
		ipn       = c.Params.Value("ipn") // v1
		ipl       = c.Params.Value("ipl") // v2
		ipls      = [2]int{0, 0}          // width, height
		iplCrop   = false
		fileExt   = strings.ToLower(filepath.Ext(objPath))
		mediaType = ""
	)

	switch fileExt {

	case ".jpg", ".jpeg":
		mediaType = "image/jpeg"

	case ".png":
		mediaType = "image/png"
		if ipn == "" && ipl == "" {
			ipl = "w2000"
		}

	case ".gif":
		mediaType = "image/gif"

	case ".svg":
		mediaType = "image/svg+xml"

	default:
		c.RenderError(400, "Bad Request (media type not support)")
		return
	}

	if (ipn == "" && ipl == "") ||
		mediaType == "image/svg+xml" {

		if fp, err := os.Open(absPath); err == nil {

			c.Response.Out.Header().Set("Cache-Control", "max-age=86400")
			http.ServeContent(c.Response.Out, c.Request.Request, objPath, time.Now(), fp)
			fp.Close()

		} else {
			c.RenderError(404, "Object Not Found #2")
		}

		return
	}

	if ipl != "" {

		ar := strings.Split(ipl, ",")

		for _, k := range ar {

			if k[0] < 'a' || k[0] > 'z' {
				continue
			}

			switch k[0] {
			case 'w':
				if len(k) > 1 {
					ipls[0], _ = strconv.Atoi(k[1:])
				}

			case 'h':
				if len(k) > 1 {
					ipls[1], _ = strconv.Atoi(k[1:])
				}

			case 'c':
				iplCrop = true
			}
		}

	} else if ipn != "" { // TORM

		switch ipn {
		case "i6040":
			ipls[0], ipls[1] = 60, 40

		case "s800":
			ipls[0], ipls[1], iplCrop = 800, 800, false

		case "s800x":
			ipls[0], ipls[1], iplCrop = 800, 8000, false

		case "thumb":
			ipls[0], ipls[1] = 150, 150

		case "medium":
			ipls[0], ipls[1] = 300, 168

		case "large":
			ipls[0], ipls[1] = 800, 450

		default:
			c.RenderError(400, "Bad Request (ipn)")
			return
		}
	}

	if ipls[0] > 0 && ipls[1] < 1 {
		ipls[1] = ipls[0]
	} else if ipls[1] > 0 && ipls[0] < 1 {
		ipls[0] = ipls[1]
	}

	for i := range ipls {
		ipls[i] = ipls[i] - (ipls[i] % iplStep)

		if ipls[i] < iplSizeMin {
			ipls[i] = iplSizeMin
		} else if ipls[i] > iplSizeMax {
			ipls[i] = iplSizeMax
		}
	}

	var (
		key = fmt.Sprintf("%s_%d.%d.%t", absPath, ipls[0], ipls[1], iplCrop)
		hid = "s2." + idhash.HashToHexString([]byte(key), 12)
	)

	if rs := store.DataLocal.NewReader([]byte(hid)).Exec(); rs.OK() {
		imBytes := rs.Item().Value
		c.Response.Out.Header().Set("Cache-Control", "max-age=86400")
		c.Response.Out.Header().Set("Content-type", mediaType)
		c.Response.Out.Write(imBytes)
		hlog.Printf("debug", "image cache hit %s, bytes %d", key, len(imBytes))
		return
	}

	pn := rezlocker.Pull()
	defer rezlocker.Push(pn)

	defer func() {
		if x := recover(); x != nil {
			c.RenderError(404, "Object Not Found #3 ")
		}
	}()

	fp, err := os.Open(absPath)
	if err != nil {
		c.RenderError(400, "Bad Request (invalid object path)")
		return
	}
	defer fp.Close()

	//
	var srcImg image.Image

	switch fileExt {

	case ".jpg", ".jpeg":
		srcImg, err = jpeg.Decode(fp)

	case ".png":
		srcImg, err = png.Decode(fp)

	case ".gif":
		srcImg, err = gif.Decode(fp)
	}

	if err != nil {
		c.RenderError(400, "Bad Request (invalid object format) "+err.Error())
		return
	}

	//
	var (
		dstImg    image.Image
		dstBuf    bytes.Buffer
		srcBounds = srcImg.Bounds()
	)

	if ipls[0] < srcBounds.Dx() || ipls[1] < srcBounds.Dy() {

		if iplCrop {
			if im, err := cutter.Crop(srcImg, cutter.Config{
				Width:   ipls[0],
				Height:  ipls[1],
				Mode:    cutter.Centered,
				Options: cutter.Ratio,
			}); err == nil {
				dstImg = resize.Thumbnail(uint(ipls[0]), uint(ipls[1]), im, resize.Lanczos3)
			}
		} else {

			rate := float32(1.0)

			if hrate := float32(ipls[1]) / float32(srcBounds.Dy()); hrate < rate {
				rate = hrate
			}

			if wrate := float32(ipls[0]) / float32(srcBounds.Dx()); wrate < rate {
				rate = wrate
			}

			if rate < 1 {
				ipls[0] = int(float32(srcBounds.Dx()) * rate)
				ipls[1] = int(float32(srcBounds.Dy()) * rate)
			}
		}

		if dstImg == nil {
			dstImg = resize.Thumbnail(uint(ipls[0]), uint(ipls[1]), srcImg, resize.Lanczos3)
		}

	} else {
		dstImg = srcImg
	}

	//
	switch fileExt {

	case ".jpg", ".jpeg":
		err = jpeg.Encode(&dstBuf, dstImg, &jpeg.Options{90})

	case ".png":

		if pngCmpPath != "" {
			pngEnc := png.Encoder{
				CompressionLevel: png.NoCompression,
			}
			if err = pngEnc.Encode(&dstBuf, dstImg); err == nil {
				var dstCmp bytes.Buffer
				if err = pngCompressCmd(&dstBuf, &dstCmp); err == nil {
					dstBuf = dstCmp
				}
			}
		} else {
			pngEnc := png.Encoder{
				CompressionLevel: png.BestCompression,
			}
			err = pngEnc.Encode(&dstBuf, dstImg)
		}

		// if imc, err := pngCompress(dstImg); err == nil {
		// 	dstImg = imc
		// }

	case ".gif":
		err = gif.Encode(&dstBuf, dstImg, &gif.Options{NumColors: 256})
	}
	if err != nil {
		c.RenderError(400, "Bad Request : "+err.Error())
		return
	}

	//
	if dstBuf.Len() > 10 {
		store.DataLocal.NewWriter([]byte(hid), dstBuf.Bytes()).
			SetTTL(imageCacheTTL()).Exec()
	}

	//
	c.Response.Out.Header().Set("Content-type", mediaType)
	c.Response.Out.Header().Set("Cache-Control", "max-age=86400")
	c.Response.Out.Write(dstBuf.Bytes())

	if st, err := fp.Stat(); err == nil {
		hlog.Printf("info", "image resize %s from %d to %d bytes",
			key, st.Size(), dstBuf.Len())
	}
}

func pngCompressCmd(r io.Reader, w *bytes.Buffer) error {

	cmd := exec.Command("pngquant", "-", fmt.Sprintf("--quality=%d-%d", pngQualityMin, pngQualityMax))

	cmd.Stdin = r
	cmd.Stdout = w

	return cmd.Run()
}

func pathFilter(path string) (string, error) {

	path = filepath.Clean(strings.Replace(strings.TrimSpace(path), " ", "-", -1))
	if !s2PathRE.MatchString(path) {
		return path, fmt.Errorf("Invalid File Name")
	}

	if !strings.HasPrefix(path, s2BucketDeft) {
		return "", errors.New("Invalid Bucket Name")
	}

	return path, nil
}

func imageCacheTTL() int64 { // ms
	return objCacheTTL + rand.Int63n(objCacheTTL)
}

/**
// 1. memory leak, 2. high CPU cost
func pngCompress(im image.Image) (image.Image, error) {

	attr, err := imagequant.NewAttributes()
	if err != nil {
		return nil, err
	}
	defer attr.Release()

	attr.SetQuality(pngQualityMin, pngQualityMax)

	rgba32data := imageToRgba32(im)
	iqm, err := imagequant.NewImage(attr, string(rgba32data),
		im.Bounds().Max.X, im.Bounds().Max.Y, 0)
	rgba32data = nil
	if err != nil {
		return nil, err
	}
	defer iqm.Release()

	res, err := iqm.Quantize(attr)
	if err != nil {
		return nil, err
	}
	defer res.Release()

	rgb8data, err := res.WriteRemappedImage()
	if err != nil {
		return nil, err
	}

	im1, err := rgb8PaletteToImage(res.GetImageWidth(), res.GetImageHeight(),
		rgb8data, res.GetPalette()), nil
	rgb8data = nil
	return im1, err
}
*/

func imageToRgba32(im image.Image) []byte {

	var (
		w   = im.Bounds().Max.X
		h   = im.Bounds().Max.Y
		ret = make([]byte, w*h*4)
		p   = 0
	)

	for y := 0; y < h; y += 1 {
		for x := 0; x < w; x += 1 {
			r16, g16, b16, a16 := im.At(x, y).RGBA() // Each value ranges within [0, 0xffff]

			ret[p+0] = uint8(r16 >> 8)
			ret[p+1] = uint8(g16 >> 8)
			ret[p+2] = uint8(b16 >> 8)
			ret[p+3] = uint8(a16 >> 8)
			p += 4
		}
	}

	return ret
}

func rgb8PaletteToImage(w, h int, rgb8data []byte, pal color.Palette) image.Image {

	var (
		rect = image.Rectangle{
			Max: image.Point{
				X: w,
				Y: h,
			},
		}
		ret = image.NewPaletted(rect, pal)
	)

	for y := 0; y < h; y += 1 {
		for x := 0; x < w; x += 1 {
			ret.SetColorIndex(x, y, rgb8data[y*w+x])
		}
	}

	return ret
}
