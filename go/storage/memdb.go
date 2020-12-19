package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	"mxmz.it/mxmz/tommaso/dto"
)

type MemProbeSpecDB struct {
	eventStoreDir string
	counter       int
	probes        map[string]*dto.ProbeSpec
	rules         map[string]*compiledProbeSpecRule
}

type compiledProbeSpecRule struct {
	dto.ProbeSpecRule
	compiledPattern *regexp.Regexp
}

type SetProbeEvent struct {
	ID   string         `json:"id,omitempty"`
	Data *dto.ProbeSpec `json:"data,omitempty"`
}

type SetRuleEvent struct {
	ID   string             `json:"id,omitempty"`
	Data *dto.ProbeSpecRule `json:"data,omitempty"`
}

type storedEvent struct {
	Probe *SetProbeEvent `json:"probe,omitempty"`
	Rule  *SetRuleEvent  `json:"rule,omitempty"`
}

func (d *MemProbeSpecDB) GetProbeTestingSpecsForNames(ctx context.Context, names *dto.MySources) ([]*dto.ProbeTestingSpec, error) {
	var specNames = map[string]struct{}{}
	var specNamesNot = map[string]struct{}{}
	for _, v := range d.rules {
		if v.Disabled {
			continue
		}
		var regex = v.compiledPattern
		for _, s := range names.Sources {
			if regex.MatchString(s) {
				for _, n := range v.SpecNames {
					if len(n) > 0 {
						if n[0] == '!' {
							specNamesNot[n[1:]] = struct{}{}
						} else {
							specNames[n] = struct{}{}
						}
					}
				}
			}
		}
	}
	var specs = []*dto.ProbeTestingSpec{}
	for k, v := range d.probes {
		if v.Disabled {
			continue
		}
		if _, ok := specNames[k]; ok {
			specs = append(specs,
				&dto.ProbeTestingSpec{
					Type:          v.Type,
					Args:          v.Args,
					Timeout:       v.Timeout,
					Description:   v.Description,
					ExpectFailure: false,
				},
			)
		} else if _, ok := specNamesNot[k]; ok {
			specs = append(specs,
				&dto.ProbeTestingSpec{
					Type:          v.Type,
					Args:          v.Args,
					Timeout:       v.Timeout,
					Description:   v.Description + " (!)",
					ExpectFailure: true,
				},
			)
		}
	}

	return specs, nil
}

func (d *MemProbeSpecDB) GetStoredProbeSpecs(ctx context.Context) ([]*dto.StoredProbeSpec, error) {
	var r = make([]*dto.StoredProbeSpec, 0, len(d.probes))
	for k, v := range d.probes {
		r = append(r, &dto.StoredProbeSpec{ID: k, Spec: v})
	}
	return r, nil
}

func (d *MemProbeSpecDB) PutStoredProbeSpec(ctx context.Context, id string, data *dto.ProbeSpec) error {
	var e = &SetProbeEvent{ID: id, Data: data}
	return firstError(d.applySetProbeEvent(e), d.storeEvent(e, nil))
}

func (d *MemProbeSpecDB) GetStoredProbeSpecRules(ctx context.Context) ([]*dto.StoredProbeSpecRule, error) {
	var r = make([]*dto.StoredProbeSpecRule, 0, len(d.probes))
	for k, v := range d.rules {
		r = append(r, &dto.StoredProbeSpecRule{ID: k, Rule: &v.ProbeSpecRule})
	}
	return r, nil
}

func (d *MemProbeSpecDB) PutStoredProbeSpecRule(ctx context.Context, id string, data *dto.ProbeSpecRule) error {
	var e = &SetRuleEvent{ID: id, Data: data}
	return firstError(d.applySetRuleEvent(e), d.storeEvent(nil, e))
}

func (d *MemProbeSpecDB) applySetProbeEvent(e *SetProbeEvent) error {
	log.Println(e.ID, e.Data)
	if e.Data == nil {
		delete(d.probes, e.ID)
	} else {
		d.probes[e.ID] = e.Data
	}
	return nil
}

func (d *MemProbeSpecDB) applySetRuleEvent(e *SetRuleEvent) error {
	log.Println(e.ID, e.Data)
	if e.Data == nil {
		delete(d.rules, e.ID)
	} else {
		var regex, err = regexp.Compile(e.Data.Pattern)
		if err != nil {
			return err
		}
		d.rules[e.ID] = &compiledProbeSpecRule{*e.Data, regex}
	}
	return nil
}

func (d *MemProbeSpecDB) init() {

	var files, err = ioutil.ReadDir(d.eventStoreDir)
	if err != nil {
		panic(err)
	}
	var fileNames = make([]string, 0, len(files))
	for _, v := range files {
		if v.Mode().IsRegular() {
			fileNames = append(fileNames, v.Name())
		}
	}
	sort.Strings(fileNames)
	for _, v := range fileNames {
		var file = fmt.Sprintf("%s/%s", d.eventStoreDir, v)
		var bytes, err = ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		var e storedEvent
		err = json.Unmarshal(bytes, &e)
		if err != nil {
			panic(err)
		}
		if e.Probe != nil {
			d.applySetProbeEvent(e.Probe)
		}
		if e.Rule != nil {
			d.applySetRuleEvent(e.Rule)
		}
	}

}
func (d *MemProbeSpecDB) storeEvent(e1 *SetProbeEvent, e2 *SetRuleEvent) error {
	d.counter++
	var stored = storedEvent{e1, e2}
	var now = time.Now()
	var file = fmt.Sprintf("%s/%021d-%015d", d.eventStoreDir, now.UnixNano(), d.counter)
	json, _ := json.MarshalIndent(stored, "", " ")
	return ioutil.WriteFile(file, json, 0644)
}

func firstError(e ...error) error {
	for _, v := range e {
		if v != nil {
			return v
		}
	}
	return nil
}

type syncMemDB struct {
	db   *MemProbeSpecDB
	lock sync.RWMutex
}

func (d *syncMemDB) GetProbeTestingSpecsForNames(ctx context.Context, names *dto.MySources) ([]*dto.ProbeTestingSpec, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.db.GetProbeTestingSpecsForNames(ctx, names)
}

func (d *syncMemDB) GetStoredProbeSpecs(ctx context.Context) ([]*dto.StoredProbeSpec, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.db.GetStoredProbeSpecs(ctx)
}

func (d *syncMemDB) PutStoredProbeSpec(ctx context.Context, id string, data *dto.ProbeSpec) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.db.PutStoredProbeSpec(ctx, id, data)
}

func (d *syncMemDB) GetStoredProbeSpecRules(ctx context.Context) ([]*dto.StoredProbeSpecRule, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.db.GetStoredProbeSpecRules(ctx)
}

func (d *syncMemDB) PutStoredProbeSpecRule(ctx context.Context, id string, data *dto.ProbeSpecRule) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.db.PutStoredProbeSpecRule(ctx, id, data)
}

func NewMemProbeSpecDB(path string) *MemProbeSpecDB {
	os.MkdirAll(path, 0750)
	var db = &MemProbeSpecDB{
		eventStoreDir: path,
		probes:        map[string]*dto.ProbeSpec{},
		rules:         map[string]*compiledProbeSpecRule{},
	}

	db.init()

	return db
}

func NewSyncMemProbeSpecDB(db *MemProbeSpecDB) *syncMemDB {
	return &syncMemDB{db: db}
}
