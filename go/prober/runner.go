package prober

import (
	"fmt"
	"sync"
	"time"

	"log"

	"mxmz.it/mxmz/tommaso/dto"
	"mxmz.it/mxmz/tommaso/system"
)

var probers = map[string]func(ifaces []string, spec *dto.ProbeSpec) *dto.ProbeResult{
	"tcp": tcpProbe,
}

type Prober struct {
}

func NewProber() *Prober {
	return &Prober{}
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
				var res = f(ifaces, spec)
				rv[idx] = res
				log.Printf("probe %s %v = %s", spec.Type, spec.Args, res.Status)
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
