package proxy

import (
	"net/http"
	"strings"
)

type Matcher interface {
	Match(r *http.Request) bool
}

type AllMatcher struct {
	Matchers []Matcher
}

func (m AllMatcher) Match(r *http.Request) bool {
	for _, mm := range m.Matchers {
		if !mm.Match(r) {
			return false
		}
	}
	return true
}

type AnyMatcher struct {
	Matchers []Matcher
}

func (m AnyMatcher) Match(r *http.Request) bool {
	for _, mm := range m.Matchers {
		if mm.Match(r) {
			return true
		}
	}
	return false
}

type HostEqualsMatcher struct {
	Host string
}

func (m HostEqualsMatcher) Match(r *http.Request) bool {
	return strings.EqualFold(r.Host, m.Host)
}

type HeaderEqualsMatcher struct {
	Header string
	Value  string
}

func (m HeaderEqualsMatcher) Match(r *http.Request) bool {
	return r.Header.Get(m.Header) == m.Value
}

type TLDMatcher struct {
	TLD string
}

func (m TLDMatcher) Match(r *http.Request) bool {
	host := r.URL.Hostname()
	labels := strings.Split(host, ".")
	if len(labels) < 2 {
		return false
	}
	return labels[len(labels)-1] == m.TLD
}

type PathPrefixMatcher struct {
	Prefix string
}

func (m PathPrefixMatcher) Match(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, m.Prefix)
}
