package flaresolverr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/flaresolverr-gateway/solver/internal/model"
)

type Client struct {
	baseURL string
	hc      *http.Client
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, hc: &http.Client{Timeout: 120 * time.Second}}
}

func (c *Client) Solve(req model.SolveRequest) (*model.FlareSolverrResult, error) {
	timeout := req.Timeout
	if timeout <= 0 { timeout = 60000 }
	if timeout > 120000 { timeout = 120000 }

	fsReq := model.FSRequest{Cmd: "request.get", URL: req.URL, MaxTimeout: timeout}
	if req.Proxy != "" { fsReq.Proxy = &model.FSProxy{URL: req.Proxy} }
	if req.Mode == "turnstile" {
		if req.TabsTillVerify <= 0 { req.TabsTillVerify = 2 }
		fsReq.TabsTillVerify = req.TabsTillVerify
	}
	if req.WaitAfter > 0 { fsReq.WaitInSeconds = req.WaitAfter }
	if len(req.Actions) > 0 { fsReq.Actions = req.Actions }

	if req.SiteKey != "" {
		actions := []model.Action{
			{Type: "wait", WaitMs: 1500},
			{Type: "js", Value: "var d=document.createElement('div');d.className='cf-turnstile';d.setAttribute('data-sitekey','" + req.SiteKey + "');d.style.cssText='position:fixed;top:10px;left:10px;z-index:99999;background:white;padding:20px';document.body.appendChild(d);var i=document.createElement('input');i.type='hidden';i.name='cf-turnstile-response';document.body.appendChild(i);var s=document.createElement('script');s.src='https://challenges.cloudflare.com/turnstile/v0/api.js';s.onload=function(){window.turnstile.render(d,{sitekey:'" + req.SiteKey + "',callback:function(t){i.value=t;}});};document.head.appendChild(s);"},
			{Type: "wait", WaitMs: 12000},
		}
		fsReq.Actions = append(fsReq.Actions, actions...)
	}

	body, _ := json.Marshal(fsReq)
	httpReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/v1", c.baseURL), bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(httpReq)
	if err != nil { return nil, fmt.Errorf("request: %w", err) }
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("FS returned %d: %s", resp.StatusCode, string(respBody)) }

	var fsResp model.FSResponse
	if err := json.Unmarshal(respBody, &fsResp); err != nil { return nil, fmt.Errorf("unmarshal: %w", err) }
	if fsResp.Status != "ok" { return nil, fmt.Errorf("FS error: %s", fsResp.Message) }

	cookies := make([]model.Cookie, len(fsResp.Solution.Cookies))
	for i, ck := range fsResp.Solution.Cookies {
		cookies[i] = model.Cookie{Name: ck.Name, Value: ck.Value, Domain: ck.Domain, Path: ck.Path, Expires: ck.Expires, HTTPOnly: ck.HTTPOnly, Secure: ck.Secure, SameSite: ck.SameSite}
	}
	headers := make(map[string]string)
	for k, v := range fsResp.Solution.Headers { headers[k] = fmt.Sprint(v) }

	return &model.FlareSolverrResult{
		URL: fsResp.Solution.URL, Status: fsResp.Solution.Status,
		Cookies: cookies, UserAgent: fsResp.Solution.UserAgent,
		TurnstileToken: fsResp.Solution.TurnstileToken,
		Response: fsResp.Solution.Response, Headers: headers,
	}, nil
}
