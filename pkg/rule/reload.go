package rule

import (
	"context"
	"github.com/prometheus/prometheus/model/rulefmt"
	"log"
	"net/http"
)

var _ Writer = &reloadableWriter{}

type reloadableWriter struct {
	delegate Writer

	reloadUrl string
}

// NewReloadable returns a wrapper of the Writer which calls the reloadURL after upserting,
// but in practice it hardly works since it takes several minutes for the change to be propagated
// to the ruler/prometheus container
func NewReloadable(w Writer, reloadUrl string) Writer {
	return &reloadableWriter{
		delegate:  w,
		reloadUrl: reloadUrl,
	}
}

func (w reloadableWriter) UpsertRuleGroup(ctx context.Context, rg *rulefmt.RuleGroup) error {
	if err := w.delegate.UpsertRuleGroup(ctx, rg); err != nil {
		return err
	}
	return w.reload()
}

func (w *reloadableWriter) DeleteRuleGroup(ctx context.Context, name string) error {
	if err := w.delegate.DeleteRuleGroup(ctx, name); err != nil {
		return err
	}
	return w.reload()
}

func (w reloadableWriter) reload() error {
	log.Println("reload")
	_, err := http.Post(w.reloadUrl, "", nil)
	return err
}
