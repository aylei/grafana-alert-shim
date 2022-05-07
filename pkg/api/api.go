package api

import (
	"github.com/aylei/alert-shim/pkg/rule"
	"github.com/gin-gonic/gin"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"net/http"
)

const (
	// deal with unified alerting probe
	dummyNamespace = "test"
	dummyGroup     = "test"

	defaultNamespace = "default"
)

func New(ruleCli rule.Client) *gin.Engine {
	router := gin.New()
	s := &server{ruleCli: ruleCli}
	v1 := router.Group("/api/v1")
	{
		v1.GET("/rules", s.ListRules)
		v1.GET("/rules/:namespace", s.ListRules)
		v1.GET("/rules/:namespace/:group", s.ListRules)
		v1.POST("/rules/:namespace", s.CreateRuleGroup)
	}

	return router
}

type server struct {
	ruleCli rule.Client
}

func (s *server) ListRules(c *gin.Context) {
	ns := c.Param("namespace")
	group := c.Param("group")

	if ns == dummyNamespace || group == dummyGroup {
		WriteDummyRuleGroup(c)
		return
	}

	rgs, err := s.ruleCli.ListRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if group != "" {
		// filter by group
		var targets []v1.RuleGroup
		for _, rg := range rgs {
			if rg.Name == group {
				targets = append(targets, rg)
			}
		}
		rgs = targets
	}

	c.YAML(http.StatusOK, gin.H{defaultNamespace: rgs})
}

func (s *server) CreateRuleGroup(c *gin.Context) {
	// TODO(aylei)
	c.YAML(http.StatusOK, nil)
}

func WriteDummyRuleGroup(c *gin.Context) {
	c.YAML(http.StatusOK, nil)
}
