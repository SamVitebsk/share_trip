package contract

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"share_trip/internal/service"
	"time"

	"github.com/google/uuid"

	config "share_trip/configs"
	"share_trip/internal/clients/contract/gen"
)

type Client struct {
	client *contractgen.ClientWithResponses
}

func NewClient(cfg config.ContractConfig) (*Client, error) {
	httpClient := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &retryTransport{
			base:       http.DefaultTransport,
			retryCount: cfg.RetryCount,
		},
	}

	genClient, err := contractgen.NewClientWithResponses(
		cfg.BaseURL,
		contractgen.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: genClient,
	}, nil
}

func (c *Client) CheckService(ctx context.Context, driverID string, serviceCode string) (service.CheckResult, error) {
	parsedID, err := uuid.Parse(driverID)
	if err != nil {
		return service.CheckResult{}, err
	}

	reqBody := contractgen.CheckServiceAvailabilityJSONRequestBody{
		ClientId:    parsedID,
		ServiceCode: contractgen.ServiceCode(serviceCode),
	}

	resp, err := c.client.CheckServiceAvailabilityWithResponse(ctx, reqBody)
	if err != nil {
		return service.CheckResult{}, err
	}

	if resp.JSON200 == nil {
		return service.CheckResult{}, http.ErrServerClosed
	}

	reason := ""
	if resp.JSON200.Reason != nil {
		reason = string(*resp.JSON200.Reason)
	}

	return service.CheckResult{
		Allowed: resp.JSON200.Allowed,
		Reason:  reason,
	}, nil
}

type retryTransport struct {
	base       http.RoundTripper
	retryCount int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}

	var resp *http.Response
	var err error

	for i := 0; i <= t.retryCount; i++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = t.base.RoundTrip(req)

		if err != nil || resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			if i < t.retryCount {
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return resp, err
		}

		break
	}

	return resp, err
}
