package storage

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"

	"mxmz.it/mxmz/tommaso/dto"
)

type objectFactory func() interface{}
type objectCallback func(o interface{})

type FileDB struct {
	basePath string
}

func (db *FileDB) Save(typ string, id string, object interface{}) error {
	if strings.Contains(typ, ".") || strings.Contains(id, ".") {
		panic("invalid identifiers")
	}
	var dir = path.Join(db.basePath, typ)
	os.MkdirAll(dir, 0750)
	var file = path.Join(dir, id)
	if object == nil {
		return os.Remove(file)
	}
	json, _ := json.MarshalIndent(object, "", " ")
	return ioutil.WriteFile(file, json, 0644)
}

func (db *FileDB) VisitOne(typ string, id string, f objectFactory, cb objectCallback) error {
	if strings.Contains(typ, ".") || strings.Contains(id, ".") {
		panic("invalid identifiers")
	}
	var dir = path.Join(db.basePath, typ)
	os.MkdirAll(dir, 0750)
	var file = path.Join(dir, id)
	var bytes, err = ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	object := f()
	err = json.Unmarshal(bytes, object)
	if err != nil {
		return err
	}
	cb(object)
	return nil
}

func (db *FileDB) VisitAll(typ string, fact objectFactory, cb objectCallback) error {
	if strings.Contains(typ, ".") {
		panic("invalid identifiers")
	}
	var dir = path.Join(db.basePath, typ)
	os.MkdirAll(dir, 0750)
	ff, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range ff {
		err := db.VisitOne(typ, f.Name(), fact, cb)
		if err != nil {
			return err
		}
	}
	return nil
}

func StoredProbeSpecFactory() interface{}     { return new(dto.StoredProbeSpec) }
func StoredProbeSpecRuleFactory() interface{} { return new(dto.StoredProbeSpecRule) }

func DefaultFileDB() *FileDB {
	return &FileDB{
		basePath: "./.temp.db",
	}
}

type SimpleProbSpecStore struct {
	DB *FileDB
}

func (s *SimpleProbSpecStore) GetProbeSpecsForNames(ctx context.Context, names *dto.MySources) ([]*dto.ProbeSpec, error) {
	rules, err := s.GetStoredProbeSpecRules(ctx)
	if err != nil {
		return nil, err
	}
	var specNames = map[string]struct{}{}
	for _, v := range rules {
		var regex, err = regexp.Compile(v.Rule.Pattern)
		if err != nil {
			return nil, err
		}
		if v.Rule.Disabled {
			continue
		}

		for _, s := range names.Sources {
			if regex.MatchString(s) {
				for _, n := range v.Rule.SpecNames {
					specNames[n] = struct{}{}
				}
			}
		}
	}
	var specs = []*dto.ProbeSpec{}
	err = s.DB.VisitAll("specs", StoredProbeSpecFactory, func(o interface{}) {
		var o1 = o.(*dto.StoredProbeSpec)
		if o1.Spec.Disabled {
			return
		}
		if _, ok := specNames[o1.ID]; ok {
			specs = append(specs, o1.Spec)
		}
	})

	return specs, nil
}

func (s *SimpleProbSpecStore) GetStoredProbeSpecs(ctx context.Context) ([]*dto.StoredProbeSpec, error) {
	var rv = []*dto.StoredProbeSpec{}
	var err = s.DB.VisitAll("specs", StoredProbeSpecFactory, func(o interface{}) {
		rv = append(rv, o.(*dto.StoredProbeSpec))
	})
	if err != nil {
		return nil, err
	}
	return rv, nil
}

func (s *SimpleProbSpecStore) PutStoredProbeSpec(ctx context.Context, id string, data *dto.ProbeSpec) error {
	if data == nil {
		return s.DB.Save("specs", id, nil)
	}
	return s.DB.Save("specs", id, &dto.StoredProbeSpec{ID: id, Spec: data})
}
func (s *SimpleProbSpecStore) GetStoredProbeSpecRules(ctx context.Context) ([]*dto.StoredProbeSpecRule, error) {
	var rv = []*dto.StoredProbeSpecRule{}
	var err = s.DB.VisitAll("rules", StoredProbeSpecRuleFactory, func(o interface{}) {
		rv = append(rv, o.(*dto.StoredProbeSpecRule))
	})
	if err != nil {
		return nil, err
	}
	return rv, nil

}
func (s *SimpleProbSpecStore) PutStoredProbeSpecRule(ctx context.Context, id string, data *dto.ProbeSpecRule) error {
	if data == nil {
		return s.DB.Save("rules", id, nil)
	}
	return s.DB.Save("rules", id, &dto.StoredProbeSpecRule{ID: id, Rule: data})
}

var volatileResultsLock sync.RWMutex
var volatileResults = map[string][]*dto.StoredProbeResult{}

type VolatileProbResultStore struct {
}

func (s *VolatileProbResultStore) PutResultsForSources(ctx context.Context, results []*dto.ProbeResult) error {
	var newResults = map[string][]*dto.StoredProbeResult{}
	for _, r := range results {
		for _, s := range r.Sources {
			stored := dto.StoredProbeResult{
				Source:  s,
				Type:    r.Spec.Type,
				Args:    r.Spec.Args,
				Time:    r.Time,
				Status:  r.Status,
				Elapsed: r.Elapsed,
				Comment: r.Comment,
			}
			newResults[s] = append(newResults[s], &stored)
		}
	}

	volatileResultsLock.Lock()
	for k, v := range newResults {
		volatileResults[k] = v
	}
	defer volatileResultsLock.Unlock()
	return nil
}
func (s *VolatileProbResultStore) GetResultsBySourcePrefix(ctx context.Context, sourcePrefix string) ([]*dto.StoredProbeResult, error) {
	var rv = []*dto.StoredProbeResult{}
	volatileResultsLock.RLock()
	for k, v := range volatileResults {
		if strings.HasPrefix(k, sourcePrefix) {
			rv = append(rv, v...)
		}
	}
	defer volatileResultsLock.RUnlock()
	return rv, nil
}

func (s *VolatileProbResultStore) GetResultsWithSubstring(ctx context.Context, substr string) ([]*dto.StoredProbeResult, error) {
	if substr == "" {
		return s.GetResultsBySourcePrefix(ctx, "")
	}
	var rv = []*dto.StoredProbeResult{}
	volatileResultsLock.RLock()
	for k, v := range volatileResults {
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
	defer volatileResultsLock.RUnlock()
	return rv, nil
}
