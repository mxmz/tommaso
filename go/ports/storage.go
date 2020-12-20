package ports

import (
	"context"
	"time"

	"mxmz.it/mxmz/tommaso/dto"
)

type ProbeSpecStore interface {
	GetProbeTestingSpecsForNames(ctx context.Context, names *dto.MySources) ([]*dto.ProbeTestingSpec, error)

	GetStoredProbeSpecs(ctx context.Context) ([]*dto.StoredProbeSpec, error)
	PutStoredProbeSpec(ctx context.Context, id string, data *dto.ProbeSpec) error

	GetStoredProbeSpecRules(ctx context.Context) ([]*dto.StoredProbeSpecRule, error)
	PutStoredProbeSpecRule(ctx context.Context, id string, data *dto.ProbeSpecRule) error
	ClearAll(ctx context.Context) error
}

type ProbeResultStore interface {
	PutResultsForSources(ctx context.Context, results []*dto.ProbeResult) error
	LastUpdateAt(ctx context.Context) time.Time
	GetResultsBySourcePrefix(ctx context.Context, sourcePrefix string) ([]*dto.StoredProbeResult, error)
	GetResultsWithSubstring(ctx context.Context, substr string) ([]*dto.StoredProbeResult, error)
	ClearAll(ctx context.Context) error
}
