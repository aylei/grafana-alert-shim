package api

import (
	"github.com/aylei/alert-shim/pkg/config"
	"github.com/aylei/alert-shim/pkg/rule"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func compose(conf *config.Config) (*server, error) {
	// TODO(aylei): plugable
	var err error
	var r rule.Reader
	switch conf.Reader.Type {
	case config.ReaderTypeNoop:
		r = &rule.NoopReader{}
	case config.ReaderTypeGeneric:
		r, err = rule.NewPromReader(conf.Reader.Generic.RulerBaseURL)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unknown reader type %s", conf.Reader.Type)
	}
	var w rule.Writer
	switch conf.Writer.Type {
	case config.WriterTypeNoop:
		w = &rule.NoopWriter{}
	case config.WriterTypeConfigMap:
		w, err = buildConfigMapWriter(conf.Writer.ConfigMap)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.Errorf("unknown writer type %s", conf.Writer.Type)
	}

	if conf.Writer.ReloadURL != "" {
		w = rule.NewReloadable(w, conf.Writer.ReloadURL)
	}

	ruleCli := rule.NewClient(r, w)
	return &server{
		ruleCli: ruleCli,
		conf:    conf,
	}, nil
}

func buildConfigMapWriter(conf *config.ConfigMapWriterConfig) (rule.Writer, error) {
	// TODO(aylei): from command-line flags?
	kubeconfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return nil, err
	}
	kubeCli, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	writer := rule.NewConfigMapWriter(kubeCli, &rule.ConfigMapOpts{
		Namespace: conf.Namespace,
		Name:      conf.Name,
		Key:       conf.Key,
	})
	return writer, nil
}
