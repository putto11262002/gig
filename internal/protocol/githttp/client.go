package githttp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GitHttpClient struct {
	httpClient *http.Client
}

func NewGitHttpClient() *GitHttpClient {
	return &GitHttpClient{
		httpClient: http.DefaultClient,
	}
}

func (c *GitHttpClient) GetRefs(ctx context.Context, gitUrl string) (*RefDiscReply, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodGet, fmt.Sprintf("%s/info/refs?service=git-upload-pack", stripTrailingSlash(gitUrl)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct http request: %w", err)
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to senf request: %w", err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("server responded with: %s", res.Status)
	}
	if res.Header.Get("Content-Type") != "application/x-git-upload-pack-advertisement" {
		return nil, fmt.Errorf("server responded with invalid content type: %s", res.Header.Get("Content-Type"))
	}
	decoder := NewRefDiscReplyDecoder(res.Body)
	defer res.Body.Close()
	reply, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *GitHttpClient) FetchPack(pr *PackReq, gitUrl string) (io.ReadCloser, error) {
	var encodedPackReq bytes.Buffer
	if err := pr.Encode(&encodedPackReq); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/git-upload-pack", gitUrl), &encodedPackReq)
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")
	if err != nil {
		return nil, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	return res.Body, nil
}

func stripTrailingSlash(url string) string {
	return strings.TrimSuffix(url, "/")
}
