package bluesky

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const publicAppView = "https://public.api.bsky.app/xrpc"

// searchPosts is blocked on the public AppView (403); use the authenticated AppView host.
const searchAppView = "https://api.bsky.app/xrpc"

const (
	FilterPostsNoReplies = "posts_no_replies"
	FilterPostsWithMedia = "posts_with_media"
)

type Client struct {
	baseURL       string
	searchBaseURL string
	httpClient    *http.Client
}

func NewClient() *Client {
	return NewClientWith(publicAppView, nil)
}

func NewClientWith(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	searchBaseURL := searchAppView
	if baseURL != publicAppView {
		searchBaseURL = baseURL
	}
	return &Client{
		baseURL:       baseURL,
		searchBaseURL: searchBaseURL,
		httpClient:    httpClient,
	}
}

type Profile struct {
	DID          string `json:"did"`
	Handle       string `json:"handle"`
	DisplayName  string `json:"displayName"`
	Description  string `json:"description"`
	Avatar       string `json:"avatar"`
	Followers    int    `json:"followersCount"`
	Following    int    `json:"followsCount"`
	Posts        int    `json:"postsCount"`
}

type FeedItem struct {
	Post Post `json:"post"`
}

type Post struct {
	URI    string     `json:"uri"`
	Author Author     `json:"author"`
	Record PostRecord `json:"record"`
	Embed  *Embed     `json:"embed,omitempty"`
}

type Author struct {
	Handle      string `json:"handle"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
}

type PostRecord struct {
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	Facets    []Facet   `json:"facets,omitempty"`
}

type Facet struct {
	Index    FacetIndex     `json:"index"`
	Features []FacetFeature `json:"features"`
}

type FacetIndex struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

type FacetFeature struct {
	Type string `json:"$type"`
	Tag  string `json:"tag,omitempty"`
	DID  string `json:"did,omitempty"`
	URI  string `json:"uri,omitempty"`
}

type Embed struct {
	Type   string       `json:"$type"`
	Images []EmbedImage `json:"images,omitempty"`
	Items  []EmbedImage `json:"items,omitempty"`
}

func (e *Embed) MediaImages() []EmbedImage {
	if e == nil {
		return nil
	}
	if len(e.Images) > 0 {
		return e.Images
	}
	return e.Items
}

type AspectRatio struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type EmbedImage struct {
	Thumb       string       `json:"thumb"`
	Thumbnail   string       `json:"thumbnail,omitempty"`
	Fullsize    string       `json:"fullsize"`
	Alt         string       `json:"alt"`
	AspectRatio *AspectRatio `json:"aspectRatio,omitempty"`
}

func (i EmbedImage) ThumbURL() string {
	if i.Thumb != "" {
		return i.Thumb
	}
	return i.Thumbnail
}

type AuthorFeedRequest struct {
	Actor  string
	Filter string
	Limit  int
	Cursor string
}

type AuthorFeedResponse struct {
	Feed   []FeedItem
	Cursor string
}

type authorFeedResponse struct {
	Feed   []FeedItem `json:"feed"`
	Cursor string     `json:"cursor,omitempty"`
}

type SearchPostsRequest struct {
	Tag    string
	Limit  int
	Cursor string
}

type SearchPostsResponse struct {
	Posts  []Post
	Cursor string
}

type searchPostsResponse struct {
	Posts  []Post `json:"posts"`
	Cursor string `json:"cursor,omitempty"`
}

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type getProfilesResponse struct {
	Profiles []Profile `json:"profiles"`
}

func (c *Client) GetProfile(ctx context.Context, actor string) (*Profile, error) {
	endpoint, err := url.Parse(c.baseURL + "/app.bsky.actor.getProfile")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("actor", actor)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("bluesky api: %s", apiErr.Message)
		}
		return nil, fmt.Errorf("bluesky api: status %d", resp.StatusCode)
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (c *Client) GetProfiles(ctx context.Context, actors []string) ([]Profile, error) {
	endpoint, err := url.Parse(c.baseURL + "/app.bsky.actor.getProfiles")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	for _, actor := range actors {
		query.Add("actors", actor)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("bluesky api: %s", apiErr.Message)
		}
		return nil, fmt.Errorf("bluesky api: status %d", resp.StatusCode)
	}

	var profilesResp getProfilesResponse
	if err := json.Unmarshal(body, &profilesResp); err != nil {
		return nil, err
	}
	return profilesResp.Profiles, nil
}

func (c *Client) GetAuthorFeed(ctx context.Context, feedReq AuthorFeedRequest) (*AuthorFeedResponse, error) {
	endpoint, err := url.Parse(c.baseURL + "/app.bsky.feed.getAuthorFeed")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("actor", feedReq.Actor)
	if feedReq.Filter != "" {
		query.Set("filter", feedReq.Filter)
	}
	if feedReq.Limit > 0 {
		query.Set("limit", strconv.Itoa(feedReq.Limit))
	}
	if feedReq.Cursor != "" {
		query.Set("cursor", feedReq.Cursor)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("bluesky api: %s", apiErr.Message)
		}
		return nil, fmt.Errorf("bluesky api: status %d", resp.StatusCode)
	}

	var feedResp authorFeedResponse
	if err := json.Unmarshal(body, &feedResp); err != nil {
		return nil, err
	}
	return &AuthorFeedResponse{
		Feed:   feedResp.Feed,
		Cursor: feedResp.Cursor,
	}, nil
}

func (c *Client) SearchPosts(ctx context.Context, searchReq SearchPostsRequest) (*SearchPostsResponse, error) {
	endpoint, err := url.Parse(c.searchBaseURL + "/app.bsky.feed.searchPosts")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("q", "#"+searchReq.Tag)
	query.Set("sort", "latest")
	if searchReq.Limit > 0 {
		query.Set("limit", strconv.Itoa(searchReq.Limit))
	}
	if searchReq.Cursor != "" {
		query.Set("cursor", searchReq.Cursor)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return nil, fmt.Errorf("bluesky api: %s", apiErr.Message)
		}
		return nil, fmt.Errorf("bluesky api: status %d", resp.StatusCode)
	}

	var searchResp searchPostsResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, err
	}
	return &SearchPostsResponse{
		Posts:  searchResp.Posts,
		Cursor: searchResp.Cursor,
	}, nil
}
