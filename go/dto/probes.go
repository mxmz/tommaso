package dto

import "time"

type ProbeResult struct {
	Spec    *ProbeTestingSpec `json:"spec"`
	Status  string            `json:"status"`
	Time    time.Time         `json:"time"`
	Elapsed int               `json:"elapsed"`
	Sources []string          `json:"sources"`
	Comment string            `json:"comment"`
}

func (r *ProbeResult) IsOK() bool {
	return (r.Status == "OK" && !r.Spec.ExpectFailure) || (r.Status != "OK" && r.Spec.ExpectFailure)
}

type ProbeSpec struct {
	Type        string   `json:"type"`
	Args        []string `json:"args"`
	Timeout     int      `json:"timeout"`
	Disabled    bool     `json:"disabled"`
	Description string   `json:"description"`
}

type ProbeTestingSpec struct {
	Type          string   `json:"type"`
	Args          []string `json:"args"`
	Timeout       int      `json:"timeout"`
	Description   string   `json:"description"`
	ExpectFailure bool     `json:"expect_failure"`
}
type ProbeSpecRule struct {
	Pattern   string   `json:"pattern"`
	SpecNames []string `json:"spec_names"`
	Disabled  bool     `json:"disabled"`
}

type StoredProbeResult struct {
	Type          string    `json:"type"`
	Args          []string  `json:"args"`
	Source        string    `json:"source"`
	Status        string    `json:"status"`
	Time          time.Time `json:"time"`
	Elapsed       int       `json:"elapsed"`
	Comment       string    `json:"comment"`
	Description   string    `json:"description"`
	ExpectFailure bool      `json:"expect_failure"`
	Pass          bool      `json:"pass"`
}

type StoredProbeSpecRule struct {
	ID   string         `json:"id"`
	Rule *ProbeSpecRule `json:"rule"`
}

type StoredProbeSpec struct {
	ID   string     `json:"id"`
	Spec *ProbeSpec `json:"spec"`
}
