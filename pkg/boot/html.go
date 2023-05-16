package boot

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/dream-mo/prom-elastic-alert/conf"
	redisx "github.com/dream-mo/prom-elastic-alert/utils/redis"
	"github.com/dream-mo/prom-elastic-alert/utils/xelastic"
	"github.com/dream-mo/prom-elastic-alert/utils/xtime"
)

func RenderAlertMessage(writer http.ResponseWriter, request *http.Request) {
	q := request.URL.Query()
	key := q.Get("key")
	if key == "" {
		_, _ = writer.Write([]byte(""))
		return
	} else {
		var message AlertSampleMessage
		v, e := redisx.Client.Get(ctx, key).Result()
		if e != nil {
			_, _ = writer.Write([]byte(""))
		} else {
			err := json.Unmarshal([]byte(v), &message)
			if err != nil {
				_, _ = writer.Write([]byte(""))
			} else {
				t, _ := template.New("index.html").Funcs(template.FuncMap{
					"json": func(v any) string {
						res, _ := json.Marshal(v)
						return string(res)
					},
					"showTime": func(v map[string]any) string {
						m := v["_source"].(map[string]any)
						return xtime.TimeFormatISO8601(xtime.Parse(m["@timestamp"].(string)))
					},
				}).Parse(htmlPage)
				body := conf.BuildFindByIdsDSLBody(message.Ids)
				client := xelastic.NewElasticClient(message.ES, message.ES.Version)
				hits, _, _ := client.FindByDSL(message.Index, body, nil)
				hitsStr, _ := json.Marshal(hits)
				_ = t.Execute(writer, map[string]any{
					"hitsStr": string(hitsStr),
					"hits":    hits,
				})
			}
		}
	}
}

var htmlPage = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width,initial-scale=1,user-scalable=0">
    <title>日志告警详细信息</title>
    <link type="text/css" href="http://summerstyle.github.io/jsonTreeViewer/libs/jsonTree/jsonTree.css" rel="stylesheet"/>
    <link type="text/css" href="https://www.layuicdn.com/layui-v2.7.4/css/layui.css" rel="stylesheet"/>
</head>
<style type="text/css">
    table {
        table-layout: fixed;
    }
    table td {
        word-wrap: break-word;
    }
    .jsontree_node {
        width: 95%;
        word-wrap: break-word;
        display: inline-block;
    }
</style>
<body>
<div style="margin: 10px">
    <blockquote class="layui-elem-quote">告警日志-取样文档</blockquote>
    <table class="layui-table" lay-size="sm">
        <colgroup>
            <col width="110">
            <col width="700">
            <col width="60">
        </colgroup>
        <thead>
        <tr>
            <th>@timestamp</th>
            <th>内容</th>
            <th style="text-align: center">操作</th>
        </tr>
        </thead>
        <tbody>
        {{range $i, $v := .hits}}
        <tr>
            <td>{{showTime $v}}</td>
            <td>{{json $v}}</td>
            <td align="center">
                <button onclick='jsonViewer({{$i}})' class="layui-btn layui-btn-xs">查看格式化日志</button>
            </td>
        </tr>
        {{end}}
        </tbody>
    </table>
</div>
</body>
<script type="text/javascript" src="http://summerstyle.github.io/jsonTreeViewer/libs/jsonTree/jsonTree.js"></script>
<script type="text/javascript" src="https://www.layuicdn.com/layui-v2.7.4/layui.js"></script>
<script>

    let listString = {{.hitsStr}}
    function renderJson(jsonContent, wrapperId) {
        // init wrapper
        let wrapper = document.getElementById(wrapperId);
        if (typeof jsonContent === 'string') {
            try {
                console.log(jsonContent)
                var data = JSON.parse(jsonContent)
            } catch (e) {
                console.log(e)
                var data = {"message": jsonContent}
            }
        } else {
            data = jsonContent;
        }
        let tree = jsonTree.create(data, wrapper);
        tree.expand()
    }

    function jsonViewer(index) {
        let lists = JSON.parse(listString)
        let jsonContent = lists[index];
        let indexVal = index + 1
        let title = '【第' + indexVal + '条】日志内容'
        let warp = 'wrapper_' + indexVal
        layer.open({
            id: warp,
            type: 1,
            title: title,
            skin: 'layui-layer-rim',
            area: ['80%', '80%'],
            content: '<div style="overflow-x: scroll" ' + 'id="' + warp + '"></div>',
            shadeClose: true,
            maxmin: true,
            fixed: false
        });
        renderJson(jsonContent, warp)
    }
</script>
</html>
`
