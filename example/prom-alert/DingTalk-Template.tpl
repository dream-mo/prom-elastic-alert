{{ $var := .externalURL}}{{ $status := .status}}{{ range $k,$v:=.alerts }} {{if eq $status "resolved"}}
## [告警恢复-通知]({{$var}}/alertmanager/)
#### 监控指标: {{$v.labels.alertname}}
{{ if eq $v.labels.severity "warning" }}
#### 告警级别: **<font color="#E6A23C">{{$v.labels.severity}}</font>**
{{ else if eq $v.labels.severity "critical"  }}
#### 告警级别: **<font color="#F56C6C">{{$v.labels.severity}}</font>**
{{ end }}
#### 当前状态: **<font color="#67C23A" size=4>已恢复</font>**
{{ if $v.labels.ipaddr }}
#### 故障主机: {{$v.labels.ipaddr}}
#### 故障业务: {{$v.labels.instance}}
{{ else }}
#### 故障主机: {{$v.labels.instance}}
{{ end }}
* ###### 告警阈值: {{$v.labels.threshold}}
* ###### 开始时间: {{GetCSTtime $v.startsAt}}
* ###### 恢复时间: {{GetCSTtime $v.endsAt}}

#### 告警恢复: <font color="#67C23A">已恢复,{{$v.annotations.description}}</font>
{{ else }}
## [监控告警-通知]({{$var}}/alertmanager/)
#### 监控指标: {{$v.labels.alertname}}
{{ if eq $v.labels.severity "warning" }}
#### 告警级别: **<font color="#E6A23C" size=4>{{$v.labels.severity}}</font>**
#### 当前状态: **<font color="#E6A23C">需要关注</font>**
{{ else if eq $v.labels.severity "critical"  }}
#### 告警级别: **<font color="#F56C6C" size=4>{{$v.labels.severity}}</font>**
#### 当前状态: **<font color="#F56C6C">需要处理</font>**
{{ end }}
{{ if $v.labels.ipaddr  }}
#### 故障主机: {{$v.labels.ipaddr}}
#### 故障业务: {{$v.labels.instance}}
{{ else }}
#### 故障主机: {{$v.labels.instance}}
{{ end }}
* ###### 告警阈值: {{$v.labels.threshold}}
* ###### 持续时间: {{$v.labels.for_time}}
* ###### 触发时间: {{GetCSTtime $v.startsAt}}
{{ if eq $v.labels.severity "warning" }}
#### 告警触发: <font color="#E6A23C">{{$v.annotations.description}}</font>
{{ else if eq $v.labels.severity "critical" }}
#### 告警触发: <font color="#F56C6C">{{$v.annotations.description}}</font>
{{ end }}
{{ end }}
{{ end }}