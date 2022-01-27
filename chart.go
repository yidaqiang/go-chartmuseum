package go_chartmuseum

import "C"
import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/errors"
	helmrepo "helm.sh/helm/v3/pkg/repo"
	"net/http"
	"os"
	"path/filepath"
)

type (
	// ChartInfo holds a chart's name, version1 as well as optional org and repo attributes.
	ChartInfo struct {
		// Name of the Chart
		Name *string
		// Version of the Chart
		Version *string

		// Repos the Chart belongs to
		Repos *[]string
	}
)

func (c ChartInfo) String() string {
	s := fmt.Sprintf("%s-%s", *c.Name, *c.Version)

	repos := c.ReposString()
	if repos != "" {
		s = fmt.Sprintf("%s/%s", repos, s)
	}

	return s
}

func (c ChartInfo) ReposString() string {
	var s string
	if c.Repos != nil && len(*c.Repos) > 0 {
		s = (*c.Repos)[0]
		for i := 1; i < (len(*c.Repos)); i++ {
			s = fmt.Sprintf("%s/%s", s, (*c.Repos)[i])
		}
	}
	return s
}

func (c ChartInfo) CheckRepos() (err error) {
	if c.ReposString() == "" {
		err = errors.Wrap(err, "repo cannot be empty")
	}
	return err
}

func (c ChartInfo) CheckReposAndName() (err error) {
	err = c.CheckRepos()

	if *c.Name == "" {
		err = errors.Wrap(err, "chart name cannot be empty")
	}
	return err
}

func (c ChartInfo) CheckALl() (err error) {
	err = c.CheckReposAndName()

	if *c.Version == "" {
		err = errors.Wrap(err, "chart version1 cannot be empty")
	}
	if err != nil {
		err = errors.Wrap(err, "check ChartInfo property error")
	}
	return err
}

type ChartService service

// UploadChart uploads a Helm chart to a ChartMuseum server
func (s *ChartService) UploadChart(ctx context.Context, ci *ChartInfo, file *os.File) (*Response, error) {
	if err := ci.CheckRepos(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("api/%s/charts", ci.ReposString())

	return s.uploadChartHelper(ctx, u, file)
}

// DeleteChartVersion deletes a version1 of Helm chart from a ChartMuseum server
func (s *ChartService) DeleteChartVersion(ctx context.Context, ci *ChartInfo) (*Response, error) {
	if err := ci.CheckALl(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("api/%s/charts/%s/%s", ci.ReposString(), *ci.Name, *ci.Version)

	return s.deleteChartHelper(ctx, u)
}

// ListCharts List all Helm Chart from a ChartMuseum server
func (s *ChartService) ListCharts(ctx context.Context, ci *ChartInfo, chartsMap *map[string]helmrepo.ChartVersions) (*Response, error) {
	if err := ci.CheckRepos(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("api/%s/charts", ci.ReposString())
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a GET request")
	}
	resp, err := s.client.Do(ctx, req, chartsMap)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to execute the GET request")
	}
	return resp, nil
}

// GetCharts Get a Helm Chart of repo from a ChartMuseum server
func (s *ChartService) GetCharts(ctx context.Context, ci *ChartInfo, charts *helmrepo.ChartVersions) (*Response, error) {
	if err := ci.CheckReposAndName(); err != nil {
		return nil, err
	}
	u := fmt.Sprintf("api/%s/charts/%s", ci.ReposString(), *ci.Name)

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a GET request")
	}
	resp, err := s.client.Do(ctx, req, charts)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to execute the GET request")
	}
	return resp, nil
}

// GetChartVersion Get a Helm Chart version of repo from a ChartMuseum server
func (s *ChartService) GetChartVersion(ctx context.Context, ci *ChartInfo, chart *helmrepo.ChartVersion) (*Response, error) {
	if err := ci.CheckALl(); err != nil {
		return nil, err
	}
	u := fmt.Sprintf("api/%s/charts/%s/%s", ci.ReposString(), *ci.Name, *ci.Version)

	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a GET request")
	}
	resp, err := s.client.Do(ctx, req, chart)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to execute the GET request")
	}
	return resp, nil
}

func (s *ChartService) DownloadChart(ctx context.Context, ci *ChartInfo, dest string) error {
	if err := ci.CheckALl(); err != nil {
		return err
	}

	u := fmt.Sprintf("%s/charts/%s-%s.tgz", ci.ReposString(), *ci.Name, *ci.Version)
	data := new(bytes.Buffer)
	_, err := s.downloadChartHelper(ctx, u, data)
	if err != nil {
		return err
	}

	destfile := filepath.Join(dest, filepath.Base(u))

	if err := AtomicWriteFile(destfile, data, 0644); err != nil {
		return err
	}
	return nil
}

func (s *ChartService) IsExistChart(ctx context.Context, ci *ChartInfo) (bool, error) {
	if err := ci.CheckReposAndName(); err != nil {
		return false, err
	}

	u := fmt.Sprintf("api/%s/charts/%s", ci.ReposString(), *ci.Name)

	return s.isExistChartHelper(ctx, u)
}

func (s *ChartService) IsExistChartVersion(ctx context.Context, ci *ChartInfo) (bool, error) {
	if err := ci.CheckALl(); err != nil {
		return false, err
	}

	u := fmt.Sprintf("api/%s/charts/%s/%s", ci.ReposString(), *ci.Name, *ci.Version)

	return s.isExistChartHelper(ctx, u)
}

// deleteChartHelper prepares and executes the upload request
func (s *ChartService) uploadChartHelper(ctx context.Context, u string, file *os.File) (*Response, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to access file")
	}
	if stat.IsDir() {
		return nil, errors.New("Chart to upload can't be a directory")
	}
	//mediaType := mime.TypeByExtension(filepath.Ext(file.Name()))
	mediaType, _ := detectContentType(file)
	req, err := s.client.NewUploadRequest(u, file, stat.Size(), mediaType)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a POST request")
	}
	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to do upload request")
	}
	return resp, nil
}

func (s *ChartService) downloadChartHelper(ctx context.Context, u string, buffer *bytes.Buffer) (*Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a GET request")
	}

	resp, err := s.client.Do(ctx, req, buffer)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to execute the GET request")
	}

	return resp, err
}

// deleteChartHelper prepares and executes the delete request
func (s *ChartService) deleteChartHelper(ctx context.Context, u string) (*Response, error) {

	req, err := s.client.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create a DELETE request")
	}
	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to execute the DELETE request")
	}
	return resp, nil
}

func (s *ChartService) isExistChartHelper(ctx context.Context, u string) (bool, error) {
	req, err := s.client.NewRequest(http.MethodHead, u, nil)
	if err != nil {
		return false, errors.Wrap(err, "Failed to create a HEAD request")
	}
	_, err = s.client.Do(ctx, req, nil)
	if err != nil {
		return false, errors.Wrap(err, "Failed to execute the HEAD request")
	}
	return true, nil
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
