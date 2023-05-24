package boot

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/openinsight-proj/elastic-alert/pkg/model"

	"github.com/creasty/defaults"
	jsonYaml "github.com/ghodss/yaml"
	"github.com/openinsight-proj/elastic-alert/pkg/utils"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

const (
	RuleFileSuffix = ".rule.yaml"
)

// FileLoader static file loader TODO(jian): use https://github.com/spf13/viper instead.
type FileLoader struct {
	RulesFolder          string `default:"rules"`
	RulesFolderRecursion bool   `default:"true"`
}

func (fl *FileLoader) InjectConfig(config map[string]any) {
	folder, ok := config["rules_folder"]
	if ok {
		folderPath, ok := folder.(string)
		if !ok {
			panic("FileLoader:rules_folder error")
		}
		fl.RulesFolder = folderPath
	}
	r, ok := config["rules_folder_recursion"]
	if ok {
		recursion, ok := r.(bool)
		if ok {
			fl.RulesFolderRecursion = recursion
		}
	}
}

func (fl *FileLoader) GetRules() map[string]*model.Rule {
	rules := map[string]*model.Rule{}
	defer func() {
		logger.Logger.Debugln("end load rule file")
		t := fmt.Sprintf("Total: %d items", len(rules))
		logger.Logger.Infoln(t)
	}()
	logger.Logger.Infoln("Start load rule file")
	path := fl.RulesFolder
	rules = fl.getRulesByPath(path)
	return rules
}

func (fl *FileLoader) getRulesByPath(path string) map[string]*model.Rule {
	rules := map[string]*model.Rule{}
	exist, _ := utils.PathExists(path)
	if !exist {
		logger.Logger.Errorln("rules_folder " + path + " not exists")
		return rules
	}
	if utils.IsDir(path) {
		var files []string
		fl.findRuleFiles(path, &files)
		for _, filePath := range files {
			rule, err := fl.getSingleRule(filePath)
			if err != nil {
				t := fmt.Sprintf("load rule file %s is error: %s", filePath, err)
				logger.Logger.Errorln(t)
				continue
			} else {
				t := fmt.Sprintf("Loading... rule file %s", filePath)
				logger.Logger.Infoln(t)
				rules[rule.UniqueId] = rule
			}
		}
		return rules
	} else {
		rule, err := fl.getSingleRule(path)
		if err != nil {
			t := fmt.Sprintf("%s error: %s", path, err)
			logger.Logger.Errorln(t)
			return rules
		} else {
			rules[rule.UniqueId] = rule
			return rules
		}
	}
}

func (fl *FileLoader) ReloadSchedulerJob(engine *ElasticAlert) {
	logger.Logger.Infoln("scheduler job reloading...")
	path := fl.RulesFolder
	rules := fl.getRulesByPath(path)
	for _, newRule := range rules {
		p := newRule.FilePath
		engine.rules.Store(newRule.UniqueId, newRule)
		_, ok := engine.schedulers.Load(newRule.UniqueId)
		if ok {
			t := fmt.Sprintf("RELOAD %s success!", p)
			logger.Logger.Infoln(t)
			engine.restartJobScheduler(newRule)
		} else {
			t := fmt.Sprintf("ADD %s success!", p)
			logger.Logger.Infoln(t)
			engine.startJobScheduler(newRule)
		}
	}
}

func (fl *FileLoader) getSingleRule(path string) (*model.Rule, error) {
	rule := model.Rule{}
	_ = defaults.Set(rule)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	} else {
		ruleSchemaJson, _ := jsonYaml.YAMLToJSON([]byte(model.RuleYamlSchema))
		ruleSchemaLoader := gojsonschema.NewBytesLoader(ruleSchemaJson)
		ruleConfJson, _ := jsonYaml.YAMLToJSON(content)
		ruleConfLoader := gojsonschema.NewBytesLoader(ruleConfJson)
		res, e := gojsonschema.Validate(ruleSchemaLoader, ruleConfLoader)
		if e != nil {
			errorMsg := "rule config file schema error: " + e.Error()
			return nil, errors.New(errorMsg)
		}
		if !res.Valid() {
			errorMsg := res.Errors()[0].String()
			return nil, errors.New(errorMsg)
		}
		e = yaml.Unmarshal(content, &rule)
		if e != nil {
			return nil, e
		} else {
			rule.RawContent = string(content)
			rule.FilePath = path
			return &rule, nil
		}
	}
}

func (fl *FileLoader) findRuleFiles(path string, files *[]string) {
	if !fl.RulesFolderRecursion {
		lists, _ := os.ReadDir(path)
		for _, item := range lists {
			p := path + "/" + item.Name()
			if strings.HasSuffix(p, RuleFileSuffix) {
				*files = append(*files, p)
			}
		}
	} else {
		lists, _ := os.ReadDir(path)
		for _, item := range lists {
			p := path + "/" + item.Name()
			if utils.IsDir(p) {
				fl.findRuleFiles(p, files)
			} else {
				if strings.HasSuffix(p, RuleFileSuffix) {
					*files = append(*files, p)
				}
			}
		}
	}
}
