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

// hpress REST API client shared by the CLI tools (cmd/clipper, cmd/cli).
//
// Auth model: the CLI holds an IAM user access key (ak_<id>_<secret>). For each
// request it mints a short-lived access-token JWT signed with the key's secret
// (HS256, Kid = access-key id, no Sub claim) and sends it as
// "Authorization: Bearer <token>". The hpress server resolves it via
// iamserver.AppVerifier.Resolve -> IAM /v2/open/app-auth/introspect. The
// access-key secret never leaves the CLI; only the signed token traverses the
// network.

// Package hpclient provides an access-key-authenticated client for the hpress
// /hp/v1 REST API.
package hpclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sysinner/innerstack/v2/pkg/inauth"

	"github.com/hooto/hpress/internal/hpapi"
)

// tokenLifetime is how long a minted access-key token is good for. It must
// comfortably exceed a CLI run (image / package uploads); IAM enforces no upper
// bound on user-type tokens, only that Exp > now.
const tokenLifetime = 3600 // seconds

// Client is an access-key-authenticated hpress API client.
type Client struct {
	baseURL string
	ak      *inauth.AccessKey
	http    *http.Client

	mu       sync.Mutex
	tokStr   string
	tokExpAt int64 // unix seconds
}

// NewClient parses an "ak_<id>_<secret>" access key into a Client.
func NewClient(baseURL, accessKey string) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("server base_url is not configured")
	}
	ak, err := inauth.ParseAccessKey(accessKey)
	if err != nil {
		return nil, fmt.Errorf("invalid access key: %w", err)
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		ak:      ak,
		http:    &http.Client{Timeout: 60 * time.Second},
	}, nil
}

// token returns a cached access-key-signed token, minting a fresh one when it is
// missing or within 60s of expiry.
func (c *Client) token() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().Unix()
	if c.tokStr != "" && c.tokExpAt-now > 60 {
		return c.tokStr, nil
	}

	// a one-key manager so inauth.AccessToken.SignToken can sign with our secret
	mgr := inauth.NewAccessKeyManager()
	if err := mgr.Set(&inauth.AccessKey{
		Id:     c.ak.Id,
		Secret: c.ak.Secret,
		Type:   "User",
		State:  inauth.AccessKey_State_Active,
	}); err != nil {
		return "", err
	}

	tok := inauth.NewAccessToken()
	tok.Header.Kid = c.ak.Id
	tok.Claims.Exp = now + tokenLifetime

	signed, err := tok.SignToken(mgr)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	if signed == "" {
		return "", fmt.Errorf("sign access token failed")
	}
	c.tokStr = signed
	c.tokExpAt = tok.Claims.Exp
	return c.tokStr, nil
}

// do sends a request (Bearer-authenticated), decodes the JSON response into out,
// and turns any hpress application-level error ({"error":{...}}) into a Go error.
func (c *Client) do(method, apiPath string, query url.Values, body []byte, out any) error {

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	u.Path = strings.TrimRight(u.Path, "/") + apiPath
	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	tok, err := c.token()
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := checkAPIError(respBytes, resp.StatusCode); err != nil {
		return err
	}
	if out != nil {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("decode response: %w (body: %s)", err, truncBody(respBytes))
		}
	}
	return nil
}

// SpecGet fetches a module's full spec (nodeModels, termModels, ...).
func (c *Client) SpecGet(modname string) (*hpapi.Spec, error) {
	var rsp hpapi.Spec
	if err := c.do(http.MethodGet, "/hp/v1/mod-set/spec-entry",
		url.Values{"name": {modname}}, nil, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// TermList fetches the terms (categories/tags) for a module + termModel.
func (c *Client) TermList(modname, modelID string) (*hpapi.TermList, error) {
	var rsp hpapi.TermList
	if err := c.do(http.MethodGet, "/hp/v1/term/list",
		url.Values{"modname": {modname}, "modelid": {modelID}}, nil, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// NodeSet creates (empty id) or updates (set id) a node and returns the stored
// node (carrying the server-assigned id on create).
func (c *Client) NodeSet(modname, modelID string, node *hpapi.Node) (*hpapi.Node, error) {
	body, err := json.Marshal(node)
	if err != nil {
		return nil, err
	}
	var rsp hpapi.Node
	if err := c.do(http.MethodPost, "/hp/v1/node/set",
		url.Values{"modname": {modname}, "modelid": {modelID}}, body, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// S2Put uploads a file (raw bytes) to storage path "/deft/<date>/<file>".
func (c *Client) S2Put(storagePath string, data []byte) error {
	// The data-URL media type follows the stored extension. The server ignores it
	// for storage (it splits on ";base64," and decodes the tail), but the correct
	// type keeps the upload self-describing; SVG is stored verbatim (vector).
	mime := "image/jpeg"
	if strings.HasSuffix(storagePath, ".svg") {
		mime = "image/svg+xml"
	}
	req := hpapi.FsFile{
		Path:   storagePath,
		Encode: "base64",
		Body:   "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data),
	}
	body, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	return c.do(http.MethodPost, "/hp/v1/s2-obj/put", nil, body, nil)
}

// SpecUploadCommit uploads a module package (.ipk, raw bytes) to
// mod-set/spec-upload-commit. name must end in ".ipk" and len(data) must stay
// under the server's upload cap (8 MiB).
func (c *Client) SpecUploadCommit(name string, data []byte) error {
	req := hpapi.SpecUploadCommit{
		Size: int64(len(data)),
		Name: name,
		Data: "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(data),
	}
	body, err := json.Marshal(&req)
	if err != nil {
		return err
	}
	// on success the handler echoes the request struct back with kind "Spec"
	var rsp hpapi.Spec
	if err := c.do(http.MethodPost, "/hp/v1/mod-set/spec-upload-commit", nil, body, &rsp); err != nil {
		return err
	}
	if rsp.Kind != "Spec" {
		return fmt.Errorf("unexpected response kind %q", rsp.Kind)
	}
	return nil
}

// ApiError carries a hpress application-level error parsed from the response
// body ({"error":{"code","message"}}). It is returned as a concrete type so
// callers can branch on Code (e.g. an auth/access denial) without string
// matching the formatted message.
type ApiError struct {
	Code    string
	Message string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("hpress %s: %s", e.Code, e.Message)
}

// checkAPIError inspects a response body for an embedded hpress error.
func checkAPIError(body []byte, status int) error {
	var probe struct {
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &probe); err == nil && probe.Error != nil && probe.Error.Code != "" {
		return &ApiError{Code: probe.Error.Code, Message: probe.Error.Message}
	}
	if status >= 400 {
		return fmt.Errorf("hpress HTTP %d: %s", status, truncBody(body))
	}
	return nil
}

func truncBody(b []byte) string {
	if len(b) > 300 {
		return string(b[:300]) + "..."
	}
	return string(b)
}
