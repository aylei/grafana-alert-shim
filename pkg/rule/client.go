package rule

import (
	"context"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	"time"
)

type Reader interface {
	// ListRules list all rules group by namespace or rule file
	ListRules(ctx context.Context) (map[string][]rulefmt.RuleGroup, error)
	// ListPromRules list all alert rules in promethues format
	ListPromRules(ctx context.Context) ([]RuleGroup, error)
}

type Writer interface {
	UpsertRuleGroup(ctx context.Context, rg *rulefmt.RuleGroup) error
	DeleteRuleGroup(ctx context.Context, groupName string) error
}

type Client interface {
	Reader
	Writer
}

func NewClient(r Reader, w Writer) Client {
	return &client{
		Reader: r,
		Writer: w,
	}
}

type client struct {
	Reader
	Writer
}

type RuleGroup struct {
	Name     string  `json:"name"`
	File     string  `json:"file"`
	Interval float64 `json:"interval"`
	Rules    []Rule  `json:"rules"`
}

type Rule struct {
	Name           string         `json:"name"`
	Query          string         `json:"query"`
	Duration       float64        `json:"duration"`
	Labels         model.LabelSet `json:"labels"`
	Annotations    model.LabelSet `json:"annotations"`
	Alerts         []*v1.Alert    `json:"alerts"`
	Health         v1.RuleHealth  `json:"health"`
	LastError      string         `json:"lastError,omitempty"`
	EvaluationTime float64        `json:"evaluationTime"`
	LastEvaluation time.Time      `json:"lastEvaluation"`
	State          string         `json:"state"`
	Type           string         `json:"type"`
}

var _ Writer = &NoopWriter{}

type NoopWriter struct{}

func (w *NoopWriter) DeleteRuleGroup(ctx context.Context, groupName string) error {
	return nil
}

func (w *NoopWriter) UpsertRuleGroup(ctx context.Context, rg *rulefmt.RuleGroup) error {
	return nil
}

var _ Reader = &NoopReader{}

type NoopReader struct{}

func (r *NoopReader) ListRules(ctx context.Context) (map[string][]rulefmt.RuleGroup, error) {
	return nil, nil
}

func (r *NoopReader) ListPromRules(ctx context.Context) ([]RuleGroup, error) {
	return nil, nil
}
