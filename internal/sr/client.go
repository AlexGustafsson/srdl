package sr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

var DefaultClient = &Client{
	BaseURL: "https://api.sr.se",
	Client:  http.DefaultClient,
}

var (
	ErrNotFound = errors.New("not found")
)

type Client struct {
	BaseURL string
	Client  *http.Client
}

type ListEpisodesInProgramOptions struct {
	Page     int
	PageSize int
}

func (c *Client) ListEpisodesInProgram(ctx context.Context, programID int, options *ListEpisodesInProgramOptions) (*EpisodesPage, error) {
	if options == nil {
		options = &ListEpisodesInProgramOptions{}
	}

	page := options.Page
	if page <= 0 {
		page = 1
	}

	pageSize := options.PageSize
	if pageSize <= 0 {
		pageSize = 30
	}

	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = "/v2/episodes/index"

	query := make(url.Values)
	query.Set("format", "json")
	// NOTE: This seems to be a magic number that's used by the SR app in all
	// requests. Unclear what it does
	query.Set("ondemandaudiotemplateid", "9")
	query.Set("programid", strconv.FormatInt(int64(programID), 10))
	query.Set("page", strconv.FormatInt(int64(page), 10))
	query.Set("size", strconv.FormatInt(int64(pageSize), 10))
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var result EpisodesPage
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetEpisode(ctx context.Context, id int) (*Episode, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = "/v2/episodes/get"

	query := make(url.Values)
	query.Set("format", "json")
	// NOTE: This seems to be a magic number that's used by the SR app in all
	// requests. Unclear what it does
	query.Set("ondemandaudiotemplateid", "9")
	query.Set("rawbody", "true")
	query.Set("id", strconv.FormatInt(int64(id), 10))
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var result struct {
		Episode Episode `json:"episode"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Episode, nil
}
