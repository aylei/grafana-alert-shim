package api

import (
	"fmt"
	"github.com/aylei/alert-shim/pkg/config"
	"github.com/aylei/alert-shim/pkg/rule"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/prometheus/model/rulefmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

//var (
//	ruleGroupRegex = regexp.MustCompile("/rules/.*/(.*)")
//)

func New(conf *config.Config) (*gin.Engine, error) {
	g := gin.Default()
	s, err := compose(conf)
	if err != nil {
		return nil, err
	}

	// avoid escaping url-encoded path param since the file name may contain '/'
	g.UseRawPath = true
	g.UnescapePathValues = false

	g.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})
	cortex := g.Group("")
	{
		cortex.GET("/rules", s.ListRules)
		cortex.GET("/rules/:namespace", s.ListRules)
		cortex.GET("/rules/:namespace/:group", s.GetRuleGroup)
		cortex.POST("/rules/:namespace", s.PostRuleGroup)
	}
	// proxy to prometheus
	prom := g.Group("/api/v1")
	{
		prom.Any("/*any", s.RouteProm)
	}

	return g, nil
}

func compose(conf *config.Config) (*server, error) {
	// TODO(aylei): plugable
	reader, err := rule.NewPromReader(conf.Reader.Generic.RulerBaseURL)
	if err != nil {
		return nil, err
	}
	writer := &rule.NoopWriter{}
	ruleCli := rule.NewClient(reader, writer)
	return &server{
		ruleCli: ruleCli,
		conf:    conf,
	}, nil
}

type server struct {
	ruleCli rule.Client

	conf *config.Config
}

func (s *server) RouteProm(c *gin.Context) {
	path := c.Param("any")
	switch {
	case path == "/status/buildinfo":
		// cortex does not serve /buildinfo and grafana relies on this
		c.JSON(http.StatusNotFound, nil)
	case path == "/rules":
		// route rules query to the ruler endpoint
		s.promeRules(c)
	//case ruleGroupRegex.MatchString(path):
	//	// grafana queries rule group under /api/v1/rules when editing an existing rule
	//	s.getRuleGroup(c, ruleGroupRegex.FindStringSubmatch(path)[1])
	default:
		// anything else goes to the query endpoint
		s.queryProxy(c)
	}
}

func (s *server) queryProxy(c *gin.Context) {
	remote, err := url.Parse(s.conf.Reader.Generic.QuerierBaseURL)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (s *server) promeRules(c *gin.Context) {
	rules, err := s.ruleCli.ListPromRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"groups": rules,
		},
	})
}

// ListRules list all alert rules in cortex format
func (s *server) ListRules(c *gin.Context) {
	rgMap, err := s.ruleCli.ListRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Render(http.StatusOK, YAMLv3{Data: rgMap})
}

// GetRuleGroup get and alert group in cortex format
func (s *server) GetRuleGroup(c *gin.Context) {
	group := c.Param("group")
	s.getRuleGroup(c, group)
}

func (s *server) getRuleGroup(c *gin.Context, group string) {
	rgMap, err := s.ruleCli.ListRules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// filter by group
	var res rulefmt.RuleGroup
	var found bool
	for _, rgs := range rgMap {
		for _, rg := range rgs {
			if rg.Name == group {
				res = rg
				found = true
				break
			}
		}
	}

	if found {
		c.Render(http.StatusOK, YAMLv3{Data: res})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("group does not exist\n")})
	}
}

// PostRuleGroup overrides the rule group
func (s *server) PostRuleGroup(c *gin.Context) {
	// TODO(aylei)
	c.YAML(http.StatusOK, nil)
}
