package proxy

import (
	"fmt"
	"net/http"
)

type ExitSelector struct {
	Cfg      *Config
	Runtimes map[string]*ExitRuntime
}

func (es *ExitSelector) RoundTrip(req *http.Request) (*http.Response, error) {
	exit := ChooseExit(es.Cfg, req)
	rt, ok := es.Runtimes[exit.Name]
	if !ok {
		return nil, fmt.Errorf("no runtime for exit %q", exit.Name)
	}
	return rt.Transport.RoundTrip(req)
}
