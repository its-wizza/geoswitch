package proxy

import (
	"net/http"
	"net/url"
)

type ExitType string

const (
	ExitDirect      ExitType = "direct"
	ExitHTTPProxy   ExitType = "http_proxy"
	ExitSOCKS5Proxy ExitType = "socks5_proxy"
)

type Exit struct {
	Name     string
	Type     ExitType
	ProxyURL *url.URL
}

type Rule struct {
	Name     string
	Matcher  Matcher
	ExitName string
}

type Config struct {
	Exits map[string]*Exit
	Rules []Rule
}

func ChooseExit(cfg *Config, r *http.Request) *Exit {
	for _, rule := range cfg.Rules {
		if rule.Matcher.Match(r) {
			if exit, ok := cfg.Exits[rule.ExitName]; ok {
				return exit
			}
		}
	}
	return cfg.Exits["default"]
}
