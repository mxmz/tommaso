package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"mxmz.it/mxmz/tommaso/dto"
	"mxmz.it/mxmz/tommaso/prober"
	"mxmz.it/mxmz/tommaso/system"
)

func main() {

	var baseURL = "http://localhost:7997"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}
	var p = NewCachedProber()
	var probed = 0
	for {

		for _, i := range system.GetNetInterfaceAddresses() {
			var ifaces = []string{i}
			var specs, err = getMyProbeTestingSpecs(ifaces, baseURL)
			fmt.Printf("getMyProbeSpecs: %s %s: err = %v\n", i, baseURL, err)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			rv := p.RunProbSpecsConcurrent(ifaces, specs)
			//	var _ = err
			var _ = rv
			fmt.Printf("probe results = %v (%d)\n", len(rv), p.Probed)
			if p.Probed != probed {
				err = pushMyProbeResults(baseURL, rv)
				fmt.Printf("pushMyProbeResults: err = %v\n", err)
			}
			probed = p.Probed
		}
		fmt.Println("sleep ...")
		time.Sleep(5 * time.Second)
	}

}

func jsonIndent(v interface{}) string {
	b, _ := json.MarshalIndent(v, " ", " ")
	return string(b)
}

func getMyProbeTestingSpecs(ifaces []string, baseURL string) ([]*dto.ProbeTestingSpec, error) {

	//var ifaces = system.GetNetInterfaceAddresses()
	var body = dto.MySources{
		Sources: ifaces,
	}

	var res, err = http.DefaultClient.Post(baseURL+"/api/agent/get-my-probe-specs", "application/json", strings.NewReader(jsonIndent(body)))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var probeSpecs []*dto.ProbeTestingSpec
	err = json.Unmarshal(resBody, &probeSpecs)
	if err != nil {
		return nil, err
	}

	return probeSpecs, nil

}

func pushMyProbeResults(baseURL string, results []*dto.ProbeResult) error {

	var res, err = http.DefaultClient.Post(baseURL+"/api/agent/push-my-probe-results", "application/json", strings.NewReader(jsonIndent(results)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusAccepted {
		return errors.New(fmt.Sprintf("Unexpected status: %d", res.StatusCode))
	}

	return nil

}

type CachedProber struct {
	cache  map[string]*dto.ProbeResult
	lock   sync.RWMutex
	Probed int
	p      *prober.Prober
}

func NewCachedProber() *CachedProber {
	return &CachedProber{
		cache: map[string]*dto.ProbeResult{},
		p:     prober.NewProber(),
	}
}

var FailCacheTTL = 20 * time.Second
var OKCacheTTL = 3 * 60 * time.Second

func (p *CachedProber) cached(spec *dto.ProbeTestingSpec) *dto.ProbeResult {
	var k = spec.Type + strings.Join(spec.Args, ":")
	p.lock.RLock()
	defer p.lock.RUnlock()
	var now = time.Now()
	var v, ok = p.cache[k]
	if ok {
		if v.Status == "FAIL" && now.Sub(v.Time) < FailCacheTTL {
			return v
		}
		if v.Status == "OK" && now.Sub(v.Time) < OKCacheTTL {
			return v
		}
	}
	return nil
}
func (p *CachedProber) setCache(res *dto.ProbeResult) {
	var spec = res.Spec
	var k = spec.Type + strings.Join(spec.Args, ":")
	p.lock.Lock()
	defer p.lock.Unlock()
	p.cache[k] = res
}

func (p *CachedProber) RunProbSpecsConcurrent(ifaces []string, specs []*dto.ProbeTestingSpec) []*dto.ProbeResult {
	var rv = make([]*dto.ProbeResult, 0, len(specs))
	var toprobe = make([]*dto.ProbeTestingSpec, 0, len(specs))

	for _, s := range specs {
		if cached := p.cached(s); cached != nil {
			rv = append(rv, cached)
		} else {
			toprobe = append(toprobe, s)
			p.Probed++
		}
	}
	for _, r := range p.p.RunProbSpecsConcurrent(ifaces, toprobe) {
		p.setCache(r)
		rv = append(rv, r)
	}

	return rv
}
