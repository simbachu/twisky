package bluesky

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	labelers      []string
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

// SetLabelers configures DIDs sent via the atproto-accept-labelers header on requests.
func (c *Client) SetLabelers(labelers []string) {
	c.labelers = append([]string(nil), labelers...)
}

type StrongRef struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

type Profile struct {
	DID                string              `json:"did"`
	Handle             string              `json:"handle"`
	DisplayName        string              `json:"displayName"`
	Description        string              `json:"description"`
	DescriptionFacets  []Facet             `json:"descriptionFacets,omitempty"`
	Avatar             string              `json:"avatar"`
	Banner             string              `json:"banner"`
	Followers          int                 `json:"followersCount"`
	Following          int                 `json:"followsCount"`
	Posts              int                 `json:"postsCount"`
	PinnedPost         *StrongRef          `json:"pinnedPost,omitempty"`
	Labels             []Label             `json:"labels,omitempty"`
	Associated         *ProfileAssociated  `json:"associated,omitempty"`
}

type ProfileAssociated struct {
	Labeler bool `json:"labeler,omitempty"`
}

type Label struct {
	Val string `json:"val"`
	Src string `json:"src"`
	URI string `json:"uri,omitempty"`
}

type PostViewer struct {
	Like   string `json:"like,omitempty"`
	Repost string `json:"repost,omitempty"`
}

type Post struct {
	URI         string      `json:"uri"`
	CID         string      `json:"cid,omitempty"`
	Author      Author      `json:"author"`
	Record      PostRecord  `json:"record"`
	Embed       *Embed      `json:"embed,omitempty"`
	LikeCount   int         `json:"likeCount,omitempty"`
	RepostCount int         `json:"repostCount,omitempty"`
	ReplyCount  int         `json:"replyCount,omitempty"`
	Labels      []Label     `json:"labels,omitempty"`
	Viewer      *PostViewer `json:"viewer,omitempty"`
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
		labels = append(labels, Label{Val: value, Src: p.Author.DID, URI: p.URI})
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

type TimelineRequest struct {
	Limit  int
	Cursor string
}

type FeedRequest struct {
	URI    string
	Limit  int
	Cursor string
}

type SavedFeed struct {
	ID     string `json:"id"`
	Pinned bool   `json:"pinned"`
	Type   string `json:"type"`
	URI    string `json:"value"`
}

type FeedGenerator struct {
	URI         string `json:"uri"`
	DisplayName string `json:"displayName"`
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

func (c *Client) setLabelerHeaders(req *http.Request) {
	if len(c.labelers) == 0 {
		return
	}
	req.Header.Set("atproto-accept-labelers", strings.Join(c.labelers, ", "))
}

func (c *Client) doGet(ctx context.Context, endpointURL string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return err
	}
	c.setLabelerHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("bluesky api: %s", apiErr.Message)
		}
		return fmt.Errorf("bluesky api: status %d", resp.StatusCode)
	}

	if dest == nil {
		return nil
	}
	return json.Unmarshal(body, dest)
}

func (c *Client) GetProfile(ctx context.Context, actor string) (*Profile, error) {
	endpoint, err := url.Parse(c.baseURL + "/app.bsky.actor.getProfile")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("actor", actor)
	endpoint.RawQuery = query.Encode()

	var profile Profile
	if err := c.doGet(ctx, endpoint.String(), &profile); err != nil {
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

	var profilesResp getProfilesResponse
	if err := c.doGet(ctx, endpoint.String(), &profilesResp); err != nil {
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

	var feedResp authorFeedResponse
	if err := c.doGet(ctx, endpoint.String(), &feedResp); err != nil {
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

	var searchResp searchPostsResponse
	if err := c.doGet(ctx, endpoint.String(), &searchResp); err != nil {
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

	var postsResp getPostsResponse
	if err := c.doGet(ctx, endpoint.String(), &postsResp); err != nil {
		return nil, err
	}
	return postsResp.Posts, nil
}
