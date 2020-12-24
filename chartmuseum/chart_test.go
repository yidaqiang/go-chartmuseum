package chartmuseum

import (
	"context"
	helm_repo "helm.sh/helm/v3/pkg/repo"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestChartService_ListCharts(t *testing.T) {
	testCase := []struct {
		Path   string
		Result map[string]helm_repo.ChartVersions
	}{
		{
			"c7n/test",
			map[string]helm_repo.ChartVersions{},
		},
	}

	client, _ := NewClient("http://chart.example.com", nil)
	for _, c := range testCase {
		result := map[string]helm_repo.ChartVersions{}
		resp, err := client.ChartService.ListCharts(context.Background(), c.Path, &result)
		if err != nil {
			t.Error(err)
		}
		t.Log(resp)
	}
}

func TestChartService_DeleteChart(t *testing.T) {
	client, _ := NewClient("http://chart.example.com", nil)

	resp, err := client.ChartService.DeleteChart(context.Background(), "c7n/test", "agile-service", "0.22.1")
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
}

func TestChartService_UploadChart(t *testing.T) {
	client, _ := NewClient("http://chart.example.com", nil)

	filepath.Walk("/tmp/chart", func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, _ := os.Open(path)
		resp, err := client.ChartService.UploadChart(context.Background(), "c7n/test", file)
		if err != nil {
			t.Error(err)
		}
		t.Log(resp.StatusCode)
		return nil
	})
}

func TestChartService_ListCharts2(t *testing.T) {
	repo := "org-review/proj-review"
	username := "root"
	password := "handhand"

	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				req.SetBasicAuth(username, password)
				return nil, nil
			},
		},
	}
	client, _ := NewClient("http://chartmuseum.c7n.devops.hand-china.com", httpClient)
	result := map[string]helm_repo.ChartVersions{}
	resp, _ := client.ChartService.ListCharts(context.Background(), repo, &result)
	t.Log(resp.Message)
}
