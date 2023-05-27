package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/xeipuuv/gojsonschema"
	"strings"

	jsonYaml "github.com/ghodss/yaml"
	"github.com/openinsight-proj/elastic-alert/pkg/client"
	"github.com/openinsight-proj/elastic-alert/pkg/model"
	"github.com/openinsight-proj/elastic-alert/pkg/server/domain"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RuleService struct {
	KubeClient *client.KubeClient
}

// validateRuleConfig validate rule through yaml schema
func validateRuleConfig(r *model.Rule) error {
	ruleSchemaJson, _ := jsonYaml.YAMLToJSON([]byte(model.RuleYamlSchema))
	ruleSchemaLoader := gojsonschema.NewBytesLoader(ruleSchemaJson)

	jsonBytes, err := json.Marshal(&r)
	if err != nil {
		errorMsg := "rule config json marshal error: " + err.Error()
		return errors.New(errorMsg)
	}

	ruleConfLoader := gojsonschema.NewBytesLoader(jsonBytes)
	res, e := gojsonschema.Validate(ruleSchemaLoader, ruleConfLoader)
	if e != nil {
		errorMsg := "rule config schema error: " + e.Error()
		return errors.New(errorMsg)
	}

	if !res.Valid() {
		errorMsg := res.Errors()[0].String()
		return errors.New(errorMsg)
	}

	return nil
}

func (rs *RuleService) CreateOrUpdateRule(rule *domain.Rule) error {
	if err := validateRuleConfig(&rule.Rule); err != nil {
		return err
	}
	cfgKey := rs.generateDataKey(rule.Policy, rule.Rule.UniqueId)
	ruleConfig := rule.Rule
	err := rs.upsertData(cfgKey, &ruleConfig)
	if err != nil {
		return err
	}
	return nil
}

func (rs *RuleService) BatchCreateOrUpdateRule(rules []*domain.Rule) error {
	cfgMaps := make(map[string]*model.Rule, len(rules))
	for _, rule := range rules {
		if err := validateRuleConfig(&rule.Rule); err != nil {
			return err
		}
		cfgKey := rs.generateDataKey(rule.Policy, rule.Rule.UniqueId)
		cfgMaps[cfgKey] = &rule.Rule
	}

	err := rs.batchUpsertData(cfgMaps)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RuleService) ListRules(policy string, uniqueIds []string) ([]domain.Rule, error) {
	// fetch configmap and data
	cfgMap, err := rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Get(context.Background(), rs.KubeClient.ConfigmapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	filterCfgKeys := make(map[string]struct{}, len(uniqueIds))
	if policy != "" && len(uniqueIds) > 0 {
		for _, id := range uniqueIds {
			filterCfgKeys[rs.generateDataKey(policy, id)] = struct{}{}
		}
	}

	rules := []domain.Rule{}
	for cfgKey, cfgVal := range cfgMap.Data {
		if len(filterCfgKeys) > 0 {
			if _, found := filterCfgKeys[cfgKey]; !found {
				// do not select this rule
				continue
			}
		}

		policyRule := domain.Rule{}
		rule := model.Rule{}
		jsonBytes, err := jsonYaml.YAMLToJSON([]byte(cfgVal))
		if err != nil {
			logger.Logger.Errorf("ruel yaml to json failed: %v", err)
			return nil, err
		}
		err = json.Unmarshal(jsonBytes, &rule)
		if err != nil {
			logger.Logger.Errorf("failed unmarshal rule: %v", err)
			return nil, err
		}
		policyRule.Policy = rs.splitPolicy(cfgKey)
		policyRule.Rule = rule

		rules = append(rules, policyRule)
	}

	return rules, nil
}

func (rs *RuleService) BatchDeleteRule(policy string, uniqueIds []string) error {

	// fetch configmap and data
	oldCfgMap, err := rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Get(context.Background(), rs.KubeClient.ConfigmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// rule not found, do nothing.
	if oldCfgMap.Data == nil {
		return nil
	}

	for _, id := range uniqueIds {
		delete(oldCfgMap.Data, rs.generateDataKey(policy, id))
	}

	// update configmap
	_, err = rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Update(context.Background(), oldCfgMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (rs *RuleService) BatchDeleteRuleByPolicy(policy string) error {
	// fetch configmap and data
	oldCfgMap, err := rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Get(context.Background(), rs.KubeClient.ConfigmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// rule not found, do nothing.
	if oldCfgMap.Data == nil {
		return nil
	}

	for key, _ := range oldCfgMap.Data {
		if rs.splitPolicy(key) == policy {
			delete(oldCfgMap.Data, key)
		}
	}

	// update configmap
	_, err = rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Update(context.Background(), oldCfgMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (rs *RuleService) DeleteRule(policy, uniqueId string) error {
	cfgKey := rs.generateDataKey(policy, uniqueId)

	// fetch configmap and data
	oldCfgMap, err := rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Get(context.Background(), rs.KubeClient.ConfigmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// rule not found, do nothing.
	if oldCfgMap.Data == nil {
		return nil
	}

	delete(oldCfgMap.Data, cfgKey)

	// update configmap
	_, err = rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Update(context.Background(), oldCfgMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (rs *RuleService) upsertData(dataKey string, newRule *model.Rule) error {
	// fetch configmap and data
	oldCfgMap, err := rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Get(context.Background(), rs.KubeClient.ConfigmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	ruleBytes, err := json.Marshal(newRule)
	if err != nil {
		logger.Logger.Errorf("failed marshal rule: %v", err)
	}

	yamlBytes, err := jsonYaml.JSONToYAML(ruleBytes)
	if err != nil {
		logger.Logger.Errorf("json failed to yaml rule: %v", err)
	}

	if oldCfgMap.Data == nil {
		oldCfgMap.Data = make(map[string]string, 1)
	}
	//insert new rule into configmap data
	oldCfgMap.Data[dataKey] = string(yamlBytes)
	// update configmap
	_, err = rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Update(context.Background(), oldCfgMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (rs *RuleService) batchUpsertData(rules map[string]*model.Rule) error {
	// fetch configmap and data
	oldCfgMap, err := rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Get(context.Background(), rs.KubeClient.ConfigmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if oldCfgMap.Data == nil {
		oldCfgMap.Data = make(map[string]string, len(rules))
	}

	for key, rule := range rules {
		ruleBytes, err := json.Marshal(rule)
		if err != nil {
			logger.Logger.Errorf("failed marshal rule: %v", err)
		}

		yamlBytes, err := jsonYaml.JSONToYAML(ruleBytes)
		if err != nil {
			logger.Logger.Errorf("json failed to yaml rule: %v", err)
		}
		oldCfgMap.Data[key] = string(yamlBytes)
	}
	// update configmap
	_, err = rs.KubeClient.Client.CoreV1().ConfigMaps(rs.KubeClient.Namespace).Update(context.Background(), oldCfgMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (rs *RuleService) generateDataKey(policy, uniqueId string) string {
	return strings.Join([]string{policy, uniqueId}, "-") + ".rule.yaml"
}

func (rs *RuleService) splitPolicy(policyUniqueId string) string {
	return strings.Split(policyUniqueId, "-")[0]
}
