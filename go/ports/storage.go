package ports

import (
	"context"

	"mxmz.it/mxmz/tommaso/dto"
)

type ProbeSpecStore interface {
	GetProbeSpecsForNames(ctx context.Context, names *dto.MySources) ([]*dto.ProbeSpec, error)

	GetStoredProbeSpecs(ctx context.Context) ([]*dto.StoredProbeSpec, error)
	PutStoredProbeSpec(ctx context.Context, id string, data *dto.ProbeSpec) error

	GetStoredProbeSpecRules(ctx context.Context) ([]*dto.StoredProbeSpecRule, error)
	PutStoredProbeSpecRule(ctx context.Context, id string, data *dto.ProbeSpecRule) error
}

type ProbeResultStore interface {
	PutResultsForSources(ctx context.Context, results []*dto.ProbeResult) error
	GetResultsBySourcePrefix(ctx context.Context, sourcePrefix string) ([]*dto.StoredProbeResult, error)
}
