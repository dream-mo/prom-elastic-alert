package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/openinsight-proj/elastic-alert/pkg/boot"
	"github.com/openinsight-proj/elastic-alert/pkg/client"
	"github.com/openinsight-proj/elastic-alert/pkg/conf"
	"github.com/openinsight-proj/elastic-alert/pkg/server/controller"
	"github.com/openinsight-proj/elastic-alert/pkg/server/serializer"
	"github.com/openinsight-proj/elastic-alert/pkg/server/services"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
)

type HttpServer struct {
	ServerConfig *conf.AppConfig
	Ea           *boot.ElasticAlert
	KubeClient   *client.KubeClient
	ruleCtrl     *controller.RuleCtrl
}

func (s *HttpServer) InitHttpServer() {
	if s.ServerConfig.Server.Enabled {
		s.ruleCtrl = &controller.RuleCtrl{
			RService: &services.RuleService{
				KubeClient: s.KubeClient,
			},
		}

		router := s.newRouter()
		err := router.Run(s.ServerConfig.Server.ListenAddr)
		if err != nil {
			logger.Logger.Errorf("init http server failed: %s", err.Error())
		}
	}
}

func (s *HttpServer) newRouter() *gin.Engine {
	r := gin.Default()

	cgRoute := r.Group("-")
	cgRoute.POST("/reload", func(c *gin.Context) {
		s.Ea.Loader.ReloadSchedulerJob(s.Ea)
		serializer.SuccessDataRes(c, "ok", fmt.Sprintf("get all rules success"))
	})

	v1Route := r.Group("/v1")

	{
		v1Route.GET("/rules", s.ruleCtrl.FindAllRules)
		v1Route.GET("/rule", s.ruleCtrl.FindRule)
		v1Route.POST("/rule", s.ruleCtrl.CreateOrUpdateRule)
		v1Route.POST("/rule/batch", s.ruleCtrl.BatchCreateOrUpdateRule)
		v1Route.PUT("/rule", s.ruleCtrl.CreateOrUpdateRule)
		v1Route.DELETE("/rule", s.ruleCtrl.DeleteRule)
	}

	return r
}
