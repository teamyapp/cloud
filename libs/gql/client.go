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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		c.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
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
		internalErr = &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		c.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	defer res.Body.Close()

	if res.StatusCode > errs.HTTPClientErrors {
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return &errs.Error{
				Code: errs.Unauthenticated,
			}
		case http.StatusForbidden:
			return &errs.Error{
				Code: errs.PermissionDenied,
			}
		case http.StatusNotFound:
			return &errs.Error{
				Code: errs.NotFound,
			}
		default:
			return &errs.Error{
				Code: errs.Unknown,
			}
		}
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.IO,
			EmbedErr: err,
		}
		c.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	err = json.Unmarshal(buf, gqlResponse)
	if err != nil {
		internalErr = &errs.Error{
			Code:     errs.Deserialization,
			EmbedErr: err,
		}
		c.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewClient(dataCollector telemetry.DataCollector, httpClient web.HTTPClient) *Client {
	return &Client{
		dataCollector: dataCollector,
		httpClient:    httpClient,
	}
}
