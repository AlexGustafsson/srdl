package sr

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AlexGustafsson/srdl/internal/htmlutil"
	"github.com/AlexGustafsson/srdl/internal/httputil"
	"golang.org/x/net/html"
)

// DefaultClient is the default [Client].
var DefaultClient = &Client{
	BaseURL: "https://api.sr.se",
	Client:  httputil.DefaultClient,
}

var (
	ErrNotFound          = errors.New("not found")
	ErrProgramIDNotFound = errors.New("program id not found")
)

type Client struct {
	// BaseURL is the base URL to the SR APIs.
	BaseURL string
	// Client is the underlying HTTP client to use.
	Client *http.Client
}

type ListEpisodesInProgramOptions struct {
	// Page [1-n]. Defaults to 1.
	Page int
	// PageSize is the number of preferred entries per page.
	PageSize int
}

// ListEpisodesInProgram list episodes of a program.
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

// GetProgram retrieves a program.
func (c *Client) GetProgram(ctx context.Context, id int) (*Program, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = "/v2/programs/" + strconv.FormatInt(int64(id), 10)

	query := make(url.Values)
	query.Set("format", "json")
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
		Program Program `json:"program"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Program, nil
}

// GetEpisode retrieves an episode.
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

// GetProgramID return the program id of a program based on its program's page.
func (c *Client) GetProgramID(ctx context.Context, programPageURL string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, programPageURL, nil)
	if err != nil {
		return -1, err
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return -1, err
	}

	if res.StatusCode == http.StatusNotFound {
		return -1, ErrNotFound
	} else if res.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}
	defer res.Body.Close()

	html, err := html.Parse(res.Body)
	if err != nil {
		return -1, err
	}

	properties, err := htmlutil.ParseMetaProperties(html)
	if err != nil {
		return -1, err
	}

	// The Android deep link for the SR app contains a link such as
	// sesrplay://play/program/4914
	appLink := properties.Get("al:android:url")
	if appLink == "" {
		return -1, ErrProgramIDNotFound
	}

	u, err := url.Parse(appLink)
	if err != nil {
		return -1, fmt.Errorf("invalid app link: %w", err)
	}

	idString := strings.TrimPrefix(u.Path, "/program/")

	id, err := strconv.ParseInt(idString, 10, 32)
	if err != nil {
		return -1, fmt.Errorf("invalid app link: %w", err)
	}

	return int(id), nil
}

func (c *Client) GetEpisodePlaylist(ctx context.Context, episodeID int) ([]PlaylistEntry, error) {
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = "/v2/playlists/getplaylistbyepisodeid"

	query := make(url.Values)
	query.Set("format", "json")
	query.Set("pagination", "false")
	query.Set("id", strconv.FormatInt(int64(episodeID), 10))
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
		Playlist []PlaylistEntry `json:"song"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Playlist, nil
}
