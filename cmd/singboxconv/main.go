// Command singboxconv converts a sing-box client JSON config into the URI
// subscription format that xray-knife understands.
//
// Input  : sing-box JSON (stdin OR --url <http://...> OR --file <path>)
// Output : one URI per line on stdout — vless:// / vmess:// / trojan:// / ss://
//
// We only emit URIs for outbounds whose `type` matches the schemes that
// vlessfilter supports. Selector/urltest/direct/dns/block outbounds are
// skipped silently (they're routing aggregators, not actual proxies).
//
// Sing-box outbounds are richer than URI fields (e.g. `reality.public_key`,
// `utls.fingerprint`) — we round-trip whatever maps cleanly and drop the rest.
// xray-knife will still test the resulting URI; if a TLS detail was lost,
// the proxy may fail handshake, which the pipeline treats as "dead" anyway.
//
// Usage:
//
//	singboxconv --url https://sub.pai.yt/singbox > paiyt.txt
//	curl -sSL https://sub.pai.yt/singbox | singboxconv > paiyt.txt
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const usage = `singboxconv — convert sing-box JSON config to vlessfilter URI lines.

Usage:
  singboxconv [--url <http-url> | --file <path>] [flags]
  singboxconv < config.json > out.txt

Flags:
  --url <url>         Fetch JSON from this URL instead of stdin
  --file <path>       Read JSON from this local file instead of stdin
  --timeout <secs>    HTTP timeout when fetching --url (default 60)
  --keep-tag          Use the outbound's "tag" as URI fragment (default: yes)
  --filter-protocol   Only emit URIs for this proto (vless,vmess,trojan,ss).
                      Empty (default) = all four supported protocols.
  --skip-bad          Skip outbounds we can't convert (default; otherwise log+skip)
`

func main() {
	os.Exit(run())
}

func run() int {
	urlFlag := flag.String("url", "", "Fetch JSON from this URL")
	fileFlag := flag.String("file", "", "Read JSON from this local file")
	timeout := flag.Int("timeout", 60, "HTTP timeout in seconds")
	filterProto := flag.String("filter-protocol", "", "Only emit URIs for this protocol")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	var data []byte
	var err error
	switch {
	case *urlFlag != "":
		data, err = fetchURL(*urlFlag, time.Duration(*timeout)*time.Second)
	case *fileFlag != "":
		data, err = os.ReadFile(*fileFlag)
	default:
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "input read failed: %v\n", err)
		return 1
	}

	var cfg singboxConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "parse JSON: %v\n", err)
		return 1
	}

	emitted := 0
	skipped := 0
	for _, ob := range cfg.Outbounds {
		if *filterProto != "" && !matchesProto(ob.Type, *filterProto) {
			continue
		}
		uri, ok := outboundToURI(ob)
		if !ok {
			skipped++
			continue
		}
		fmt.Println(uri)
		emitted++
	}
	fmt.Fprintf(os.Stderr, "singboxconv: emitted=%d skipped=%d total=%d\n",
		emitted, skipped, len(cfg.Outbounds))
	return 0
}

func fetchURL(rawURL string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "vlessfilter-singboxconv/1.0")
	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 16<<20)) // 16 MiB
}

// singboxConfig captures only the parts of a sing-box config we care about.
// Unknown fields are silently dropped by encoding/json's default behavior.
type singboxConfig struct {
	Outbounds []outbound `json:"outbounds"`
}

type outbound struct {
	Type       string         `json:"type"`
	Tag        string         `json:"tag"`
	Server     string         `json:"server"`
	ServerPort int            `json:"server_port"`
	UUID       string         `json:"uuid"`
	Password   string         `json:"password"`
	Method     string         `json:"method"`     // ss
	AlterID    int            `json:"alter_id"`   // vmess
	Security   string         `json:"security"`   // vmess: auto/aes-128-gcm
	Flow       string         `json:"flow"`       // vless: xtls-rprx-vision
	TLS        *tlsBlock      `json:"tls"`
	Transport  *transportBlk  `json:"transport"`
	Plugin     string         `json:"plugin"`      // ss
	PluginOpts string         `json:"plugin_opts"` // ss
}

type tlsBlock struct {
	Enabled    bool         `json:"enabled"`
	Insecure   bool         `json:"insecure"`
	ServerName string       `json:"server_name"`
	ALPN       []string     `json:"alpn"`
	UTLS       *utlsBlock   `json:"utls"`
	Reality    *realityBlk  `json:"reality"`
}

type utlsBlock struct {
	Enabled     bool   `json:"enabled"`
	Fingerprint string `json:"fingerprint"`
}

type realityBlk struct {
	Enabled   bool   `json:"enabled"`
	PublicKey string `json:"public_key"`
	ShortID   string `json:"short_id"`
}

type transportBlk struct {
	Type    string            `json:"type"` // ws / grpc / http / quic
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	// gRPC-specific
	ServiceName string `json:"service_name"`
}

func matchesProto(sbType, want string) bool {
	want = strings.ToLower(want)
	switch sbType {
	case "vless", "vmess", "trojan":
		return sbType == want
	case "shadowsocks":
		return want == "ss" || want == "shadowsocks"
	}
	return false
}

func outboundToURI(ob outbound) (string, bool) {
	if ob.Server == "" || ob.ServerPort == 0 {
		return "", false
	}
	switch ob.Type {
	case "vless":
		return vlessURI(ob), true
	case "vmess":
		return vmessURI(ob), true
	case "trojan":
		return trojanURI(ob), true
	case "shadowsocks":
		return ssURI(ob)
	}
	return "", false
}

func vlessURI(ob outbound) string {
	q := url.Values{}
	q.Set("type", "tcp")
	if ob.Transport != nil {
		switch ob.Transport.Type {
		case "ws":
			q.Set("type", "ws")
			if ob.Transport.Path != "" {
				q.Set("path", ob.Transport.Path)
			}
			if h, ok := ob.Transport.Headers["Host"]; ok && h != "" {
				q.Set("host", h)
			}
		case "grpc":
			q.Set("type", "grpc")
			if ob.Transport.ServiceName != "" {
				q.Set("serviceName", ob.Transport.ServiceName)
			}
		case "http":
			q.Set("type", "http")
		}
	}
	if ob.TLS != nil && ob.TLS.Enabled {
		if ob.TLS.Reality != nil && ob.TLS.Reality.Enabled {
			q.Set("security", "reality")
			if ob.TLS.Reality.PublicKey != "" {
				q.Set("pbk", ob.TLS.Reality.PublicKey)
			}
			if ob.TLS.Reality.ShortID != "" {
				q.Set("sid", ob.TLS.Reality.ShortID)
			}
		} else {
			q.Set("security", "tls")
		}
		if ob.TLS.ServerName != "" {
			q.Set("sni", ob.TLS.ServerName)
		}
		if ob.TLS.UTLS != nil && ob.TLS.UTLS.Enabled && ob.TLS.UTLS.Fingerprint != "" {
			q.Set("fp", ob.TLS.UTLS.Fingerprint)
		}
	}
	if ob.Flow != "" {
		q.Set("flow", ob.Flow)
	}
	frag := safeFragment(ob.Tag)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		ob.UUID, ob.Server, ob.ServerPort, q.Encode(), frag)
}

func vmessURI(ob outbound) string {
	// vmess uses a base64-JSON URI per v2rayN convention.
	v := map[string]any{
		"v":    "2",
		"ps":   ob.Tag,
		"add":  ob.Server,
		"port": fmt.Sprintf("%d", ob.ServerPort),
		"id":   ob.UUID,
		"aid":  fmt.Sprintf("%d", ob.AlterID),
		"net":  "tcp",
		"type": "none",
		"tls":  "",
	}
	if ob.Security != "" {
		v["scy"] = ob.Security
	}
	if ob.Transport != nil {
		switch ob.Transport.Type {
		case "ws":
			v["net"] = "ws"
			if ob.Transport.Path != "" {
				v["path"] = ob.Transport.Path
			}
			if h, ok := ob.Transport.Headers["Host"]; ok && h != "" {
				v["host"] = h
			}
		case "grpc":
			v["net"] = "grpc"
			if ob.Transport.ServiceName != "" {
				v["path"] = ob.Transport.ServiceName
			}
		}
	}
	if ob.TLS != nil && ob.TLS.Enabled {
		v["tls"] = "tls"
		if ob.TLS.ServerName != "" {
			v["sni"] = ob.TLS.ServerName
		}
	}
	body, _ := json.Marshal(v)
	return "vmess://" + base64.StdEncoding.EncodeToString(body)
}

func trojanURI(ob outbound) string {
	q := url.Values{}
	q.Set("security", "tls")
	if ob.TLS != nil && ob.TLS.Enabled {
		if ob.TLS.ServerName != "" {
			q.Set("sni", ob.TLS.ServerName)
		}
		if ob.TLS.Insecure {
			q.Set("allowInsecure", "1")
		}
	}
	q.Set("type", "tcp")
	if ob.Transport != nil {
		switch ob.Transport.Type {
		case "ws":
			q.Set("type", "ws")
			if ob.Transport.Path != "" {
				q.Set("path", ob.Transport.Path)
			}
			if h, ok := ob.Transport.Headers["Host"]; ok && h != "" {
				q.Set("host", h)
			}
		case "grpc":
			q.Set("type", "grpc")
			if ob.Transport.ServiceName != "" {
				q.Set("serviceName", ob.Transport.ServiceName)
			}
		}
	}
	frag := safeFragment(ob.Tag)
	return fmt.Sprintf("trojan://%s@%s:%d?%s#%s",
		url.QueryEscape(ob.Password), ob.Server, ob.ServerPort, q.Encode(), frag)
}

// ssURI emits ss://method:password@host:port#tag with userinfo base64-encoded
// (the v2rayN/SIP002 format). xray-knife and most clients accept this.
//
// Skips configs with v2ray-plugin since the URI form for those is
// implementation-specific and tends to break our pipeline.
func ssURI(ob outbound) (string, bool) {
	if ob.Plugin != "" {
		return "", false
	}
	method := ob.Method
	if method == "" || method == "none" {
		return "", false
	}
	userinfo := method + ":" + ob.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(userinfo))
	frag := safeFragment(ob.Tag)
	return fmt.Sprintf("ss://%s@%s:%s#%s",
		encoded, ob.Server, strconv.Itoa(ob.ServerPort), frag), true
}

// safeFragment URL-escapes the tag for use as the URI fragment. Newlines
// and CRs would corrupt the line-per-URI output.
func safeFragment(tag string) string {
	tag = strings.ReplaceAll(tag, "\n", " ")
	tag = strings.ReplaceAll(tag, "\r", " ")
	return url.QueryEscape(tag)
}
