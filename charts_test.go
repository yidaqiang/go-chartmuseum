package chartmuseum

import (
	"fmt"
	"github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"testing"
)

const (
	server   = "http://localhost:8080"
	user     = "admin"
	password = "password"

	testServer = "https://charts.ydq.io"
	testRepo   = "test"

	tmpPath = "/tmp"
)

var (
	localClient *Client
	testClient  *Client
)

func init() {
	localClient, _ = NewBasicAuthClient(user, password, WithBaseURL(server))

	testClient, _ = NewClient(WithBaseURL(testServer))
}

func TestChartService_ListCharts(t *testing.T) {
	convey.Convey("获取 chart 库的所有应用", t, func() {
		testCases := []struct {
			input    string
			expected int
		}{
			{
				input:    testRepo,
				expected: 3,
			},
		}
		for _, tc := range testCases {
			charts, _, err := testClient.Charts.ListCharts(tc.input)
			if err != nil {
				t.Error(err)
			}
			convey.So(len(*charts), convey.ShouldEqual, tc.expected)
			mysqlCharts := (*charts)["mysql"]
			convey.So(mysqlCharts, convey.ShouldNotBeEmpty)
		}
	})
}

func TestChartService_ListVersions(t *testing.T) {
	convey.Convey("获取 chart 应用的所有版本", t, func() {
		testCases := []struct {
			inputRepo        string
			inputChartOption ChartOption
			expected         int
		}{
			{
				inputRepo:        testRepo,
				inputChartOption: NewChartOption("mysql"),
				expected:         3,
			},
			{
				inputRepo:        testRepo,
				inputChartOption: NewChartOption("redis"),
				expected:         5,
			},
			{
				inputRepo:        testRepo,
				inputChartOption: NewChartOption("postgresql"),
				expected:         2,
			},
		}
		for _, tc := range testCases {
			versions, _, err := testClient.Charts.ListVersions(tc.inputRepo, tc.inputChartOption)
			if err != nil {
				t.Error(err)
			}
			convey.So(len(*versions), convey.ShouldEqual, tc.expected)
			convey.So((*versions)[0].Metadata.Name, convey.ShouldEqual, *(tc.inputChartOption.Name))
		}
	})
}

func TestChartService_GetVersion(t *testing.T) {
	convey.Convey("获取 chart 应用的某个版本", t, func() {
		testCases := []struct {
			inputRepo               string
			inputChartVersionOption ChartVersionOption
			expected                string
		}{
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("mysql", "8.8.19"),
				expected:                "8.8.19",
			},
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("mysql", "9.3.4"),
				expected:                "9.3.4",
			},
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("mysql", "10.0.0"),
				expected:                "",
			},
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("redis", "15.7.6"),
				expected:                "15.7.6",
			},
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("postgresql", "10.16.2"),
				expected:                "10.16.2",
			},
		}
		for _, tc := range testCases {
			version, _, err := testClient.Charts.GetVersion(tc.inputRepo, tc.inputChartVersionOption)
			if err != nil {
				if tc.expected == "" {
					fmt.Print(err)
					continue
				}
			} else {
				t.Error(err)
			}
			convey.So((*version).Version, convey.ShouldEqual, tc.expected)

		}
	})

}

func TestChartService_IsExist(t *testing.T) {
	convey.Convey("验证 chart 应用的版本是存在", t, func() {
		testCases := []struct {
			inputRepo        string
			inputChartOption ChartOption
			expected         bool
		}{
			{
				inputRepo:        testRepo,
				inputChartOption: NewChartOption("mysql"),
				expected:         true,
			},
			{
				inputRepo:        testRepo,
				inputChartOption: NewChartOption("mysql-ha"),
				expected:         false,
			},
		}
		for _, tc := range testCases {
			result, _, err := testClient.Charts.IsExist(tc.inputRepo, tc.inputChartOption)

			if err != nil {
				if !result {
					fmt.Print(err)
					continue
				}
				t.Error(err)
			}
			convey.So(result, convey.ShouldEqual, tc.expected)
		}
	})
}

func TestChartService_IsExistVersion(t *testing.T) {
	convey.Convey("验证 chart 应用的版本是存在", t, func() {
		testCases := []struct {
			inputRepo               string
			inputChartVersionOption ChartVersionOption
			expected                bool
		}{
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("redis", "0.1.0"),
				expected:                false,
			},
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("redis", "17.2.0"),
				expected:                true,
			},
		}
		for _, tc := range testCases {
			result, _, err := testClient.Charts.IsExistVersion(tc.inputRepo, tc.inputChartVersionOption)
			if err != nil {
				if !result {
					fmt.Print(err)
					continue
				}
				t.Error(err)
			}
			convey.So(result, convey.ShouldEqual, tc.expected)
		}
	})
}

func TestChartService_DownloadChart(t *testing.T) {
	convey.Convey("下载 chart", t, func() {
		testCase := []struct {
			inputRepo               string
			inputChartVersionOption ChartVersionOption
			expected                bool
		}{
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("redis", "17.2.0"),
				expected:                true,
			},
			{
				inputRepo:               testRepo,
				inputChartVersionOption: NewChartVersionOption("mysql", "9.3.4"),
				expected:                true,
			},
		}

		for _, tc := range testCase {
			_, err := testClient.Charts.DownloadChart(tc.inputRepo, tmpPath, tc.inputChartVersionOption)
			if err != nil {
				t.Error(err)
			}
			path := fmt.Sprintf("%s/%s-%s.tgz", tmpPath, *tc.inputChartVersionOption.Name, *tc.inputChartVersionOption.Version)
			file, err := ioutil.ReadFile(path)
			if err != nil {
				t.Error(err)
			}
			convey.So(file, convey.ShouldNotBeEmpty)
		}
	})
}

func TestChartService_UploadChart(t *testing.T) {
	convey.Convey("下载 chart", t, func() {
		chartPath := "./testdata/chart-demo-0.1.0.tgz"
		resp, err := localClient.Charts.UploadChart(testRepo, chartPath)
		if err != nil {
			t.Log(err)
		}
		convey.So(resp, convey.ShouldNotBeEmpty)
	})
}

func TestChartService_DeleteChart(t *testing.T) {
	convey.Convey("删除 chart", t, func() {
		chartInfo := NewChartVersionOption("chart-demo", "0.1.0")
		resp, err := localClient.Charts.DeleteChart(testRepo, chartInfo)
		if err != nil {
			t.Log(err)
		}
		convey.So(resp, convey.ShouldNotBeEmpty)
	})
}

func TestChartService_GetLatestChartVersionWithRegex(t *testing.T) {
	convey.Convey("获取最新 chart 版本 ", t, func() {
		regexStr := "^15(.\\d)+$"
		co := NewChartOption("redis")
		version, err := testClient.Charts.GetLatestChartVersionWithRegex(testRepo, co, regexStr)
		if err != nil {
			t.Error(err)
		}
		convey.So(version, convey.ShouldEqual, "15.7.6")
	})
}
