<h1 align="center">go-chartmuseum</h1>

<div align="center">
go library for chartmuseum
</div>

[English](./README.md) | [中文简体](./README-zh_CN.md) 

## ✨ Features

- 🌈 Wrap the chartmuseum api as a go library.
- 📦 Plain and simple error handling.
- 🛡 Perfect test cases.

## API

----

Helm Chart Repository

- [ ] `GET /index.yaml`  - retrieved when you run `helm repo add chartmuseum http://localhost:8080/`
- [x] `GET /charts/mychart-0.1.0.tgz`  retrieved when you run `helm install chartmuseum/mychart`
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

- [x] `GET /` - HTML welcome page
- [x] `GET /info` - returns current ChartMuseum version
- [x] `GET /health` - returns 200 OK

## 📦 Install

```bash
go get github.com/yidaqiang/go-chartmuseum
```

## 🔨 Usage

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

## ⌨️ Development

clone locally:

```bash
$ git clone git@github.com:yidaqiang/go-chartmuseum.git
$ cd go-chartmuseum
$ go mod tidy
```

## 🤝 Contributing

Read our [contributing guide]() and let's build a better antd together.

We welcome all contributions. Please read our [CONTRIBUTING.md](https://github.com/yidaqiang/go-chartmuseum/blob/master/.github/CONTRIBUTING.md) first. You can submit any ideas as [pull requests](https://github.com/yidaqiang/go-chartmuseum/pulls) or as [GitHub issues](https://github.com/yidaqiang/go-chartmuseum/issues). If you'd like to improve code, check out the [Development Instructions]() and have a good time! :)

If you are a collaborator, please follow our [Pull Request principle]() to create a Pull Request with [collaborator template]().

## ❤ Sponsors and Backers



