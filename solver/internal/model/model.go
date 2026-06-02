package model

import "time"

type SolveRequest struct {
	URL            string   `json:"url"`
	Proxy          string   `json:"proxy,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	Mode           string   `json:"mode,omitempty"`
	SiteKey        string   `json:"sitekey,omitempty"`
	TabsTillVerify int      `json:"tabs_till_verify,omitempty"`
	WaitAfter      int      `json:"wait_after,omitempty"`
	Actions        []Action `json:"actions,omitempty"`
}

type SolveResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type ResultResponse struct {
	TaskID   string              `json:"task_id"`
	Status   string              `json:"status"`
	Solution *FlareSolverrResult `json:"solution,omitempty"`
	Error    string              `json:"error,omitempty"`
	Duration int64               `json:"duration_ms,omitempty"`
}

type FlareSolverrResult struct {
	URL            string            `json:"url"`
	Status         int               `json:"status"`
	Cookies        []Cookie          `json:"cookies"`
	UserAgent      string            `json:"userAgent"`
	TurnstileToken string            `json:"turnstile_token,omitempty"`
	Response       string            `json:"response,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
}

type Cookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

type FSRequest struct {
	Cmd               string   `json:"cmd"`
	URL               string   `json:"url"`
	MaxTimeout        int      `json:"maxTimeout,omitempty"`
	Session           string   `json:"session,omitempty"`
	ReturnOnlyCookies bool     `json:"returnOnlyCookies,omitempty"`
	Proxy             *FSProxy `json:"proxy,omitempty"`
	Cookies           []Cookie `json:"cookies,omitempty"`
	TabsTillVerify    int      `json:"tabs_till_verify,omitempty"`
	WaitInSeconds     int      `json:"waitInSeconds,omitempty"`
	Actions           []Action `json:"actions,omitempty"`
}

type FSProxy struct {
	URL      string `json:"url"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type FSResponse struct {
	Status         string     `json:"status"`
	Message        string     `json:"message"`
	Solution       FSSolution `json:"solution"`
	StartTimestamp int64      `json:"startTimestamp"`
	EndTimestamp   int64      `json:"endTimestamp"`
}

type FSSolution struct {
	URL            string                 `json:"url"`
	Status         int                    `json:"status"`
	Headers        map[string]interface{} `json:"headers"`
	Response       string                 `json:"response"`
	Cookies        []FSCookie             `json:"cookies"`
	UserAgent      string                 `json:"userAgent"`
	TurnstileToken string                 `json:"turnstile_token,omitempty"`
}

type FSCookie struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HTTPOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}

type Action struct {
	Type     string `json:"type"`
	Selector string `json:"selector,omitempty"`
	Value    string `json:"value,omitempty"`
	WaitMs   int    `json:"wait_ms,omitempty"`
}

type Task struct {
	ID        string
	Status    string
	Req       SolveRequest
	Result    *FlareSolverrResult
	Error     string
	CreatedAt time.Time
	StartedAt time.Time
	EndedAt   time.Time
}
