package storage

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"mxmz.it/mxmz/tommaso/dto"
	"mxmz.it/mxmz/tommaso/ports"
)

var volatileResultsLock sync.RWMutex
var volatileResults = map[string][]*dto.StoredProbeResult{}
var volatileUpdatedAt time.Time

type VolatileProbResultStore struct {
}

var _ ports.ProbeResultStore = (*VolatileProbResultStore)(nil)

func (s *VolatileProbResultStore) LastUpdateAt(ctx context.Context) time.Time {
	volatileResultsLock.RLock()
	defer volatileResultsLock.RUnlock()
	return volatileUpdatedAt
}
func (s *VolatileProbResultStore) PutResultsForSources(ctx context.Context, results []*dto.ProbeResult) error {
	var newResults = map[string][]*dto.StoredProbeResult{}
	for _, r := range results {
		s := strings.Join(r.Sources, " ")
		stored := dto.StoredProbeResult{
			Source:        s,
			Type:          r.Spec.Type,
			Args:          r.Spec.Args,
			Time:          r.Time,
			Status:        r.Status,
			Elapsed:       r.Elapsed,
			Comment:       r.Comment,
			Description:   r.Spec.Description,
			ExpectFailure: r.Spec.ExpectFailure,
			Pass:          (r.Status == "OK" && !r.Spec.ExpectFailure) || (r.Status != "OK" && r.Spec.ExpectFailure),
		}
		newResults[s] = append(newResults[s], &stored)
	}

	volatileResultsLock.Lock()
	for k, v := range newResults {
		volatileResults[k] = v
	}
	volatileUpdatedAt = time.Now()
	defer volatileResultsLock.Unlock()
	return nil
}

var recordTTL = 8 * time.Hour

func areAllOlder(l []*dto.StoredProbeResult, refTime time.Time) bool {
	for _, v := range l {
		if refTime.Sub(v.Time) < recordTTL {
			return false
		}
	}
	return true
}

func (s *VolatileProbResultStore) purge(ks []string) {
	if len(ks) > 0 {
		volatileResultsLock.Lock()
		defer volatileResultsLock.Unlock()
		for _, s := range ks {
			delete(volatileResults, s)
			fmt.Printf("purged %", s)
		}
	}
}

func (s *VolatileProbResultStore) GetResultsBySourcePrefix(ctx context.Context, sourcePrefix string) ([]*dto.StoredProbeResult, error) {
	var rv = []*dto.StoredProbeResult{}
	var toBeRemoved = []string{}
	defer s.purge(toBeRemoved)
	var now = time.Now()
	volatileResultsLock.RLock()
	defer volatileResultsLock.RUnlock()
	for k, v := range volatileResults {
		if areAllOlder(v, now) {
			toBeRemoved = append(toBeRemoved, k)
		} else if strings.HasPrefix(k, sourcePrefix) {
			rv = append(rv, v...)
		}
	}

	return rv, nil
}

func (s *VolatileProbResultStore) GetResultsWithSubstring(ctx context.Context, substr string) ([]*dto.StoredProbeResult, error) {
	if substr == "" {
		return s.GetResultsBySourcePrefix(ctx, "")
	}
	var rv = []*dto.StoredProbeResult{}
	var toBeRemoved = []string{}
	defer s.purge(toBeRemoved)
	var now = time.Now()
	volatileResultsLock.RLock()
	defer volatileResultsLock.RUnlock()
	for k, v := range volatileResults {
		if areAllOlder(v, now) {
			toBeRemoved = append(toBeRemoved, k)
		} else {
			if strings.Contains(k, substr) {
				rv = append(rv, v...)
			} else {
				for _, l := range v {
					if len(l.Args) > 0 && strings.Contains(l.Args[0], substr) {
						rv = append(rv, l)
					}
				}
			}
		}
	}
	return rv, nil
}
