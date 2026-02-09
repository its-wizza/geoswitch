package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"geoswitch/internal/proxy"
)

func loadConfig() proxy.Config {
	alProxy, _ := url.Parse("http://localhost:8081")
	deProxy, _ := url.Parse("http://localhost:8082")
	return proxy.Config{
		Exits: map[string]*proxy.Exit{
			"default": {Name: "default", Type: proxy.ExitDirect},
			"albania": {Name: "albania", Type: proxy.ExitHTTPProxy, ProxyURL: alProxy},
			"germany": {Name: "germany", Type: proxy.ExitHTTPProxy, ProxyURL: deProxy},
		},
		Rules: []proxy.Rule{
			{Name: "al-tld", Matcher: proxy.TLDMatcher{TLD: "al"}, ExitName: "albania"},
			{Name: "de-host", Matcher: proxy.HostEqualsMatcher{Host: "example.de"}, ExitName: "germany"},
		},
	}
}

func main() {
	// Start test servers first
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Println("Albania exit listening on :8081")
		muxAl := http.NewServeMux()
		muxAl.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("ALBANIA [%s] %s %s Host:%s", r.RemoteAddr, r.Method, r.URL.Path, r.Header.Get("Host"))
			w.Header().Set("X-Exit", "albania")
			io.WriteString(w, "ALBANIA EXIT")
		})
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	go func() {
		defer wg.Done()
		log.Println("Germany exit listening on :8082")
		muxDe := http.NewServeMux()
		muxDe.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("GERMANY [%s] %s %s Host:%s", r.RemoteAddr, r.Method, r.URL.Path, r.Header.Get("Host"))
			w.Header().Set("X-Exit", "germany")
			io.WriteString(w, "GERMANY EXIT")
		})
		log.Fatal(http.ListenAndServe(":8082", nil))
	}()

	// Give servers 1s to start
	time.Sleep(1 * time.Second)

	// Now start your main proxy
	cfg := loadConfig()

	runtimes := make(map[string]*proxy.ExitRuntime)
	for name, exit := range cfg.Exits {
		runtimes[name] = proxy.BuildExitRuntime(exit)
	}

	selector := &proxy.ExitSelector{
		Cfg:      &cfg,
		Runtimes: runtimes,
	}
	revProxy := proxy.NewReverseProxy(selector)

	srv := &proxy.ProxyServer{
		Proxy: revProxy,
	}

	log.Println("Main proxy listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", srv))
}
