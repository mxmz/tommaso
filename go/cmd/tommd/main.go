package main

import (
	"encoding/json"

	"mxmz.it/mxmz/tommaso/dto"
	"mxmz.it/mxmz/tommaso/prober"
)

func main() {

	var specs = []*dto.ProbeSpec{
		{
			Type:    "tcp",
			Args:    []string{"172.27.203.2", "80"},
			Timeout: 5,
		},
		{
			Type:    "tcp",
			Args:    []string{"172.27.203.2", "80"},
			Timeout: 5000,
		},
		{
			Type:    "tcp",
			Args:    []string{"172.27.203.2", "80"},
			Timeout: 5000,
		},
		{
			Type:    "tcp",
			Args:    []string{"172.27.203.2", "80"},
			Timeout: 5000,
		},
		{
			Type:    "tcp",
			Args:    []string{"172.27.203.2", "80"},
			Timeout: 5000,
		},
		{
			Type:    "tcp",
			Args:    []string{"172.27.203.2", "80"},
			Timeout: 8,
		},
	}

	var p = &prober.Prober{}
	rv := p.RunProbSpecsConcurrent(specs)

	//	var _ = err
	var _ = rv
	//fmt.Printf("%v\n", jsonIndent(rv))
}

func jsonIndent(v interface{}) string {
	b, _ := json.MarshalIndent(v, " ", " ")
	return string(b)
}
