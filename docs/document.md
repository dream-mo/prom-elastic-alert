## 安装与运行

### 依赖Redis

- 自行安装Redis或者直接使用现成的Redis即可

### 安装方式

- 1.下载[release](https://github.com/dream-mo/prom-elastic-alert/releases)二进制文件,拷贝config.yaml,运行即可
- 2.进入compose目录,使用docker-compose运行([example](https://github.com/dream-mo/prom-elastic-alert/tree/main/example))
- 3.自行编译, git clone 项目到本地, 之后go build即可


## config.yaml主配置文件详解

- 参考 [docs/config.example.yaml](https://github.com/dream-mo/prom-elastic-alert/blob/main/docs/config.example.yaml)注释

## *.rule.yaml子配置文件详解

- 参考 [docs/example.rule.yaml](https://github.com/dream-mo/prom-elastic-alert/blob/main/docs/example.rule.yaml)注释
- *.rule.yaml配置文件支持热更新加载,更新文件自动reload对应的任务

## PrometheusAlert-钉钉模板

- 查看[example/prom-alert/DingTalk-Template.tpl](https://github.com/dream-mo/prom-elastic-alert/blob/main/example/prom-alert/DingTalk-Template.tpl)

