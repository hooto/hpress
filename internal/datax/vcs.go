// Copyright 2019 Eryx <evorui аt gmail dοt com>, All rights reserved.
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

package datax

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hooto/hini4g/hini"
)

var (
	vcsMu        sync.Mutex
	vcsGitBin    = "git"
	vcsActive    *vcsActiveItem
	VcsGitVerReg = regexp.MustCompile(`^[a-f0-9]{30,50}$`)
)

type VcsRepoItem struct {
	Url      string `json:"url"`
	Branch   string `json:"branch"`
	Dir      string `json:"dir"`
	AuthUser string `json:"auth_user,omitempty"`
	AuthPass string `json:"auth_pass,omitempty"`
}

const (
	VcsActionInit    uint32 = 1 << 16
	VcsActionPullOK  uint32 = 1 << 19
	VcsActionPullER  uint32 = 1 << 20
	vcsOnUpdateRange uint32 = 600
)

type vcsActiveItem struct {
	dir    string
	Cmd    *exec.Cmd
	outbuf *bytes.Buffer
	errbuf *bytes.Buffer
	output string
}

func (it *vcsActiveItem) Output() string {
	it.output += it.outbuf.String()
	it.output += it.errbuf.String()
	return it.output
}

func targetDir(dir string) string {
	return filepath.Clean(dir)
}

func OpActionAllow(opbase, op uint32) bool {
	return (op & opbase) == op
}

func vcsAction(vit *VcsRepoItem) (string, error) {

	vcsMu.Lock()
	defer vcsMu.Unlock()

	//
	err := vcsGitPrepare(vit)
	if err != nil {
		slog.Info(fmt.Sprintf("git pull %s", err.Error()))
		return "", err
	}

	//
	err = vcsGitFetch(vit)
	if err != nil {
		slog.Info(fmt.Sprintf("git pull %s", err.Error()))
		return "", err
	}

	ver, err := vcsGitCheckoutAndMerge(vit)
	if err != nil {
		return "", err
	}

	return ver[:12], nil
}

func vcsGitPrepare(vit *VcsRepoItem) error {

	var (
		tdir     = targetDir(vit.Dir)
		conf     = tdir + "/.git/config"
		cfp, err = os.Open(conf)
		url      = vit.Url
	)

	if err != nil {

		if !os.IsNotExist(err) {
			return err
		}

		if err = os.MkdirAll(tdir, 0755); err != nil {
			return err
		}

		if _, err = exec.Command(vcsGitBin, "init", tdir).Output(); err != nil {
			return err
		}

		cfp, err = os.Open(conf)
		if err != nil {
			return err
		}
	}
	defer cfp.Close()

	bs, err := ioutil.ReadAll(cfp)
	if err != nil {
		return err
	}
	opts, err := hini.ParseString(string(bs))
	if err != nil {
		return err
	}

	if vit.AuthUser != "" && vit.AuthPass != "" {
		if strings.HasPrefix(url, "http://") {
			url = strings.Replace(url, "http://", "http://"+vit.AuthUser+":"+vit.AuthPass+"@", 1)
		} else if strings.HasPrefix(url, "https://") {
			url = strings.Replace(url, "https://", "https://"+vit.AuthUser+":"+vit.AuthPass+"@", 1)
		}
	}

	local_url, ok := opts.ValueOK("remote/origin/url")
	if !ok {
		args := []string{
			"--git-dir=" + tdir + "/.git",
			"remote",
			"add",
			"origin",
			url,
		}
		if _, err = exec.Command(vcsGitBin, args...).Output(); err != nil {
			return err
		}

		opts, err = hini.ParseFile(conf)
		if err != nil {
			return err
		}
		local_url, ok = opts.ValueOK("remote/origin/url")
		if !ok {
			return errors.New("git remote set-url fail")
		}
	}

	if local_url.String() != url {
		args := []string{
			"--git-dir=" + tdir + "/.git",
			"remote",
			"set-url",
			"origin",
			url,
		}
		if _, err = exec.Command(vcsGitBin, args...).Output(); err != nil {
			return err
		}
	}

	return nil
}

func vcsGitFetch(vit *VcsRepoItem) error {

	//
	msg := ""
	tdir := targetDir(vit.Dir)
	args := []string{
		"--git-dir=" + tdir + "/.git",
		"fetch",
		"origin",
		vit.Branch,
	}
	os.Setenv("GIT_TERMINAL_PROMPT", "0")
	cmd := exec.Command(vcsGitBin, args...)

	vcsActive := &vcsActiveItem{
		dir:    vit.Dir,
		Cmd:    cmd,
		errbuf: &bytes.Buffer{},
		outbuf: &bytes.Buffer{},
	}

	cmd.Stderr = vcsActive.errbuf
	cmd.Stdout = vcsActive.outbuf

	if err := cmd.Start(); err != nil {
		return err
	} else {
		cmd.Wait()
	}

	if vcsActive.Cmd.ProcessState != nil && vcsActive.Cmd.ProcessState.Exited() {

		if vcsActive.Cmd.ProcessState.Success() {
			//
		} else {
			if strings.Contains(vcsActive.Output(), "could not read Username") {
				msg = "VcsRepo Auth Fail"
			} else {
				msg = "unknown error " + vcsActive.Cmd.ProcessState.String()
			}
		}
	} else {
		msg = "unknown error"
	}

	if vcsActive.Cmd.Process != nil {
		vcsActive.Cmd.Process.Kill()
		time.Sleep(5e8)
	}

	if msg != "" {
		return errors.New(msg)
	}

	return nil
}

func vcsGitCheckoutAndMerge(vit *VcsRepoItem) (string, error) {

	//
	tdir := targetDir(vit.Dir)
	args := []string{
		"--git-dir=" + tdir + "/.git",
		"checkout",
		vit.Branch,
	}
	cmd := exec.Command(vcsGitBin, args...)
	cmd.Dir = tdir
	if _, err := cmd.Output(); err != nil {
		return "", errors.New("failed to checkout branch " + err.Error())
	}

	//
	args = []string{
		"--git-dir=" + tdir + "/.git",
		"merge",
		vit.Branch,
		"FETCH_HEAD",
	}
	cmd = exec.Command(vcsGitBin, args...)
	cmd.Dir = tdir
	if _, err := cmd.Output(); err != nil {
		return "", errors.New("failed to merge branch " + err.Error())
	}

	//
	args = []string{
		"--git-dir=" + tdir + "/.git",
		"log",
		"--format=%H",
		"-n",
		"1",
	}
	out, err := exec.Command(vcsGitBin, args...).Output()
	if err != nil {
		return "", errors.New("failed to get last log id " + err.Error())
	}

	ver := strings.TrimSpace(string(out))
	if VcsGitVerReg.MatchString(ver) {
		return ver, nil
	}

	return "", errors.New("fail to get last log id")
}
