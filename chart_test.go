package go_chartmuseum

import (
	"context"
	helm_repo "helm.sh/helm/v3/pkg/repo"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	baseUrl  = "https://chart.ydq.io"
	name     = "gitlab-ha"
	version1 = "0.4.3"
	version2 = "0.4.0"
	username = "vista"
	password = "yishuida1023CM"
	repos    = []string{"test"}

	tp = BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}
)

func TestChartService_ListCharts(t *testing.T) {
	testCase := []struct {
		ci     ChartInfo
		Result map[string]helm_repo.ChartVersions
	}{
		{
			ChartInfo{
				Name:    nil,
				Version: nil,
				Repos:   &repos,
			},
			map[string]helm_repo.ChartVersions{},
		},
	}

	client, _ := NewClient(baseUrl, nil)
	for _, c := range testCase {
		result := map[string]helm_repo.ChartVersions{}
		resp, err := client.ChartService.ListCharts(context.Background(), &c.ci, &result)
		if err != nil {
			t.Error(err)
		}
		for k, v := range result {
			t.Logf("%s : %+v", k, v)
		}
		t.Logf("%+v", resp)
	}
}

func TestChartService_DeleteChart(t *testing.T) {
	testCase := []struct {
		ChartInfo
		result bool
	}{

		{
			ChartInfo{
				Name:    &name,
				Version: &version2,
				Repos:   &repos,
			},
			true,
		},
		{
			ChartInfo{
				Name:    &name,
				Version: &version1,
				Repos:   &repos,
			},
			false,
		},
	}

	client, _ := NewClient(baseUrl, tp.Client())
	for _, tc := range testCase {
		resp, err := client.ChartService.DeleteChartVersion(context.Background(), &tc.ChartInfo)
		if err != nil {
			t.Error(err)
		}
		t.Log(resp)
	}

}

func TestChartService_UploadChart(t *testing.T) {
	client, _ := NewClient(baseUrl, tp.Client())

	filepath.Walk("/tmp/chart", func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, _ := os.Open(path)
		ci := ChartInfo{
			Repos: &repos,
		}
		resp, err := client.ChartService.UploadChart(context.Background(), &ci, file)
		if err != nil {
			t.Error(err)
		}
		t.Log(resp.StatusCode)
		return nil
	})
}

func TestChartService_GetCharts(t *testing.T) {
	ci := ChartInfo{
		Name:  &name,
		Repos: &repos,
	}

	client, _ := NewClient(baseUrl, tp.Client())
	result := helm_repo.ChartVersions{}
	resp, _ := client.ChartService.GetCharts(context.Background(), &ci, &result)
	for _, r := range result {
		t.Logf("chart %s version : %+v", r.Metadata.Name, r)
	}
	t.Log(resp.Message)
}

func TestChartService_GetChartVersion(t *testing.T) {
	ci := ChartInfo{
		Name:    &name,
		Version: &version2,
		Repos:   &repos,
	}

	client, _ := NewClient(baseUrl, tp.Client())
	result := helm_repo.ChartVersion{}
	resp, _ := client.ChartService.GetChartVersion(context.Background(), &ci, &result)

	t.Logf("chart %s version: %+v", result.Metadata.Name, result)
	t.Log(resp.Message)
}

func TestChartInfo_String(t *testing.T) {
	name := "devops-service"
	version := "1.0.0"
	repos1 := new([]string)
	repos2 := []string{"foo"}
	repos3 := []string{"foo", "boo", "haha"}

	testCase := []struct {
		ci     ChartInfo
		Result string
	}{
		{
			ChartInfo{
				Name:    &name,
				Version: &version,
				Repos:   repos1,
			},
			"devops-service-1.0.0",
		},
		{
			ChartInfo{
				Name:    &name,
				Version: &version,
				Repos:   &repos2,
			},
			"foo/devops-service-1.0.0",
		},
		{
			ChartInfo{
				Name:    &name,
				Version: &version,
				Repos:   &repos3,
			},
			"foo/boo/haha/devops-service-1.0.0",
		},
	}

	for _, tc := range testCase {
		if tc.ci.String() != tc.Result {
			t.Error("Not the desired result")
		}
	}
}

func TestChartInfo_ReposString(t *testing.T) {
	testCase := []struct {
		repos  []string
		result string
	}{
		{
			[]string{},
			"",
		},
		{
			[]string{"vista"},
			"vista",
		},
		{
			[]string{"vista", "app"},
			"vista/app",
		},
		{
			[]string{"vista", "prod", "app", "running"},
			"vista/prod/app/running",
		},
	}

	for _, tc := range testCase {
		ci := ChartInfo{Repos: &tc.repos}
		if ci.ReposString() != tc.result {
			t.Error("repos not the expected value")
		}
	}
}
