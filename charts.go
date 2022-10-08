package chartmuseum

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

const (
	repoUrlTpl         = "api/%s/charts"
	chartUrlTpl        = "api/%s/charts/%s"
	chartVersionUrlTpl = "api/%s/charts/%s/%s"
	downloadUrlTpl     = "%s/charts/%s-%s.tgz"
)

type ChartService struct {
	client *Client
}

type ChartOption struct {
	Name *string `json:"chart-name,omitempty"`
}

type ChartVersionOption struct {
	ChartOption
	Version *string `json:"chart-version,omitempty"`
}

func NewChartOption(name string) ChartOption {
	return ChartOption{
		Name: &name,
	}
}

func NewChartVersionOption(name string, version string) ChartVersionOption {
	return ChartVersionOption{
		ChartOption: NewChartOption(name),
		Version:     &version,
	}
}

func (c *ChartService) ListCharts(repo string, options ...RequestOptionFunc) (*map[string]helmrepo.ChartVersions, *Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf(repoUrlTpl, repoUrl)

	req, err := c.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	cvsMap := map[string]helmrepo.ChartVersions{}
	resp, err := c.client.Do(req, &cvsMap)
	if err != nil {
		return nil, resp, err
	}
	return &cvsMap, resp, err
}

func (c *ChartService) ListVersions(repo string, chartOptions ChartOption, options ...RequestOptionFunc) (*helmrepo.ChartVersions, *Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf(chartUrlTpl, repoUrl, *chartOptions.Name)

	req, err := c.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	cvs := helmrepo.ChartVersions{}
	resp, err := c.client.Do(req, &cvs)
	if err != nil {
		return nil, resp, err
	}
	return &cvs, resp, err
}
func (c *ChartService) GetVersion(repo string, chartVersionOptions ChartVersionOption, options ...RequestOptionFunc) (*helmrepo.ChartVersion, *Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return nil, nil, err
	}

	u := fmt.Sprintf(chartVersionUrlTpl, repoUrl, *chartVersionOptions.Name, *chartVersionOptions.Version)

	req, err := c.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	cv := helmrepo.ChartVersion{}
	resp, err := c.client.Do(req, &cv)
	if err != nil {
		return nil, resp, err
	}

	return &cv, resp, err
}
func (c *ChartService) IsExist(repo string, chartOptions ChartOption, options ...RequestOptionFunc) (bool, *Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return false, nil, err
	}

	u := fmt.Sprintf(chartUrlTpl, repoUrl, *chartOptions.Name)

	req, err := c.client.NewRequest(http.MethodHead, u, nil, options)
	if err != nil {
		return false, nil, err
	}

	resp, err := c.client.Do(req, nil)
	if err != nil {
		return false, resp, err
	}
	return true, resp, err
}

func (c *ChartService) IsExistVersion(repo string, chartVersionOptions ChartVersionOption, options ...RequestOptionFunc) (bool, *Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return false, nil, err
	}

	u := fmt.Sprintf(chartVersionUrlTpl, repoUrl, *chartVersionOptions.Name, *chartVersionOptions.Version)

	req, err := c.client.NewRequest(http.MethodHead, u, nil, options)
	if err != nil {
		return false, nil, err
	}

	resp, err := c.client.Do(req, nil)
	if err != nil {
		return false, resp, err
	}
	return true, resp, err
}

func (c *ChartService) DownloadChart(repo string, dest string, chartVersionOptions ChartVersionOption, options ...RequestOptionFunc) (*Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(downloadUrlTpl, repoUrl, *chartVersionOptions.Name, *chartVersionOptions.Version)
	data := new(bytes.Buffer)

	req, err := c.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req, data)
	if err != nil {
		return resp, err
	}

	destFile := filepath.Join(dest, filepath.Base(u))

	if err := AtomicWriteFile(destFile, data, 0644); err != nil {
		return resp, err
	}
	return resp, err
}

func (c *ChartService) UploadChart(repo, chartFilePath string, options ...RequestOptionFunc) (*Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(repoUrlTpl, repoUrl)

	file, err := os.Open(chartFilePath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if stat.IsDir() {
		return nil, errors.New("can't be update a directory")
	}
	mediaType, _ := detectContentType(file)
	options = append(options, WithUpload(mediaType, stat.Size()))
	req, err := c.client.NewRequest(http.MethodPost, u, file, options)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, err
}

func (c *ChartService) DeleteChart(repo string, chartVersionOptions ChartVersionOption, options ...RequestOptionFunc) (*Response, error) {
	repoUrl, err := parseRepoUrl(repo)
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf(chartVersionUrlTpl, repoUrl, *chartVersionOptions.Name, *chartVersionOptions.Version)

	req, err := c.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req, nil)
	if err != nil {
		return resp, err
	}
	return resp, err
}

func (c *ChartService) GetLatestChartVersionWithRegex(repo string, chart ChartOption, regex string, options ...RequestOptionFunc) (version string, err error) {
	version = ""
	versions, _, err := c.ListVersions(repo, chart, options...)
	if err != nil {
		return version, err
	}
	if len(*versions) > 0 {
		reg := regexp.MustCompile(regex)
		for _, v := range *versions {
			tagName := v.Version
			if reg.MatchString(tagName) {
				if version == "" {
					version = tagName
				}
				// log.Debugf("%s version %s", app, targetVersion)
				if VersionOrdinal(version) < VersionOrdinal(tagName) {
					version = tagName
				}
			}
		}
	}

	return version, err
}

func VersionOrdinal(version string) string {
	// ISO/IEC 14651:2011
	const maxByte = 1<<8 - 1
	vo := make([]byte, 0, len(version)+8)
	j := -1
	for i := 0; i < len(version); i++ {
		b := version[i]
		if '0' > b || b > '9' {
			vo = append(vo, b)
			j = -1
			continue
		}
		if j == -1 {
			vo = append(vo, 0x00)
			j = len(vo) - 1
		}
		if vo[j] == 1 && vo[j+1] == '0' {
			vo[j+1] = b
			continue
		}
		if vo[j]+1 > maxByte {
			panic("VersionOrdinal: invalid version")
		}
		vo = append(vo, b)
		vo[j]++
	}
	return string(vo)
}

// detectContentType returns a valid content-type and "application/octet-stream" if error or no match
func detectContentType(file *os.File) (string, error) {
	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return "application/octet-stream", err
	}

	// Reset the read pointer.
	file.Seek(0, 0)

	// Always returns a valid content-type and "application/octet-stream" if no others seemed to match.
	return http.DetectContentType(buffer), nil
}
