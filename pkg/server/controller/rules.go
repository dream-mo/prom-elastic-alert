package controller

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/openinsight-proj/elastic-alert/pkg/server/domain"
	"github.com/openinsight-proj/elastic-alert/pkg/server/serializer"
	"github.com/openinsight-proj/elastic-alert/pkg/server/services"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
)

type RuleCtrl struct {
	RService *services.RuleService
}

// FindRule query one rule or batch query
func (rc *RuleCtrl) FindRule(c *gin.Context) {
	policy := c.Query("policy")
	uniqueIds := c.Query("rule")
	ids := strings.Split(uniqueIds, ",")
	rules, err := rc.RService.ListRules(policy, ids)
	if err != nil {
		logger.Logger.Errorf("fetch rules error: %v", err)
		serializer.ErrorRes(c, serializer.ErrCopier, fmt.Sprintf("fetch rules error:%v", err))
		return
	}

	serializer.SuccessDataRes(c, rules, fmt.Sprintf("get rules success"))
}

// FindAllRules query all rules
func (rc *RuleCtrl) FindAllRules(c *gin.Context) {
	rules, err := rc.RService.ListRules("", []string{})
	if err != nil {
		logger.Logger.Errorf("fetch rules error: %v", err)
		serializer.ErrorRes(c, serializer.ErrCopier, fmt.Sprintf("fetch rules error:%v", err))
		return
	}

	serializer.SuccessDataRes(c, rules, fmt.Sprintf("get all rules success"))
}

// DeleteRule delete rule or batch delete
func (rc *RuleCtrl) DeleteRule(c *gin.Context) {
	policy := c.Query("policy")
	uniqueIds := c.Query("rule")
	ids := strings.Split(uniqueIds, ",")
	var err error

	if uniqueIds == "" {
		// delete by policy
		err = rc.RService.BatchDeleteRuleByPolicy(policy)
	} else {
		if len(ids) > 1 {
			err = rc.RService.BatchDeleteRule(policy, ids)
		} else {
			err = rc.RService.DeleteRule(policy, ids[0])
		}
	}

	if err != nil {
		logger.Logger.Errorf("delete rule error: %v", err)
		serializer.ErrorRes(c, serializer.ErrDelete, fmt.Sprintf("delete rule error:%v", err))
		return
	}

	serializer.SuccessDataRes(c, "ok", fmt.Sprintf("delete rule success"))
}

// CreateOrUpdateRule create one rule or update by upsert data
func (rc *RuleCtrl) CreateOrUpdateRule(c *gin.Context) {
	var ruleBody domain.Rule
	err := c.BindJSON(&ruleBody)
	if err != nil {
		logger.Logger.Errorf("bind json body error: %v", err)
		serializer.ErrorRes(c, serializer.ErrBindJson, fmt.Sprintf("bind json body error:%v", err))
		return
	}

	err = rc.RService.CreateOrUpdateRule(&ruleBody)
	if err != nil {
		logger.Logger.Errorf("create rule error: %v", err)
		serializer.ErrorRes(c, serializer.ErrCreate, fmt.Sprintf("create rule error:%v", err))
		return
	}

	serializer.SuccessDataRes(c, "ok", fmt.Sprintf("create rule success"))
	return
}

// BatchCreateOrUpdateRule batch create one rule or update by upsert data
func (rc *RuleCtrl) BatchCreateOrUpdateRule(c *gin.Context) {
	var rulesBody []*domain.Rule
	err := c.BindJSON(&rulesBody)
	if err != nil {
		logger.Logger.Errorf("bind json body error: %v", err)
		serializer.ErrorRes(c, serializer.ErrBindJson, fmt.Sprintf("bind json body error:%v", err))
		return
	}

	err = rc.RService.BatchCreateOrUpdateRule(rulesBody)
	if err != nil {
		logger.Logger.Errorf("batch create rule error: %v", err)
		serializer.ErrorRes(c, serializer.ErrCreate, fmt.Sprintf("batch create rule error:%v", err))
		return
	}

	serializer.SuccessDataRes(c, "ok", fmt.Sprintf("batch create rule success"))
	return
}
