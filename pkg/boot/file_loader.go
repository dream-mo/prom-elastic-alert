package boot

import (
	"errors"
	"fmt"
	"os"
	BuiltPath "path"
	"strings"

	"github.com/creasty/defaults"
	"github.com/fsnotify/fsnotify"
	jsonYaml "github.com/ghodss/yaml"
	"github.com/openinsight-proj/elastic-alert/pkg/conf"
	"github.com/openinsight-proj/elastic-alert/pkg/utils"
	"github.com/openinsight-proj/elastic-alert/pkg/utils/logger"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

const (
	RuleFileSuffix = ".rule.yaml"
)

type FileLoader struct {
	RulesFolder          string `default:"rules"`
	RulesFolderRecursion bool   `default:"true"`
	fsWatcherDirs        map[string]bool
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
	fl.fsWatcherDirs = make(map[string]bool)
}

func (fl *FileLoader) GetRules() map[string]*conf.Rule {
	rules := map[string]*conf.Rule{}
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

func (fl *FileLoader) getRulesByPath(path string) map[string]*conf.Rule {
	rules := map[string]*conf.Rule{}
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
			fDir := BuiltPath.Dir(filePath)
			if _, ok := fl.fsWatcherDirs[fDir]; !ok {
				fl.fsWatcherDirs[fDir] = true
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
	engine.rules.Range(func(key, value any) bool {
		alertRule := value.(*conf.Rule)
		rule := alertRule
		go fl.handleFileChange(rule.FilePath, engine)
		return true
	})
	for fDir := range fl.fsWatcherDirs {
		go fl.handleDirChange(fDir, engine)
	}
}

func (fl *FileLoader) handleDirChange(path string, engine *ElasticAlert) {
	logger.Logger.Infoln("Listen file watcher dir: " + path)
	FsWatcher(path, func(event *fsnotify.Event, e error) {
		if event.Has(fsnotify.Create) {
			newPath := event.Name
			if ok, _ := utils.PathExists(newPath); !ok {
				return
			}
			if utils.IsDir(newPath) {
				t := fmt.Sprintf("Create new dir %s, reloading...", newPath)
				logger.Logger.Infoln(t)
				if _, ok := fl.fsWatcherDirs[newPath]; !ok {
					if fl.RulesFolderRecursion {
						fl.handleDirChange(newPath, engine)
					} else {
						return
					}
				}
			}
			if !strings.HasSuffix(newPath, RuleFileSuffix) {
				return
			}
			fl.handleFileChange(newPath, engine)
			t := fmt.Sprintf("Create new file %s, reloading...", newPath)
			logger.Logger.Infoln(t)
			rules := fl.getRulesByPath(newPath)
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
	})
}

func (fl *FileLoader) handleFileChange(filePath string, engine *ElasticAlert) {
	FsWatcher(filePath, func(event *fsnotify.Event, e error) {
		if event.Has(fsnotify.Write) {
			t := fmt.Sprintf("File has change %s, reloading...", event.String())
			logger.Logger.Infoln(t)
			newRule, e := fl.getSingleRule(filePath)
			if e != nil {
				t := fmt.Sprintf("RELOAD %s failed reason: %s", filePath, e.Error())
				logger.Logger.Warningln(t)
			} else {
				engine.rules.Store(newRule.UniqueId, newRule)
				engine.restartJobScheduler(newRule)
				t := fmt.Sprintf("RELOAD %s success!", filePath)
				logger.Logger.Infoln(t)
			}
		}
	})
}

func (fl *FileLoader) getSingleRule(path string) (*conf.Rule, error) {
	rule := conf.Rule{}
	_ = defaults.Set(rule)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	} else {
		ruleSchemaJson, _ := jsonYaml.YAMLToJSON([]byte(conf.RuleYamlSchema))
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

// FsWatcher can listen file or dir change, if you get a change event, will call callback function
func FsWatcher(path string, callback func(event *fsnotify.Event, e error)) {

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Logger.Errorln(err)
	}
	defer watcher.Close()

	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				callback(&event, nil)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				t := fmt.Sprintf("fsnotify watcher error %s", err.Error())
				logger.Logger.Errorln(t)
				callback(nil, err)
			}
		}
	}()

	// Add a path.
	err = watcher.Add(path)
	if err != nil {
		logger.Logger.Errorln(err)
	}

	<-make(chan struct{})
}
