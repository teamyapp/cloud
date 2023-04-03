package gql

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/web"
)

type graphQLRequest struct {
	Operation *string     `json:"operation"`
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

type GraphQLResponse[Data any, Error any] struct {
	Data   *Data   `json:"data"`
	Errors []Error `json:"errors"`
}

type QueryOptions struct {
	Operation *string     `json:"operation"`
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

type MutationOptions struct {
	Operation *string     `json:"operation"`
	Mutation  string      `json:"mutation"`
	Variables interface{} `json:"variables"`
}

type Client struct {
	dataCollector telemetry.DataCollector
	httpClient    web.HTTPClient
}

func (c *Client) Query(
	ct context.Context,
	endpoint string,
	headers map[string]string,
	queryOptions QueryOptions,
	gqlResponse interface{},
) *errs.Error {
	gqlRequest := graphQLRequest{
		Operation: queryOptions.Operation,
		Query:     queryOptions.Query,
		Variables: queryOptions.Variables,
	}
	return c.sendRequest(ct, endpoint, headers, gqlRequest, gqlResponse)
}

func (c *Client) Mutate(
	ct context.Context,
	endpoint string,
	headers map[string]string,
	mutationOptions MutationOptions,
	gqlResponse interface{},
) *errs.Error {
	gqlRequest := graphQLRequest{
		Operation: mutationOptions.Operation,
		Query:     mutationOptions.Mutation,
		Variables: mutationOptions.Variables,
	}
	return c.sendRequest(ct, endpoint, headers, gqlRequest, gqlResponse)
}

func (c *Client) sendRequest(
	ct context.Context,
	endpoint string,
	headers map[string]string,
	gqlRequest graphQLRequest,
	gqlResponse interface{},
) *errs.Error {
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	for headerKey, headerVal := range headers {
		req.Header.Set(headerKey, headerVal)
	}

	internalErr := web.WriteJSONToRequest(req, gqlRequest)
	if internalErr != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	req = req.WithContext(ct)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	defer res.Body.Close()

	if res.StatusCode > errs.HTTPClientErrors {
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return errs.NewError(errs.Unauthenticated, "unauthenticated")
		case http.StatusForbidden:
			return errs.NewError(errs.PermissionDenied, "permission denied")
		case http.StatusNotFound:
			return errs.NewError(errs.NotFound, "not found")
		default:
			return errs.NewError(errs.Unknown, "unknown")
		}
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return errs.NewError(errs.IO, err.Error())
	}

	err = json.Unmarshal(buf, gqlResponse)
	if err != nil {
		return errs.NewError(errs.Deserialization, err.Error())
	}

	return nil
}

func NewClient(dataCollector telemetry.DataCollector, httpClient web.HTTPClient) *Client {
	return &Client{
		dataCollector: dataCollector,
		httpClient:    httpClient,
	}
}
