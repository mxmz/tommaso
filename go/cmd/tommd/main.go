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

	var bseURL = "http://localhost:7997"
	if len(os.Args) > 1 {
		bseURL = os.Args[1]
	}
	for {

		var specs, err = getMyProbeSpecs(bseURL)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		var p = &prober.Prober{}
		rv := p.RunProbSpecsConcurrent(specs)

		//	var _ = err
		var _ = rv
		fmt.Printf("%v\n", jsonIndent(rv))
		err = pushMyProbeResults(bseURL, rv)
		fmt.Printf("%v\n", err)
		time.Sleep(30 * time.Second)
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
