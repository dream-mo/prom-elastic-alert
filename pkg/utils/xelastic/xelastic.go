package xelastic

import (
	"crypto/tls"
	"net/http"

	"github.com/dream-mo/prom-elastic-alert/conf"
	"github.com/dream-mo/prom-elastic-alert/utils/logger"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)

type ElasticClient interface {
	FindByDSL(index string, dsl string, source []string) ([]any, int, int)
	CountByDSL(index string, dsl string) (int, int)
}

func NewElasticClient(esConfig conf.EsConfig, version string) ElasticClient {
	client, err := elasticsearch7.NewClient(elasticsearch7.Config{
		Addresses: esConfig.Addresses,
		Username:  esConfig.Username,
		Password:  esConfig.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	if err != nil {
		logger.Logger.Errorln(err)
		return nil
	}
	c := &ElasticClientV7{
		client: client,
	}
	return c
}
