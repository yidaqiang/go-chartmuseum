package chartmuseum

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	helm_repo "helm.sh/helm/v3/pkg/repo"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

type ChartService service

// UploadChart uploads a Helm chart to a ChartMuseum server
func (s *ChartService) UploadChart(ctx context.Context, path string, file *os.File) (*Response, error) {
	u := fmt.Sprintf("api/%s/charts", path)

	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to access file")
	}
	if stat.IsDir() {
		return nil, errors.New("Chart to upload can't be a directory")
	}
	mediaType := mime.TypeByExtension(filepath.Ext(file.Name()))

	req, err := s.client.NewUploadRequest(u, file, stat.Size(), mediaType)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating upload request")
	}
	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to do upload request")
	}
	return resp, nil
}


// DeleteChart deletes a Helm chart from a ChartMuseum server
func (s *ChartService) DeleteChart(ctx context.Context, path, name, version string) (*Response, error) {
	u := fmt.Sprintf("api/%s/charts/%s/%s", path, name, version)

	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating delete request")
	}
	resp, err := s.client.Do(ctx, req, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to do delete request")
	}
	return resp, nil
}

func (s *ChartService) ListChartVersion(ctx context.Context, path, name string, charts *helm_repo.ChartVersions) (*Response, error) {
	u := fmt.Sprintf("api/%s/charts/%s", path, name)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating get request")
	}
	resp, err := s.client.Do(ctx, req, charts)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to do get request")
	}
	return resp, nil
}

func (s *ChartService) ListCharts(ctx context.Context, path string, chartsMap *map[string]helm_repo.ChartVersions) (*Response, error) {
	u := fmt.Sprintf("api/%s/charts", path)

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating get request")
	}
	resp, err := s.client.Do(ctx, req, chartsMap)
	if err != nil {
		return resp, errors.Wrap(err, "Failed to do get request")
	}
	return resp, nil
}

func (s *ChartService) IsExist(ctx context.Context, path, name, version string) (bool, error) {
	u := fmt.Sprintf("api/%s/charts/%s/%s", path, name, version)
	req, err := s.client.NewRequest("HEAD", u, nil)
	if err != nil {
		return false, errors.Wrap(err, "Failed creating get request")
	}
	_ , err = s.client.Do(ctx, req, nil)
	if err != nil {
		return false, errors.Wrap(err, "Failed to do get request")
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

