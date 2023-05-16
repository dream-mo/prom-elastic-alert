package conf

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dream-mo/prom-elastic-alert/utils"
	"github.com/dream-mo/prom-elastic-alert/utils/xtime"
)

type Rule struct {
	UniqueId string          `yaml:"unique_id"`
	Enabled  bool            `yaml:"enabled" default:"true"`
	ES       EsConfig        `yaml:"es"`
	Index    string          `yaml:"index"`
	RunEvery xtime.TimeLimit `yaml:"run_every"`
	Query    struct {
		Type   string `yaml:"type"`
		Config struct {
			Timeframe xtime.TimeLimit `yaml:"timeframe"`
			NumEvents uint            `yaml:"num_events"`
		} `yaml:"config"`
		QueryString string            `yaml:"query_string"`
		Labels      map[string]string `yaml:"labels"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"query"`
	RawContent string
	FilePath   string
}

func (rl *Rule) GetQueryStringDSL(from int, size int, start time.Time, end time.Time) string {
	q := `
{   
	"size": 0,
    "query":{
        "bool":{
            "must":[
                {
                    "query_string":{
                        "query": %s
                    }
                },
                {
                    "range":{
                        "@timestamp":{
                            "format":"strict_date_optional_time",
                            "gte":"%s",
                            "lte":"%s"
                        }
                    }
                }
            ]
        }
    },
    "sort":[
        {
            "@timestamp":{
                "order":"asc"
            }
        }
    ],
	"from": %d,
    "size": %d
}
    `
	dsl := fmt.Sprintf(q, strconv.Quote(rl.Query.QueryString), xtime.TimeFormatISO8601(start), xtime.TimeFormatISO8601(end), from, size)
	return dsl
}

func (rl *Rule) GetQueryStringCountDSL(start time.Time, end time.Time) string {
	q := `
{
	"size": 0,
    "query":{
        "bool":{
            "must":[
                {
                    "query_string":{
                        "query": %s
                    }
                },
                {
                    "range":{
                        "@timestamp":{
                            "format":"strict_date_optional_time",
                            "gte":"%s",
                            "lte":"%s"
                        }
                    }
                }
            ]
        }
    }
}
    `
	dsl := fmt.Sprintf(q, strconv.Quote(rl.Query.QueryString), xtime.TimeFormatISO8601(start), xtime.TimeFormatISO8601(end))
	return dsl
}

func (rl *Rule) GetMetricsQueryFingerprint(statusCode int) string {
	f := []string{rl.UniqueId, rl.FilePath, rl.GetEsAddress(), rl.Index, strconv.Itoa(statusCode)}
	return utils.MD5(strings.Join(f, ""))
}

func (rl *Rule) GetMetricsOpRedisFingerprint(cmd string, key string, statusCode int) string {
	return GetMetricsOpRedisFingerprint(rl.UniqueId, rl.FilePath, cmd, key, statusCode)
}

func (rl *Rule) GetEsAddress() string {
	return strings.Join(rl.ES.Addresses, ",")
}

func GetMetricsOpRedisFingerprint(uniqueId string, path string, cmd string, key string, statusCode int) string {
	f := []string{uniqueId, path, cmd, key, strconv.Itoa(statusCode)}
	return utils.MD5(strings.Join(f, ""))
}

func GetMetricsWebhookNotifyFingerprint(uniqueId string, path string, statusCode int) string {
	f := []string{uniqueId, path, strconv.Itoa(statusCode)}
	return utils.MD5(strings.Join(f, ""))
}

func BuildFindByIdsDSLBody(ids []string) string {
	m := map[string]any{
		"query": map[string]any{
			"ids": map[string]any{
				"values": ids,
			},
		},
		"sort": []map[string]any{
			{
				"@timestamp": map[string]string{
					"order": "asc",
				},
			},
		},
	}
	bs, _ := json.Marshal(m)
	return string(bs)
}

var AppYamlSchema = `
type: object
required: []
properties:
  exporter:
    type: object
    required: []
    properties:
      enabled: {type: boolean}
      listen_addr: {type: string}
  loader:
    type: object
    required: []
    properties:
      type: {type: string, enum: ["FileLoader"]}
      config: {type: object, required: [], properties: {rules_folder: {type: string}, rules_folder_recursion: {type: boolean}}}
  alert:
    type: object
    required: []
    properties:
      alertmanager: {type: object, required: [], properties: {url: {type: string}, basic_auth: {type: object, required: [], properties: {username: {type: string}, password: {type: string}}}}}
      generator: {type: object, required: [], properties: {base_url: {type: string}, expire: {type: object, required: [], properties: {days: {type: number}}}}}
  redis:
    type: object
    required: []
    properties:
      addr: {type: string}
      port: {type: number}
      password: {type: string}
      db: {type: number}
  run_every:
    type: object
    required: []
    properties:
      seconds: {type: number}
      minutes: {type: number}
      days: {type: number}
  buffer_time:
    type: object
    required: []
    properties:
      seconds: {type: number}
      minutes: {type: number}
      days: {type: number}
  alert_time_limit:
    type: object
    required: []
    properties:
      seconds: {type: number}
      minutes: {type: number}
      days: {type: number}
  max_scrolling_count:
    type: number
`

var RuleYamlSchema = `
type: object
required: ["unique_id", "es", "index", "run_every", "query"]
properties:
  unique_id:
    type: string
  enabled:
    type: boolean
  es:
    type: object
    required: []
    properties:
      addresses: {type: array, items: {type: string}}
      username: {type: string}
      password: {type: string}
      version: {type: string, enum: ["v7", "v8"]}
  index:
    type: string
  run_every:
    type: object
    required: []
    properties:
      seconds: {type: number}
      minutes: {type: number}
      days: {type: number}
  query:
    type: object
    required: ["type", "query_string", "config", "labels", "annotations"]
    properties:
      type: {type: string, enum: ["frequency"]}
      query_string: {type: string}
      config: {type: object, required: ["timeframe", "num_events"], properties: {timeframe: {type: object, required: [], properties: {minutes: {type: number}}}, num_events: {type: number}}}
      labels: {type: object, required: ["alertname"], properties: {alertname: {type: string}, instance: {type: string}, severity: {type: string}, for_time: {type: string}, threshold: {type: string}}}
      annotations: {type: object, required: [], properties: {description: {type: string}, summary: {type: string}}}
`
