<h1 align="center">go-chartmuseum</h1>

<div align="center">
ChartMuseum API Go 语言库
</div>

[English](./README.md) | [中文简体](./README-zh_CN.md)

## ✨ 特性

- 🌈 对 ChartMuseum API 的Go语言封装。
- 📦 简洁明了的错误处理。
- 🛡 的完善用例测试。

## API

----

Helm Chart Repository

- [ ] `GET /index.yaml`  - retrieved when you run `helm repo add chartmuseum http://localhost:8080/`
- [ ] `GET /charts/mychart-0.1.0.tgz`  retrieved when you run `helm install chartmuseum/mychart`
- [ ] `GET /charts/mychart-0.1.0.tgz.prov`  - retrieved when you run `helm install` with the `--verify flag`

Chart Manipulation

- [x] `POST /api/charts` - upload a new chart version
- [ ] `POST /api/prov` - upload a new provenance file
- [x] `DELETE /api/charts/<name>/<version>` - delete a chart version (and corresponding provenance file)
- [x] `GET /api/charts` - list all charts
- [x] `GET /api/charts/<name>` - list all versions of a chart
- [x] `GET /api/charts/<name>/<version>` - describe a chart version
- [x] `HEAD /api/charts/<name>` - check if chart exists (any versions)
- [x] `HEAD /api/charts/<name>/<version>` - check if chart version exists

Server Info

- [ ] `GET /` - HTML welcome page
- [ ] `GET /info` - returns current ChartMuseum version
- [ ] `GET /health` - returns 200 OK

## 📦 安装

```bash
go get github.com/yidaqiang/go-chartmuseum
```

## 🔨 示例

```go
package main

import (
	"fmt"
	"github.com/yidaqiang/go-chartmuseum"
)

const (
	chartmuseumServer = "https://chart.example.com"
	chartRepo         = "test/repo"
	username          = "admin"
	password          = "password"
)

func main() {
	client, err := chartmuseum.NewBasicAuthClient(username, password, chartmuseum.WithBaseURL(chartmuseumServer))
	if err != nil {
		fmt.Error(err)
	}
	charts, _, err := client.Charts.ListCharts(chartRepo)
	if err != nil {
		return
	}
	fmt.Printf("found %d charts", len(*charts))
}
```

## ⌨ 本地开发

clone locally:

```bash
$ git clone git@github.com:yidaqiang/go-chartmuseum.git
$ cd go-chartmuseum
$ go mod tidy
```

## 🤝 参与共建

请参考[贡献指南]()。

> 强烈推荐阅读 [《提问的智慧》](https://github.com/ryanhanwu/How-To-Ask-Questions-The-Smart-Way)、[《如何向开源社区提问题》](https://github.com/seajs/seajs/issues/545) 和 [《如何有效地报告 Bug》](http://www.chiark.greenend.org.uk/%7Esgtatham/bugs-cn.html)、[《如何向开源项目提交无法解答的问题》](https://zhuanlan.zhihu.com/p/25795393)，更好的问题更容易获得帮助。

## ❤ 赞助者



