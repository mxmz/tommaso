package dto

import "time"

type ProbeResult struct {
	Spec    *ProbeSpec
	Status  string
	Time    time.Time
	Elapsed int
	Sources []string
	Comment string
}

type ProbeSpec struct {
	Type    string
	Args    []string
	Timeout int
}
