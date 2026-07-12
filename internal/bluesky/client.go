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
	DID                string  `json:"did"`
	Handle             string  `json:"handle"`
	DisplayName        string  `json:"displayName"`
	Description        string  `json:"description"`
	DescriptionFacets  []Facet `json:"descriptionFacets,omitempty"`
	Avatar             string  `json:"avatar"`
	Followers          int     `json:"followersCount"`
	Following          int     `json:"followsCount"`
	Posts              int     `json:"postsCount"`
}

type Label struct {
	Val string `json:"val"`
	Src string `json:"src"`
	URI string `json:"uri,omitempty"`
}

type Post struct {
	URI         string     `json:"uri"`
	Author      Author     `json:"author"`
	Record      PostRecord `json:"record"`
	Embed       *Embed     `json:"embed,omitempty"`
	LikeCount   int        `json:"likeCount,omitempty"`
	RepostCount int        `json:"repostCount,omitempty"`
	ReplyCount  int        `json:"replyCount,omitempty"`
	Labels      []Label    `json:"labels,omitempty"`
}

type Author struct {
	DID         string  `json:"did"`
	Handle      string  `json:"handle"`
	DisplayName string  `json:"displayName"`
	Avatar      string  `json:"avatar"`
	Labels      []Label `json:"labels,omitempty"`
}

type PostRecord struct {
	Text       string          `json:"text"`
	CreatedAt  time.Time       `json:"createdAt"`
	Facets     []Facet         `json:"facets,omitempty"`
	Reply      *RecordReplyRef `json:"reply,omitempty"`
	SelfLabels *SelfLabels     `json:"labels,omitempty"`
}

type SelfLabels struct {
	Type   string          `json:"$type"`
	Values []SelfLabelValue `json:"values"`
}

type SelfLabelValue struct {
	Val string `json:"val"`
}

// AllLabels returns post view labels plus self-labels from the record.
func (p Post) AllLabels() []Label {
	labels := append([]Label{}, p.Labels...)
	if p.Author.DID == "" {
		return labels
	}
	for _, value := range p.Record.SelfLabelValues() {
		labels = append(labels, Label{Val: value, Src: p.Author.DID})
	}
	return labels
}

func (r PostRecord) SelfLabelValues() []string {
	if r.SelfLabels == nil {
		return nil
	}
	values := make([]string, 0, len(r.SelfLabels.Values))
	for _, value := range r.SelfLabels.Values {
		if value.Val != "" {
			values = append(values, value.Val)
		}
	}
	return values
}

type RecordReplyRef struct {
	Root   StrongRef `json:"root"`
	Parent StrongRef `json:"parent"`
}

type StrongRef struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

// ReplyParentURI returns the parent post URI when this post is a reply.
func (p Post) ReplyParentURI() string {
	if p.Record.Reply == nil {
		return ""
	}
	return p.Record.Reply.Parent.URI
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

type getPostsResponse struct {
	Posts []Post `json:"posts"`
}

const maxGetPostsURIs = 25

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

func (c *Client) GetPosts(ctx context.Context, uris []string) ([]Post, error) {
	if len(uris) == 0 {
		return nil, nil
	}

	posts := make([]Post, 0, len(uris))
	for start := 0; start < len(uris); start += maxGetPostsURIs {
		end := start + maxGetPostsURIs
		if end > len(uris) {
			end = len(uris)
		}
		chunk, err := c.getPosts(ctx, uris[start:end])
		if err != nil {
			return nil, err
		}
		posts = append(posts, chunk...)
	}
	return posts, nil
}

func (c *Client) getPosts(ctx context.Context, uris []string) ([]Post, error) {
	endpoint, err := url.Parse(c.baseURL + "/app.bsky.feed.getPosts")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	for _, uri := range uris {
		query.Add("uris", uri)
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

	var postsResp getPostsResponse
	if err := json.Unmarshal(body, &postsResp); err != nil {
		return nil, err
	}
	return postsResp.Posts, nil
}
