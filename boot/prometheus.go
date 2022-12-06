package boot

import (
	"fmt"
	"github.com/dream-mo/prom-elastic-alert/conf"
	"github.com/dream-mo/prom-elastic-alert/utils/logger"
	redisx "github.com/dream-mo/prom-elastic-alert/utils/redis"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
	"sync"
)

type ElasticAlertPrometheusMetrics struct {
	Query         sync.Map // map[string]QueryMetrics
	OpRedis       sync.Map // map[string]OpRedisMetrics
	WebhookNotify sync.Map // map[string]WebhookNotifyMetrics
}

func NewElasticAlertPrometheusMetrics() *ElasticAlertPrometheusMetrics {
	return &ElasticAlertPrometheusMetrics{
		Query:         sync.Map{},
		OpRedis:       sync.Map{},
		WebhookNotify: sync.Map{},
	}
}

type OpRedisMetrics struct {
	UniqueId string
	Path     string
	Cmd      string
	Key      string
	Status   int
	Value    int64
}

type QueryMetrics struct {
	UniqueId  string
	Path      string
	EsAddress string
	Index     string
	Status    int
	Value     int64
}

type WebhookNotifyMetrics struct {
	UniqueId string
	Path     string
	Status   int
	Value    int64
}

type RuleStatusCollector struct {
	Ea                *ElasticAlert
	AppInfoDesc       *prometheus.Desc
	RuleStatusDesc    *prometheus.Desc
	LinkRedisDesc     *prometheus.Desc
	QueryDesc         *prometheus.Desc
	OpRedisDesc       *prometheus.Desc
	WebhookNotifyDesc *prometheus.Desc
}

func (rc *RuleStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- rc.RuleStatusDesc
	ch <- rc.AppInfoDesc
	ch <- rc.LinkRedisDesc
	ch <- rc.QueryDesc
	ch <- rc.OpRedisDesc
	ch <- rc.WebhookNotifyDesc
}

func (rc *RuleStatusCollector) Collect(ch chan<- prometheus.Metric) {
	rc.collectAppInfo(ch)
	rc.collectLinkRedisStatus(ch)
	rc.Ea.rules.Range(func(key, value any) bool {
		rule := value.(*conf.Rule)
		rc.collectRuleStatus(ch, rule)
		rc.collectQueryMetrics(ch, rule)
		rc.collectOpRedisMetrics(ch, rule)
		rc.collectWebhookNotifyMetrics(ch, rule)
		return true
	})
}

func (rc *RuleStatusCollector) collectQueryMetrics(ch chan<- prometheus.Metric, rule *conf.Rule) {
	val, ok := rc.Ea.metrics.Load(rule.UniqueId)
	if ok {
		m := val.(*ElasticAlertPrometheusMetrics)
		m.Query.Range(func(key, value any) bool {
			v := value.(QueryMetrics)
			labelValues := []string{v.UniqueId, v.Path, v.EsAddress, v.Index, strconv.Itoa(v.Status)}
			ch <- prometheus.MustNewConstMetric(rc.QueryDesc, prometheus.CounterValue, float64(v.Value), labelValues...)
			return true
		})
	}
}

func (rc *RuleStatusCollector) collectOpRedisMetrics(ch chan<- prometheus.Metric, rule *conf.Rule) {
	val, ok := rc.Ea.metrics.Load(rule.UniqueId)
	if ok {
		m := val.(*ElasticAlertPrometheusMetrics)
		m.OpRedis.Range(func(key, value any) bool {
			v := value.(OpRedisMetrics)
			labelValues := []string{v.UniqueId, v.Path, v.Cmd, v.Key, strconv.Itoa(v.Status)}
			ch <- prometheus.MustNewConstMetric(rc.OpRedisDesc, prometheus.CounterValue, float64(v.Value), labelValues...)
			return true
		})
	}
}

func (rc *RuleStatusCollector) collectWebhookNotifyMetrics(ch chan<- prometheus.Metric, rule *conf.Rule) {
	val, ok := rc.Ea.metrics.Load(rule.UniqueId)
	if ok {
		m := val.(*ElasticAlertPrometheusMetrics)
		m.WebhookNotify.Range(func(key, value any) bool {
			v := value.(WebhookNotifyMetrics)
			labelValues := []string{v.UniqueId, v.Path, strconv.Itoa(v.Status)}
			ch <- prometheus.MustNewConstMetric(rc.WebhookNotifyDesc, prometheus.CounterValue, float64(v.Value), labelValues...)
			return true
		})
	}
}

func (rc *RuleStatusCollector) collectLinkRedisStatus(ch chan<- prometheus.Metric) {
	_, err := redisx.Client.Ping(ctx).Result()
	v := float64(1)
	if err != nil {
		v = 0
		t := fmt.Sprintf("Ping Redis has error: %s", err.Error())
		logger.Logger.Errorln(t)
	}
	addr := fmt.Sprintf("%s:%d", rc.Ea.appConf.Redis.Addr, rc.Ea.appConf.Redis.Port)
	ch <- prometheus.MustNewConstMetric(rc.LinkRedisDesc, prometheus.GaugeValue, v, addr)
}

func (rc *RuleStatusCollector) collectAppInfo(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(rc.AppInfoDesc, prometheus.GaugeValue, float64(1), Version)
}

func (rc *RuleStatusCollector) collectRuleStatus(ch chan<- prometheus.Metric, rule *conf.Rule) {
	v := float64(0)
	if rule.Enabled {
		v = 1
	}
	labels := []string{
		rule.UniqueId,
		rule.FilePath,
		strings.Join(rule.ES.Addresses, ","),
		rule.Index, strconv.Itoa(rule.RunEvery.GetSeconds()),
		rule.Query.Type,
	}
	ch <- prometheus.MustNewConstMetric(rc.RuleStatusDesc, prometheus.GaugeValue, v, labels...)
}

func NewRuleStatusCollector(ea *ElasticAlert) *RuleStatusCollector {
	return &RuleStatusCollector{
		Ea: ea,
		RuleStatusDesc: prometheus.NewDesc(
			ea.buildFQName("rule"),
			"Show rule status: enabled(1)、disabled(0)",
			[]string{"unique_id", "path", "es_address", "index", "run_every", "type"},
			prometheus.Labels{}),
		AppInfoDesc: prometheus.NewDesc(
			ea.buildFQName("info"),
			"Information about the application",
			[]string{"version"},
			prometheus.Labels{},
		),
		LinkRedisDesc: prometheus.NewDesc(
			ea.buildFQName("link_redis"),
			"Application link redis status: healthy(1) 、unhealthy(0)",
			[]string{"addr"},
			prometheus.Labels{},
		),
		QueryDesc: prometheus.NewDesc(
			ea.buildFQName("query"),
			"Show every rule elasticsearch query times",
			[]string{"unique_id", "path", "es_address", "index", "status"},
			prometheus.Labels{},
		),
		OpRedisDesc: prometheus.NewDesc(
			ea.buildFQName("op_redis"),
			"Show operate redis command times",
			[]string{"unique_id", "path", "cmd", "key", "status"},
			prometheus.Labels{},
		),
		WebhookNotifyDesc: prometheus.NewDesc(
			ea.buildFQName("webhook_notify"),
			"Show call webhook notify alert times",
			[]string{"unique_id", "path", "status"},
			prometheus.Labels{},
		),
	}
}
