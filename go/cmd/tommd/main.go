package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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
	var p = prober.NewProber()
	var probed = 0
	for {
		var specs, err = getMyProbeSpecs(baseURL)
		fmt.Printf("getMyProbeSpecs: %s: err = %v\n", baseURL, err)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		rv := p.RunProbSpecsConcurrent(specs)
		//	var _ = err
		var _ = rv
		fmt.Printf("probe results = %v\n", len(rv))
		if p.Probed != probed {
			err = pushMyProbeResults(baseURL, rv)
			fmt.Printf("pushMyProbeResults: err = %v\n", err)
		}
		probed = p.Probed
		fmt.Println("sleep ...")
		time.Sleep(5 * time.Second)
	}

}

func jsonIndent(v interface{}) string {
	b, _ := json.MarshalIndent(v, " ", " ")
	return string(b)
}

func getMyProbeSpecs(baseURL string) ([]*dto.ProbeSpec, error) {

	var ifaces = system.GetNetInterfaceAddresses()
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
	var probeSpecs []*dto.ProbeSpec
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
