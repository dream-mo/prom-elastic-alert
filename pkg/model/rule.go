package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/openinsight-proj/elastic-alert/pkg/utils"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/xtime"
)

type EsConfig struct {
	Addresses   []string `json:"addresses" yaml:"addresses"`
	Username    string   `json:"username" yaml:"username"`
	Password    string   `json:"password" yaml:"password"`
	ConnTimeout uint     `json:"conn_timeout" yaml:"conn_timeout" default:"10"`
	Version     string   `json:"version" yaml:"version" default:"v7"`
}

type Query struct {
	Type   string `json:"type" yaml:"type"`
	Config struct {
		Timeframe xtime.TimeLimit `json:"timeframe" yaml:"timeframe"`
		NumEvents uint            `json:"num_events" yaml:"num_events"`
	} `json:"config" yaml:"config"`
	QueryString string            `json:"query_string" yaml:"query_string"`
	Labels      map[string]string `json:"labels" yaml:"labels"`
	Annotations map[string]string `json:"annotations" yaml:"annotations"`
}

type Rule struct {
	UniqueId   string          `json:"unique_id" yaml:"unique_id"`
	Enabled    bool            `json:"enabled" yaml:"enabled" default:"true"`
	ES         EsConfig        `json:"es" yaml:"es"`
	Index      string          `json:"index" yaml:"index"`
	RunEvery   xtime.TimeLimit `json:"run_every" yaml:"run_every"`
	Query      Query           `json:"query" yaml:"query"`
	RawContent string          `json:"-"`
	FilePath   string          `json:"-"`
}

func (rl *Rule) GetQueryStringDSL(from int, size int, start time.Time, end time.Time) string {
	q := `
{
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

func GetMetricsAlertFingerprint(uniqueId string, path string, key string) string {
	f := []string{uniqueId, path, key}
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
