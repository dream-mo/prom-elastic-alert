package exporter

import (
	"context"
	"fmt"
	"github.com/openinsight-proj/elastic-alert/pkg/boot"
	"strconv"
	"strings"
	"sync"

	"github.com/openinsight-proj/elastic-alert/pkg/conf"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	redisx "github.com/openinsight-proj/elastic-alert/pkg/utils/redis"
	"github.com/prometheus/client_golang/prometheus"
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
	Ea                 *boot.ElasticAlert
	AppInfoDesc        *prometheus.Desc
	RuleStatusDesc     *prometheus.Desc
	LinkRedisDesc      *prometheus.Desc
	QueryDesc          *prometheus.Desc
	OpRedisDesc        *prometheus.Desc
	WebhookNotifyDesc  *prometheus.Desc
	ElasticMetricsDesc *prometheus.Desc
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
	rc.Ea.Rules.Range(func(key, value any) bool {
		rule := value.(*conf.Rule)
		rc.collectRuleStatus(ch, rule)
		rc.collectQueryStatusMetrics(ch, rule)
		rc.collectOpRedisMetrics(ch, rule)
		rc.collectWebhookNotifyMetrics(ch, rule)
		//TODO: elastic-alert-metrics
		return true
	})
}

func (rc *RuleStatusCollector) collectQueryStatusMetrics(ch chan<- prometheus.Metric, rule *conf.Rule) {
	val, ok := rc.Ea.Metrics.Load(rule.UniqueId)
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

func (rc *RuleStatusCollector) collectElasticAlertMetrics(ch chan<- prometheus.Metric, rule *conf.Rule) {

}

func (rc *RuleStatusCollector) collectOpRedisMetrics(ch chan<- prometheus.Metric, rule *conf.Rule) {
	val, ok := rc.Ea.Metrics.Load(rule.UniqueId)
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
	val, ok := rc.Ea.Metrics.Load(rule.UniqueId)
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
	_, err := redisx.Client.Ping(context.Background()).Result()
	v := float64(1)
	if err != nil {
		v = 0
		t := fmt.Sprintf("Ping Redis has error: %s", err.Error())
		logger.Logger.Errorln(t)
	}
	addr := fmt.Sprintf("%s:%d", rc.Ea.AppConf.Redis.Addr, rc.Ea.AppConf.Redis.Port)
	ch <- prometheus.MustNewConstMetric(rc.LinkRedisDesc, prometheus.GaugeValue, v, addr)
}

func (rc *RuleStatusCollector) collectAppInfo(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(rc.AppInfoDesc, prometheus.GaugeValue, float64(1), boot.Version)
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

func NewRuleStatusCollector(ea *boot.ElasticAlert) *RuleStatusCollector {
	return &RuleStatusCollector{
		Ea: ea,
		RuleStatusDesc: prometheus.NewDesc(
			ea.BuildFQName("rule"),
			"Show rule status: enabled(1)、disabled(0)",
			[]string{"unique_id", "path", "es_address", "index", "run_every", "type"},
			prometheus.Labels{}),
		AppInfoDesc: prometheus.NewDesc(
			ea.BuildFQName("info"),
			"Information about the application",
			[]string{"version"},
			prometheus.Labels{},
		),
		LinkRedisDesc: prometheus.NewDesc(
			ea.BuildFQName("link_redis"),
			"Application link redis status: healthy(1) 、unhealthy(0)",
			[]string{"addr"},
			prometheus.Labels{},
		),
		QueryDesc: prometheus.NewDesc(
			ea.BuildFQName("query"),
			"Show every rule elasticsearch query times",
			[]string{"unique_id", "path", "es_address", "index", "status"},
			prometheus.Labels{},
		),
		OpRedisDesc: prometheus.NewDesc(
			ea.BuildFQName("op_redis"),
			"Show operate redis command times",
			[]string{"unique_id", "path", "cmd", "key", "status"},
			prometheus.Labels{},
		),
		WebhookNotifyDesc: prometheus.NewDesc(
			ea.BuildFQName("webhook_notify"),
			"Show call webhook notify alert times",
			[]string{"unique_id", "path", "status"},
			prometheus.Labels{},
		),
		ElasticMetricsDesc: prometheus.NewDesc(
			ea.BuildFQName("elastic_metrics_total"),
			"Counter for each reported match",
			[]string{"unique_id", "path", "status"},
			prometheus.Labels{},
		),
	}
}
