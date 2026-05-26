// Command singboxconv converts a sing-box OR Xray JSON config into URI
// subscription format that xray-knife understands.
//
// FORMATS HANDLED:
//
//   sing-box (https://sing-box.sagernet.org):
//     {"outbounds":[{"type":"vless","server":"...","server_port":443,
//                    "uuid":"...","tls":{...},"transport":{"type":"ws",...}}]}
//
//   Xray (https://xtls.github.io):
//     {"outbounds":[{"protocol":"vless","settings":{"vnext":[{
//                    "address":"...","port":443,
//                    "users":[{"id":"...","encryption":"none","flow":"..."}]}]},
//                    "streamSettings":{"network":"ws","security":"tls",...}}]}
//
// Auto-detection per-outbound: if `type` field is present we treat as sing-box,
// if `protocol` is present we treat as Xray. Falls through to skip.
//
// Input: stdin OR --url <http-url> OR --file <path>
// Output: one URI per line on stdout — vless:// / vmess:// / trojan:// / ss://
//
// Skipped silently:
//   - Aggregator outbounds: selector, urltest, freedom, blackhole, dns, loopback
//   - Unsupported protocols: hysteria/hysteria2, tuic, wireguard
//   - Shadowsocks with v2ray-plugin (xray-knife doesn't speak it)
//
// Usage:
//
//	singboxconv --url https://sub.pai.yt/singbox > paiyt.txt
//	curl -sSL https://example/xray.json | singboxconv > out.txt
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

const usage = `singboxconv — convert sing-box or Xray JSON config to URI lines.

Usage:
  singboxconv [--url <http-url> | --file <path>] [flags]
  singboxconv < config.json > out.txt

Flags:
  --url <url>         Fetch JSON from this URL instead of stdin
  --file <path>       Read JSON from this local file instead of stdin
  --timeout <secs>    HTTP timeout when fetching --url (default 60)
  --filter-protocol   Only emit URIs for this proto (vless,vmess,trojan,ss).
                      Empty (default) = all four supported protocols.
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

	// Generic decode — each outbound is a raw object so we can sniff its
	// shape (sing-box vs Xray) before structured unmarshal.
	var top struct {
		Outbounds []json.RawMessage `json:"outbounds"`
	}
	if err := json.Unmarshal(data, &top); err != nil {
		fmt.Fprintf(os.Stderr, "parse JSON: %v\n", err)
		return 1
	}

	emitted := 0
	skipped := 0
	for _, raw := range top.Outbounds {
		uri, ok := convertOutbound(raw, *filterProto)
		if !ok {
			skipped++
			continue
		}
		fmt.Println(uri)
		emitted++
	}
	fmt.Fprintf(os.Stderr, "singboxconv: emitted=%d skipped=%d total=%d\n",
		emitted, skipped, len(top.Outbounds))
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
	return io.ReadAll(io.LimitReader(resp.Body, 16<<20))
}

// convertOutbound dispatches on format. Sniffs the object: presence of
// `type` -> sing-box; presence of `protocol` -> Xray. Returns "", false if
// neither matches or the protocol is unsupported.
func convertOutbound(raw json.RawMessage, filterProto string) (string, bool) {
	var sniff struct {
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	}
	if err := json.Unmarshal(raw, &sniff); err != nil {
		return "", false
	}
	switch {
	case sniff.Type != "":
		return convertSingboxOutbound(raw, filterProto)
	case sniff.Protocol != "":
		return convertXrayOutbound(raw, filterProto)
	}
	return "", false
}

// ============================================================================
// SING-BOX FORMAT
// ============================================================================

type singboxOutbound struct {
	Type       string        `json:"type"`
	Tag        string        `json:"tag"`
	Server     string        `json:"server"`
	ServerPort int           `json:"server_port"`
	UUID       string        `json:"uuid"`
	Password   string        `json:"password"`
	Method     string        `json:"method"`
	AlterID    int           `json:"alter_id"`
	Security   string        `json:"security"`
	Flow       string        `json:"flow"`
	TLS        *tlsBlock     `json:"tls"`
	Transport  *transportBlk `json:"transport"`
	Plugin     string        `json:"plugin"`
	PluginOpts string        `json:"plugin_opts"`
}

type tlsBlock struct {
	Enabled    bool        `json:"enabled"`
	Insecure   bool        `json:"insecure"`
	ServerName string      `json:"server_name"`
	ALPN       []string    `json:"alpn"`
	UTLS       *utlsBlock  `json:"utls"`
	Reality    *realityBlk `json:"reality"`
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
	Type        string            `json:"type"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	ServiceName string            `json:"service_name"`
}

func convertSingboxOutbound(raw json.RawMessage, filterProto string) (string, bool) {
	var ob singboxOutbound
	if err := json.Unmarshal(raw, &ob); err != nil {
		return "", false
	}
	if filterProto != "" && !singboxMatchesProto(ob.Type, filterProto) {
		return "", false
	}
	if ob.Server == "" || ob.ServerPort == 0 {
		return "", false
	}
	switch ob.Type {
	case "vless":
		return singboxVlessURI(ob), true
	case "vmess":
		return singboxVmessURI(ob), true
	case "trojan":
		return singboxTrojanURI(ob), true
	case "shadowsocks":
		return singboxSsURI(ob)
	}
	return "", false
}

func singboxMatchesProto(sbType, want string) bool {
	want = strings.ToLower(want)
	switch sbType {
	case "vless", "vmess", "trojan":
		return sbType == want
	case "shadowsocks":
		return want == "ss" || want == "shadowsocks"
	}
	return false
}

func singboxVlessURI(ob singboxOutbound) string {
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

func singboxVmessURI(ob singboxOutbound) string {
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

func singboxTrojanURI(ob singboxOutbound) string {
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

func singboxSsURI(ob singboxOutbound) (string, bool) {
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

// ============================================================================
// XRAY FORMAT
// ============================================================================

type xrayOutbound struct {
	Tag            string        `json:"tag"`
	Protocol       string        `json:"protocol"`
	Settings       *xraySettings `json:"settings"`
	StreamSettings *xrayStream   `json:"streamSettings"`
}

type xraySettings struct {
	Vnext   []xrayVnext  `json:"vnext"`
	Servers []xrayServer `json:"servers"`
}

type xrayVnext struct {
	Address string     `json:"address"`
	Port    int        `json:"port"`
	Users   []xrayUser `json:"users"`
}

type xrayUser struct {
	ID         string `json:"id"`
	AlterID    int    `json:"alterId"`
	Security   string `json:"security"`
	Encryption string `json:"encryption"`
	Flow       string `json:"flow"`
}

type xrayServer struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Method   string `json:"method"`
	Password string `json:"password"`
}

type xrayStream struct {
	Network         string                `json:"network"`
	Security        string                `json:"security"`
	TLSSettings     *xrayTLS              `json:"tlsSettings"`
	RealitySettings *xrayReality          `json:"realitySettings"`
	WSSettings      *xrayWS               `json:"wsSettings"`
	GRPCSettings    *xrayGRPC             `json:"grpcSettings"`
	TCPSettings     *xrayTCP              `json:"tcpSettings"`
	HTTPSettings    *xrayHTTP             `json:"httpSettings"`
}

type xrayTLS struct {
	ServerName  string   `json:"serverName"`
	Fingerprint string   `json:"fingerprint"`
	ALPN        []string `json:"alpn"`
}

type xrayReality struct {
	ServerName  string `json:"serverName"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"publicKey"`
	ShortID     string `json:"shortId"`
}

type xrayWS struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

type xrayGRPC struct {
	ServiceName string `json:"serviceName"`
	MultiMode   bool   `json:"multiMode"`
}

type xrayTCP struct {
	Header *xrayHeader `json:"header"`
}

type xrayHeader struct {
	Type string `json:"type"`
}

type xrayHTTP struct {
	Path string   `json:"path"`
	Host []string `json:"host"`
}

func convertXrayOutbound(raw json.RawMessage, filterProto string) (string, bool) {
	var ob xrayOutbound
	if err := json.Unmarshal(raw, &ob); err != nil {
		return "", false
	}
	proto := strings.ToLower(ob.Protocol)
	// Skip aggregator/local outbounds
	switch proto {
	case "freedom", "blackhole", "dns", "loopback":
		return "", false
	}
	if filterProto != "" && !xrayMatchesProto(proto, filterProto) {
		return "", false
	}
	switch proto {
	case "vless":
		return xrayVlessURI(ob)
	case "vmess":
		return xrayVmessURI(ob)
	case "trojan":
		return xrayTrojanURI(ob)
	case "shadowsocks":
		return xraySsURI(ob)
	}
	return "", false
}

func xrayMatchesProto(p, want string) bool {
	want = strings.ToLower(want)
	switch p {
	case "vless", "vmess", "trojan":
		return p == want
	case "shadowsocks":
		return want == "ss" || want == "shadowsocks"
	}
	return false
}

// xrayBuildStreamQuery extracts transport + tls/reality params from
// streamSettings into a url.Values that's shared across vless/trojan URIs.
func xrayBuildStreamQuery(stream *xrayStream) url.Values {
	q := url.Values{}
	if stream == nil {
		q.Set("type", "tcp")
		q.Set("security", "none")
		return q
	}
	network := stream.Network
	if network == "" {
		network = "tcp"
	}
	security := stream.Security
	if security == "" {
		security = "none"
	}
	q.Set("type", network)
	q.Set("security", security)

	switch network {
	case "ws":
		if stream.WSSettings != nil {
			if stream.WSSettings.Path != "" {
				q.Set("path", stream.WSSettings.Path)
			}
			if h, ok := stream.WSSettings.Headers["Host"]; ok && h != "" {
				q.Set("host", h)
			} else if h, ok := stream.WSSettings.Headers["host"]; ok && h != "" {
				q.Set("host", h)
			}
		}
	case "grpc":
		if stream.GRPCSettings != nil {
			if stream.GRPCSettings.ServiceName != "" {
				q.Set("serviceName", stream.GRPCSettings.ServiceName)
			}
			if stream.GRPCSettings.MultiMode {
				q.Set("mode", "multi")
			} else {
				q.Set("mode", "gun")
			}
		}
	case "tcp":
		if stream.TCPSettings != nil && stream.TCPSettings.Header != nil && stream.TCPSettings.Header.Type != "" {
			q.Set("headerType", stream.TCPSettings.Header.Type)
		}
	case "h2", "http":
		if stream.HTTPSettings != nil {
			if stream.HTTPSettings.Path != "" {
				q.Set("path", stream.HTTPSettings.Path)
			}
			if len(stream.HTTPSettings.Host) > 0 {
				q.Set("host", strings.Join(stream.HTTPSettings.Host, ","))
			}
		}
	}

	switch security {
	case "tls":
		if stream.TLSSettings != nil {
			if stream.TLSSettings.ServerName != "" {
				q.Set("sni", stream.TLSSettings.ServerName)
			}
			if stream.TLSSettings.Fingerprint != "" {
				q.Set("fp", stream.TLSSettings.Fingerprint)
			}
			if len(stream.TLSSettings.ALPN) > 0 {
				q.Set("alpn", strings.Join(stream.TLSSettings.ALPN, ","))
			}
		}
	case "reality":
		if stream.RealitySettings != nil {
			if stream.RealitySettings.ServerName != "" {
				q.Set("sni", stream.RealitySettings.ServerName)
			}
			if stream.RealitySettings.Fingerprint != "" {
				q.Set("fp", stream.RealitySettings.Fingerprint)
			}
			if stream.RealitySettings.PublicKey != "" {
				q.Set("pbk", stream.RealitySettings.PublicKey)
			}
			if stream.RealitySettings.ShortID != "" {
				q.Set("sid", stream.RealitySettings.ShortID)
			}
		}
	}
	return q
}

func xrayVlessURI(ob xrayOutbound) (string, bool) {
	if ob.Settings == nil || len(ob.Settings.Vnext) == 0 {
		return "", false
	}
	sv := ob.Settings.Vnext[0]
	if sv.Address == "" || sv.Port == 0 || len(sv.Users) == 0 {
		return "", false
	}
	user := sv.Users[0]
	q := xrayBuildStreamQuery(ob.StreamSettings)
	enc := user.Encryption
	if enc == "" {
		enc = "none"
	}
	q.Set("encryption", enc)
	if user.Flow != "" {
		q.Set("flow", user.Flow)
	}
	frag := safeFragment(ob.Tag)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s",
		user.ID, sv.Address, sv.Port, q.Encode(), frag), true
}

func xrayVmessURI(ob xrayOutbound) (string, bool) {
	if ob.Settings == nil || len(ob.Settings.Vnext) == 0 {
		return "", false
	}
	sv := ob.Settings.Vnext[0]
	if sv.Address == "" || sv.Port == 0 || len(sv.Users) == 0 {
		return "", false
	}
	user := sv.Users[0]
	q := xrayBuildStreamQuery(ob.StreamSettings)

	v := map[string]any{
		"v":    "2",
		"ps":   ob.Tag,
		"add":  sv.Address,
		"port": fmt.Sprintf("%d", sv.Port),
		"id":   user.ID,
		"aid":  fmt.Sprintf("%d", user.AlterID),
		"net":  q.Get("type"),
		"type": q.Get("headerType"),
		"host": q.Get("host"),
		"path": q.Get("path"),
		"tls":  q.Get("security"),
		"sni":  q.Get("sni"),
		"alpn": q.Get("alpn"),
		"fp":   q.Get("fp"),
	}
	if user.Security != "" {
		v["scy"] = user.Security
	} else {
		v["scy"] = "auto"
	}
	if v["tls"] == "none" {
		v["tls"] = ""
	}
	body, _ := json.Marshal(v)
	return "vmess://" + base64.StdEncoding.EncodeToString(body), true
}

func xrayTrojanURI(ob xrayOutbound) (string, bool) {
	if ob.Settings == nil || len(ob.Settings.Servers) == 0 {
		return "", false
	}
	sv := ob.Settings.Servers[0]
	if sv.Address == "" || sv.Port == 0 || sv.Password == "" {
		return "", false
	}
	q := xrayBuildStreamQuery(ob.StreamSettings)
	frag := safeFragment(ob.Tag)
	return fmt.Sprintf("trojan://%s@%s:%d?%s#%s",
		url.QueryEscape(sv.Password), sv.Address, sv.Port, q.Encode(), frag), true
}

func xraySsURI(ob xrayOutbound) (string, bool) {
	if ob.Settings == nil || len(ob.Settings.Servers) == 0 {
		return "", false
	}
	sv := ob.Settings.Servers[0]
	if sv.Address == "" || sv.Port == 0 || sv.Method == "" {
		return "", false
	}
	userinfo := sv.Method + ":" + sv.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(userinfo))
	frag := safeFragment(ob.Tag)
	return fmt.Sprintf("ss://%s@%s:%d#%s",
		encoded, sv.Address, sv.Port, frag), true
}

// ============================================================================
// SHARED HELPERS
// ============================================================================

func safeFragment(tag string) string {
	tag = strings.ReplaceAll(tag, "\n", " ")
	tag = strings.ReplaceAll(tag, "\r", " ")
	if tag == "" {
		tag = "node"
	}
	return url.QueryEscape(tag)
}
