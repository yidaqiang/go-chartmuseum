package chartmuseum

import (
	"bytes"
	"net/http"
)

type Healthy struct {
	Healthy bool
}

type Version struct {
	Version string
}

type InfoService struct {
	client *Client
}

func (s *InfoService) Index(options ...RequestOptionFunc) (string, error) {
	u := "/"
	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return "", err
	}
	data := new(bytes.Buffer)
	_, err = s.client.Do(req, data)
	if err != nil {
		return "", err
	}
	return data.String(), err
}

func (s *InfoService) Health(options ...RequestOptionFunc) (*Healthy, error) {
	u := "/health"
	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, err
	}
	h := Healthy{}
	_, err = s.client.Do(req, &h)
	if err != nil {
		return nil, err
	}
	return &h, err
}

func (s *InfoService) Info(options ...RequestOptionFunc) (*Version, error) {
	u := "/info"
	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, err
	}
	v := Version{}
	_, err = s.client.Do(req, &v)
	if err != nil {
		return nil, err
	}
	return &v, err
}
