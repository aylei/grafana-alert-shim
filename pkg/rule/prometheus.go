package rule

import (
	"context"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
	"unsafe"
)

var _ Reader = &promReader{}

type promReader struct {
	promAPI v1.API
}

func NewPromReader(baseURL string) (Reader, error) {
	cli, err := api.NewClient(api.Config{Address: baseURL})
	if err != nil {
		return nil, err
	}
	return &promReader{promAPI: v1.NewAPI(cli)}, nil
}

func (r *promReader) ListRules(ctx context.Context) (map[string][]rulefmt.RuleGroup, error) {
	res, err := r.promAPI.Rules(ctx)
	if err != nil {
		return nil, err
	}
	rgMap := map[string][]rulefmt.RuleGroup{}
	for i := range res.Groups {
		rg, err := fromPromV1(&res.Groups[i])
		if err != nil {
			return nil, err
		}
		// FIXME(aylei): might not work on win container
		splits := strings.Split(res.Groups[i].File, "/")
		groupKey := splits[len(splits)-1]
		rgMap[groupKey] = append(rgMap[groupKey], *rg)
	}
	return rgMap, nil
}

func (r *promReader) ListPromRules(ctx context.Context) ([]RuleGroup, error) {
	res, err := r.promAPI.Rules(ctx)
	if err != nil {
		return nil, err
	}
	rgs := make([]RuleGroup, len(res.Groups))
	for i := range res.Groups {
		rg, err := translateWithType(&res.Groups[i])
		if err != nil {
			return nil, err
		}
		rgs[i] = *rg
	}
	return rgs, nil
}

func translateWithType(rg *v1.RuleGroup) (*RuleGroup, error) {
	fullPath := rg.File
	splits := strings.Split(fullPath, "/")
	newRg := RuleGroup{
		Name:     rg.Name,
		File:     splits[len(splits)-1],
		Interval: rg.Interval,
		Rules:    make([]Rule, len(rg.Rules)),
	}
	for i, r := range rg.Rules {
		var rule Rule
		switch v := r.(type) {
		case v1.RecordingRule:
			rule = Rule{
				Name:           v.Name,
				Query:          v.Query,
				Labels:         v.Labels,
				Health:         v.Health,
				LastError:      v.LastError,
				EvaluationTime: v.EvaluationTime,
				LastEvaluation: v.LastEvaluation,
				Type:           string(v1.RuleTypeRecording),
			}
		case v1.AlertingRule:
			rule = Rule{
				Name:           v.Name,
				Query:          v.Query,
				Duration:       v.Duration,
				Labels:         v.Labels,
				Annotations:    v.Annotations,
				Alerts:         v.Alerts,
				Health:         v.Health,
				LastError:      v.LastError,
				EvaluationTime: v.EvaluationTime,
				LastEvaluation: v.LastEvaluation,
				State:          v.State,
				Type:           string(v1.RuleTypeAlerting),
			}
		default:
			return nil, errors.Errorf("unknown rule type %s", v)
		}
		newRg.Rules[i] = rule
	}
	return &newRg, nil
}

func fromPromV1(rg *v1.RuleGroup) (*rulefmt.RuleGroup, error) {
	formatted := rulefmt.RuleGroup{
		Name:     rg.Name,
		Interval: model.Duration(time.Duration(rg.Interval) * time.Second),
		Rules:    make([]rulefmt.RuleNode, len(rg.Rules)),
	}

	for i, r := range rg.Rules {
		var rule rulefmt.RuleNode
		switch v := r.(type) {
		case v1.RecordingRule:
			rule = rulefmt.RuleNode{
				Record: stringNode(v.Name),
				Expr:   stringNode(v.Query),
				Labels: fromLabelSet(v.Labels),
			}
		case v1.AlertingRule:
			rule = rulefmt.RuleNode{
				Alert:       stringNode(v.Name),
				Expr:        stringNode(v.Query),
				For:         model.Duration(time.Duration(v.Duration) * time.Second),
				Labels:      fromLabelSet(v.Labels),
				Annotations: fromLabelSet(v.Annotations),
			}
		default:
			return nil, errors.Errorf("unknown rule type %s", v)
		}
		formatted.Rules[i] = rule
	}
	return &formatted, nil
}

func stringNode(s string) yaml.Node {
	node := yaml.Node{}
	node.SetString(s)
	return node
}

func fromLabelSet(ls model.LabelSet) map[string]string {
	return *(*map[string]string)(unsafe.Pointer(&ls))
}
