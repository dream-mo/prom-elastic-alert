[English](./README.md) | 中文

elastic-alert 是一个基于查询 Elasticsearch 的告警组件.

## 简介

本项目[elastic-alert](https://github.com/openinsight-proj/elastic-alert)主要是解决市面上针对 Elastic 栈,
没有更多独立的日志告警开源组件可供选择.

虽然之前有使用过[Elastalert](https://github.com/Yelp/elastalert)项目, 但是该项目已经不维护,并且我们在实际使用的过程中遇到了一些问题:

- 1.组件使用Python编写,性能较差有时候造成告警延迟
- 2.告警收敛、告警聚合、收敛等功能较弱
- 3.组件运行数据不能对接 Prometheus 监控体系

本项目灵感来自于[Elastalert](https://github.com/Yelp/elastalert)

## 特性及优点

- 使用 Golang 编写,跨平台、体积小、性能有足够的优势
- 提供了完整的 API
- 自身不实现告警聚合、收敛、分组等,这是 alertmanager 的优势所在,没必要自己再造轮子.引入[PrometheusAlert](https://github.com/feiyu563/PrometheusAlert)实现多类型告警
- 内置 exporter,可以接入 Prometheus 监控体系,查看当前组件运行状态、数据等
- 支持 Elasticsearch7、Elasticsearch8(未来支持)
- 提供现成的 Grafana 面板 json 文件

## 架构图

![架构图](docs/img/architecture.png)

## 告警样例

### 钉钉通知

![钉钉告警图](docs/img/alert.png)

### 告警详情

![告警详情图](docs/img/detail.png)

### Grafana 面板

![Grafana面板图](docs/img/grafana.png)

### 快速入门

通过 [Docker Compose](./CONTRIBUTING.md) 启动

### 文档

详细文档:  [使用文档](docs/document.md)

### 开源许可

遵循 Apache 2.0 协议，详细请查看 [LICENSE](LICENSE)