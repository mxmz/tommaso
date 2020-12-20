package storage

/*
import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"mxmz.it/mxmz/tommaso/dto"
	"mxmz.it/mxmz/tommaso/ports"
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

var _ ports.ProbeSpecStore = (*SimpleProbSpecStore)(nil)

type SimpleProbSpecStore struct {
	DB *FileDB
}

func (s *SimpleProbSpecStore) GetProbeTestingSpecsForNames(ctx context.Context, names *dto.MySources) ([]*dto.ProbeTestingSpec, error) {
	rules, err := s.GetStoredProbeSpecRules(ctx)
	if err != nil {
		return nil, err
	}
	var specNames = map[string]struct{}{}
	var specNamesNot = map[string]struct{}{}
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
	err = s.DB.VisitAll("specs", StoredProbeSpecFactory, func(o interface{}) {
		var o1 = o.(*dto.StoredProbeSpec)
		if o1.Spec.Disabled {
			return
		}
		if _, ok := specNames[o1.ID]; ok {
			specs = append(specs,
				&dto.ProbeTestingSpec{
					Type:          o1.Spec.Type,
					Args:          o1.Spec.Args,
					Timeout:       o1.Spec.Timeout,
					Description:   o1.Spec.Description,
					ExpectFailure: false,
				},
			)
		} else if _, ok := specNamesNot[o1.ID]; ok {
			specs = append(specs,
				&dto.ProbeTestingSpec{
					Type:          o1.Spec.Type,
					Args:          o1.Spec.Args,
					Timeout:       o1.Spec.Timeout,
					Description:   o1.Spec.Description + " (!)",
					ExpectFailure: true,
				},
			)
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
*/
