package prober

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"log"

	"mxmz.it/mxmz/tommaso/dto"
	"mxmz.it/mxmz/tommaso/system"
)

var FailCacheTTL = 20 * time.Second
var OKCacheTTL = 60 * time.Second

var probers = map[string]func(ifaces []string, spec *dto.ProbeSpec) *dto.ProbeResult{
	"tcp": tcpProbe,
}

type Prober struct {
	cache  map[string]*dto.ProbeResult
	lock   sync.RWMutex
	Probed int
}

func NewProber() *Prober {
	return &Prober{
		cache: map[string]*dto.ProbeResult{},
	}
}

func (p *Prober) cached(spec *dto.ProbeSpec) *dto.ProbeResult {
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
func (p *Prober) setCache(res *dto.ProbeResult) {
	var spec = res.Spec
	var k = spec.Type + strings.Join(spec.Args, ":")
	p.lock.Lock()
	defer p.lock.Unlock()
	p.cache[k] = res
	p.Probed++
}

func (p *Prober) RunProbSpecs(specs []*dto.ProbeSpec) []*dto.ProbeResult {
	ifaces := system.GetNetInterfaceAddresses()
	var rv = []*dto.ProbeResult{}
	for _, s := range specs {
		f, ok := probers[s.Type]
		if ok {
			var res = f(ifaces, s)
			rv = append(rv, res)
		}
	}
	return rv
}

func (p *Prober) RunProbSpecsConcurrent(specs []*dto.ProbeSpec) []*dto.ProbeResult {
	ifaces := system.GetNetInterfaceAddresses()
	var wg = sync.WaitGroup{}
	var rv = make([]*dto.ProbeResult, len(specs))
	for i, s := range specs {
		f, ok := probers[s.Type]
		if ok {
			var idx = i
			wg.Add(1)
			var spec = s
			go func() {
				var cached = p.cached(spec)
				if cached == nil {
					var res = f(ifaces, spec)
					rv[idx] = res
					p.setCache(res)
					log.Printf("probe %s %v = %s", spec.Type, spec.Args, res.Status)
				} else {
					rv[idx] = cached
				}
				wg.Done()

			}()
		}
	}
	wg.Wait()
	return rv
}

func tcpProbe(ifaces []string, spec *dto.ProbeSpec) *dto.ProbeResult {
	tm := time.Now()
	if spec.Args == nil || len(spec.Args) < 2 {
		return &dto.ProbeResult{
			Spec:    spec,
			Status:  "FAIL",
			Time:    tm,
			Elapsed: 0,
			Comment: "Bad arguments in probe spec",
			Sources: ifaces,
		}
	}

	rv, err := TcpProbe(fmt.Sprintf("%s:%s", spec.Args[0], spec.Args[1]), time.Duration(spec.Timeout)*time.Millisecond)

	if err != nil {
		return &dto.ProbeResult{
			Spec:    spec,
			Status:  "FAIL",
			Time:    tm,
			Elapsed: int(time.Now().Sub(tm) / time.Millisecond),
			Comment: err.Error(),
			Sources: ifaces,
		}
	}
	return &dto.ProbeResult{
		Spec:    spec,
		Status:  "OK",
		Time:    tm,
		Elapsed: int(rv / time.Millisecond),
		Comment: "",
		Sources: ifaces,
	}

}
