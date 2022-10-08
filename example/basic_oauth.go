package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yidaqiang/go-chartmuseum"
)

const (
	chartmuseumServer = "https://chart.example.com"
	chartRepo         = "test/repo"
	username          = "admin"
	password          = "password"
)

func basicAuthExample() {
	client, err := chartmuseum.NewBasicAuthClient(username, password, chartmuseum.WithBaseURL(chartmuseumServer))
	if err != nil {
		logrus.Error(err)
	}
	charts, _, err := client.Charts.ListCharts(chartRepo)
	if err != nil {
		return
	}
	fmt.Printf("found %d charts", len(*charts))
}
