package rule

import (
	"context"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/prometheus/model/rulefmt"
	"go.uber.org/zap"
)

type Reader interface {
	ListRules(ctx context.Context) ([]v1.RuleGroup, error)
}

type Writer interface {
	CreateRuleGroup(ctx context.Context, rg *rulefmt.RuleGroup) error
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

type NoopWriter struct{}

func (w *NoopWriter) CreateRuleGroup(ctx context.Context, rg *rulefmt.RuleGroup) error {
	zap.L().Info("CreateRuleGroup", zap.Any("body", rg))
	return nil
}
