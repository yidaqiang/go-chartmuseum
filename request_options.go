package chartmuseum

import (
	"context"
	"github.com/hashicorp/go-retryablehttp"
)

// RequestOptionFunc can be passed to all API requests to customize the API request.
type RequestOptionFunc func(*retryablehttp.Request) error

// WithContext runs the request with the provided context
func WithContext(ctx context.Context) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		*req = *req.WithContext(ctx)
		return nil
	}
}

func WithUpload(mediaType string, size int64) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		req.Header.Set("Content-Type", mediaType)

		req.ContentLength = size
		return nil
	}
}
