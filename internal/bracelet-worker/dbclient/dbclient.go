package dbclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DBEvent struct {
	Method     string `json:"method"`
	EntityName string `json:"entity_name"`
	EntityData any    `json:"entity_data"`
}

type Client struct {
	dbService  string
	httpClient *http.Client
}

func New(dbService string, httpClient *http.Client) *Client {
	return &Client{
		dbService:  strings.TrimRight(dbService, "/"),
		httpClient: httpClient,
	}
}

func (c *Client) Send(ctx context.Context, event DBEvent, result any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.dbService+"/event",
		bytes.NewReader(payload),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("database service returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	if result == nil || len(bytes.TrimSpace(body)) == 0 || string(bytes.TrimSpace(body)) == "null" {
		return nil
	}

	return json.Unmarshal(body, result)
}
