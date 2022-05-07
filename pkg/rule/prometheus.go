package rule

import (
	"context"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
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

func (r *promReader) ListRules(ctx context.Context) ([]v1.RuleGroup, error) {
	res, err := r.promAPI.Rules(ctx)
	if err != nil {
		return nil, err
	}
	return res.Groups, nil
}

//func translate(rgs []v1.RuleGroup) ([]rulefmt.RuleGroup, error) {
//	s, err := yaml.Marshal(rgs)
//	if err != nil {
//		return nil, err
//	}
//
//	var results []rulefmt.RuleGroup
//	err = yaml.Unmarshal(s, &results)
//	if err != nil {
//		return nil, err
//	}
//
//	return results, nil
//}
