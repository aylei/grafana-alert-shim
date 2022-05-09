package rule

import (
	"context"
	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ Writer = &configmapWriter{}

type configmapWriter struct {
	namespace string
	name      string
	key       string

	kubeCli kubernetes.Interface
}

type ConfigMapOpts struct {
	Namespace string
	Name      string
	Key       string
}

func NewConfigMapWriter(kubeCli kubernetes.Interface, opts *ConfigMapOpts) Writer {
	return &configmapWriter{
		namespace: opts.Namespace,
		name:      opts.Name,
		key:       opts.Key,
		kubeCli:   kubeCli,
	}
}

func (w *configmapWriter) UpsertRuleGroup(ctx context.Context, rg *rulefmt.RuleGroup) error {

	return w.mutateRuleGroups(ctx, func(rgs *rulefmt.RuleGroups) error {
		var found bool
		for i := range rgs.Groups {
			if rgs.Groups[i].Name == rg.Name {
				// FIXME(aylei): possibly overwrite others' change
				rgs.Groups[i] = *rg
				found = true
				break
			}
		}
		if !found {
			rgs.Groups = append(rgs.Groups, *rg)
		}
		return nil
	})
}

func (w *configmapWriter) DeleteRuleGroup(ctx context.Context, name string) error {
	return w.mutateRuleGroups(ctx, func(rgs *rulefmt.RuleGroups) error {
		var found bool
		for i := range rgs.Groups {
			if rgs.Groups[i].Name == name {
				rgs.Groups = append(rgs.Groups[:i], rgs.Groups[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("rule group %s not found", name)
		}
		return nil
	})
}

func (w *configmapWriter) mutateRuleGroups(ctx context.Context, mutateFn func(groups *rulefmt.RuleGroups) error) error {
	cm, err := w.kubeCli.CoreV1().ConfigMaps(w.namespace).Get(ctx, w.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	ruleYaml, ok := cm.Data[w.key]
	if !ok {
		return errors.Errorf("config %s not found in configmap %s/%s", w.key, w.namespace, w.name)
	}

	rgs := rulefmt.RuleGroups{}
	err = yaml.Unmarshal([]byte(ruleYaml), &rgs)
	if err != nil {
		return errors.Errorf("cannot decode rule groups, %s", err.Error())
	}

	err = mutateFn(&rgs)
	if err != nil {
		return err
	}

	newYaml, err := yaml.Marshal(rgs)
	if err != nil {
		return err
	}
	cm.Data[w.key] = string(newYaml)

	_, err = w.kubeCli.CoreV1().ConfigMaps(w.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	// avoiding retry reduce the possibility of overwriting others' change,
	// but client-side retry is outside our control and is still risky
	if err != nil {
		return errors.Errorf("update configmap failed: %s", err.Error())
	}
	return nil
}
