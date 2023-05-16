package xelastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dream-mo/prom-elastic-alert/utils/logger"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type ElasticClientV7 struct {
	client *elasticsearch7.Client
}

var ctx = context.Background()

func (ec *ElasticClientV7) FindByDSL(index string, dsl string, source []string) ([]any, int, int) {
	req := esapi.SearchRequest{
		Index:        []string{index},
		DocumentType: []string{"_doc"},
		Body:         strings.NewReader(dsl),
	}
	if source != nil {
		req.Source = source
	}
	dst := &bytes.Buffer{}
	_ = json.Compact(dst, []byte(dsl))
	res, e := req.Do(ctx, ec.client)
	hits := []any{}
	totalValue := 0
	if e != nil {
		t := fmt.Sprintf("%s : %s", index, e.Error())
		logger.Logger.Errorln(t)
		return hits, totalValue, res.StatusCode
	} else {
		m := ec.parseResponseBody(res)
		j, ok := m["hits"]
		if ok {
			hitsVal := j.(map[string]any)
			hits = hitsVal["hits"].([]any)
			total := hitsVal["total"].(map[string]any)
			totalFloat := total["value"].(float64)
			totalValue = int(totalFloat)
		}
		return hits, totalValue, res.StatusCode
	}
}

func (ec *ElasticClientV7) CountByDSL(index string, dsl string) (int, int) {
	req := esapi.CountRequest{
		Index:        []string{index},
		DocumentType: []string{"_doc"},
		Body:         strings.NewReader(dsl),
	}
	res, e := req.Do(ctx, ec.client)
	if e != nil {
		t := fmt.Sprintf("%s : %s", index, e.Error())
		logger.Logger.Errorln(t)
		return 0, res.StatusCode
	} else {
		m := ec.parseResponseBody(res)
		c, ok := m["count"]
		if ok {
			countFloat := c.(float64)
			return int(countFloat), res.StatusCode
		} else {
			return 0, res.StatusCode
		}
	}
}

func (ec *ElasticClientV7) parseResponseBody(resp *esapi.Response) map[string]any {
	s := map[string]any{}
	if !resp.IsError() {
		bs, _ := io.ReadAll(resp.Body)
		if !json.Valid(bs) {
			return s
		} else {
			_ = json.Unmarshal(bs, &s)
		}
	}
	return s
}

func (ec *ElasticClientV7) FindByFilter() {

}
