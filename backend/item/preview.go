package item

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	neturl "net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

var (
	errInvalidInstagramURL = errors.New("invalid instagram url")
	errPreviewFetchFailed  = errors.New("unable to fetch instagram preview")
	errWebProfileBlocked   = errors.New("instagram web_profile_info blocked: rate limited or missing browser cookie")
	usernameFromTitleRe    = regexp.MustCompile(`^(.*?) on Instagram:`)
	instagramUsernameRe    = regexp.MustCompile(`^[A-Za-z0-9._]{1,30}$`)
)

const (
	platformInstagram = "instagram"
	platformTiktok    = "tiktok"
	platformYoutube   = "youtube"
	platformUnknown   = "unknown"
)

// youtubeCache stores resolved audio URLs with TTL.
type youtubeCacheEntry struct {
	url       string
	expiresAt time.Time
}

var youtubeCache = struct {
	sync.RWMutex
	entries map[string]youtubeCacheEntry
}{entries: make(map[string]youtubeCacheEntry)}

func detectPlatform(rawURL string) string {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return platformUnknown
	}
	host := strings.ToLower(u.Hostname())
	switch {
	case host == "instagram.com" || host == "www.instagram.com":
		return platformInstagram
	case host == "tiktok.com" || host == "www.tiktok.com":
		return platformTiktok
	case host == "youtube.com" || host == "www.youtube.com" || host == "youtu.be" || host == "music.youtube.com":
		return platformYoutube
	default:
		return platformUnknown
	}
}

func extractYouTubeVideoID(rawURL string) string {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return ""
	}
	if u.Hostname() == "youtu.be" {
		return strings.TrimPrefix(u.Path, "/")
	}
	return u.Query().Get("v")
}

type previewClient struct {
	http          *http.Client
	oembedToken   string
	oembedVersion string
	webCookie     string
}

func newPreviewClient() *previewClient {
	version := strings.TrimSpace(os.Getenv("INSTAGRAM_OEMBED_API_VERSION"))
	if version == "" {
		version = "v21.0"
	}
	jar, _ := cookiejar.New(nil)

	return &previewClient{
		http:          &http.Client{Timeout: 12 * time.Second, Jar: jar},
		oembedToken:   strings.TrimSpace(os.Getenv("INSTAGRAM_OEMBED_ACCESS_TOKEN")),
		oembedVersion: version,
		webCookie:     strings.TrimSpace(os.Getenv("INSTAGRAM_WEB_COOKIE")),
	}
}

func (c *previewClient) Preview(ctx context.Context, rawURL string) (*PreviewResponse, error) {
	platform := detectPlatform(rawURL)

	switch platform {
	case platformInstagram:
		return c.previewInstagram(ctx, rawURL)
	case platformTiktok:
		return c.previewTiktok(ctx, rawURL)
	case platformYoutube:
		return c.previewYoutube(ctx, rawURL)
	default:
		return nil, errInvalidInstagramURL
	}
}

func (c *previewClient) previewInstagram(ctx context.Context, rawURL string) (*PreviewResponse, error) {
	target, ok := parseInstagramURL(rawURL)
	if !ok {
		return nil, errInvalidInstagramURL
	}

	if target.kind == instagramKindProfile {
		preview, err := c.previewFromWebProfileInfo(ctx, target)
		if err == nil {
			return preview, nil
		}

		preview, mobileErr := c.previewFromMobileProfileInfo(ctx, target)
		if mobileErr == nil {
			return preview, nil
		}

		preview, jsonErr := c.previewFromProfilePageJSON(ctx, target)
		if jsonErr == nil {
			return preview, nil
		}

		preview, fallbackErr := c.previewFromOpenGraph(ctx, target.url)
		if fallbackErr == nil {
			if preview.IGUsername == "" {
				preview.IGUsername = target.username
			}
			return preview, nil
		}

		if isInstagramAuthError(err) && isInstagramAuthError(mobileErr) && isInstagramAuthError(jsonErr) {
			return nil, fmt.Errorf("%w: instagram returned 401 to server requests; copy a logged-in instagram.com Cookie header into INSTAGRAM_WEB_COOKIE", errPreviewFetchFailed)
		}

		return nil, fmt.Errorf("%w: web_profile_info failed (%v); mobile_profile_info failed (%v); profile_json failed (%v); open_graph failed (%v)", errPreviewFetchFailed, err, mobileErr, jsonErr, fallbackErr)
	}

	if c.oembedToken != "" {
		preview, err := c.previewFromOEmbed(ctx, target.url)
		if err == nil {
			return preview, nil
		}
	}

	preview, err := c.previewFromOpenGraph(ctx, target.url)
	if err == nil {
		return preview, nil
	}

	preview, err = previewFromMediaEndpoint(target)
	if err == nil {
		return preview, nil
	}

	if c.oembedToken == "" {
		return nil, fmt.Errorf("%w: set INSTAGRAM_OEMBED_ACCESS_TOKEN for reliable Instagram post and reel previews", errPreviewFetchFailed)
	}
	return nil, errPreviewFetchFailed
}

func (c *previewClient) previewTiktok(ctx context.Context, rawURL string) (*PreviewResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, errPreviewFetchFailed
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.tiktok.com/")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: tiktok request failed", errPreviewFetchFailed)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: tiktok returned %d", errPreviewFetchFailed, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: tiktok read body failed", errPreviewFetchFailed)
	}

	htmlStr := string(body)
	meta, _ := extractPreviewMetaFromBytes(body)

	// Try to extract video from og:video
	videoURL := ""
	idx := strings.Index(htmlStr, `property="og:video"`)
	if idx >= 0 {
		contentStart := strings.Index(htmlStr[idx:], `content="`)
		if contentStart >= 0 {
			contentStart += idx + len(`content="`)
			contentEnd := strings.Index(htmlStr[contentStart:], `"`)
			if contentEnd >= 0 {
				videoURL = strings.ReplaceAll(htmlStr[contentStart:contentStart+contentEnd], "&amp;", "&")
			}
		}
	}

	// Extract username from URL path
	username := ""
	if u, err := neturl.Parse(rawURL); err == nil {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) >= 1 && parts[0] != "" {
			username = parts[0]
		}
	}
	if username == "" && meta.username != "" {
		username = meta.username
	}

	// Fallback: use TikTok oEmbed for thumbnail
	thumbnail := meta.image
	if thumbnail == "" {
		thumbnail = c.tiktokOEmbedThumbnail(ctx, rawURL)
	}

	return &PreviewResponse{
		IGURL:      rawURL,
		IGImageURL: thumbnail,
		IGUsername: username,
		VideoURL:   videoURL,
	}, nil
}

func (c *previewClient) tiktokOEmbedThumbnail(ctx context.Context, rawURL string) string {
	endpoint := "https://www.tiktok.com/oembed?url=" + neturl.QueryEscape(rawURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	res, err := c.http.Do(req)
	if err != nil {
		return ""
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return ""
	}

	var payload struct {
		ThumbnailURL string `json:"thumbnail_url"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return ""
	}
	return payload.ThumbnailURL
}

func (c *previewClient) previewYoutube(ctx context.Context, rawURL string) (*PreviewResponse, error) {
	videoID := extractYouTubeVideoID(rawURL)
	if videoID == "" {
		return nil, errInvalidInstagramURL
	}

	// Check cache first
	youtubeCache.RLock()
	entry, ok := youtubeCache.entries[videoID]
	youtubeCache.RUnlock()

	if ok && time.Now().Before(entry.expiresAt) {
		return &PreviewResponse{
			IGURL:      rawURL,
			IGImageURL: fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID),
			AudioURL:   entry.url,
		}, nil
	}

	// Resolve audio URL via yt-dlp
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--get-url",
		"-f", "bestaudio[ext=m4a]/bestaudio",
		"--no-warnings",
		rawURL,
	)
	out, err := cmd.Output()
	if err != nil {
		// Fallback to thumbnail-only preview
		return &PreviewResponse{
			IGURL:      rawURL,
			IGImageURL: fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID),
		}, nil
	}

	audioURL := strings.TrimSpace(string(out))
	if audioURL == "" {
		return &PreviewResponse{
			IGURL:      rawURL,
			IGImageURL: fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID),
		}, nil
	}

	// Cache for 4 hours
	youtubeCache.Lock()
	youtubeCache.entries[videoID] = youtubeCacheEntry{
		url:       audioURL,
		expiresAt: time.Now().Add(4 * time.Hour),
	}
	youtubeCache.Unlock()

	return &PreviewResponse{
		IGURL:      rawURL,
		IGImageURL: fmt.Sprintf("https://i.ytimg.com/vi/%s/maxresdefault.jpg", videoID),
		AudioURL:   audioURL,
	}, nil
}

func (c *previewClient) previewFromWebProfileInfo(ctx context.Context, target instagramTarget) (*PreviewResponse, error) {
	if c.webCookie == "" {
		_ = c.primeInstagramCookies(ctx, target.url)
	}

	endpoint, err := neturl.Parse("https://www.instagram.com/api/v1/users/web_profile_info/")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("username", target.username)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	setInstagramWebHeaders(req, target.url, "application/json")
	if c.webCookie != "" {
		req.Header.Set("Cookie", c.webCookie)
	}
	if csrfToken := c.instagramCSRFToken(); csrfToken != "" {
		req.Header.Set("X-CSRFToken", csrfToken)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusForbidden {
		return nil, errWebProfileBlocked
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: web_profile_info returned %d", errPreviewFetchFailed, res.StatusCode)
	}

	var payload struct {
		Data struct {
			User *struct {
				Username        string `json:"username"`
				FullName        string `json:"full_name"`
				ProfilePicURL   string `json:"profile_pic_url"`
				ProfilePicURLHD string `json:"profile_pic_url_hd"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: web_profile_info json decode failed", errPreviewFetchFailed)
	}
	if payload.Data.User == nil {
		return nil, fmt.Errorf("%w: web_profile_info returned no user", errPreviewFetchFailed)
	}

	image := strings.TrimSpace(payload.Data.User.ProfilePicURLHD)
	if image == "" {
		image = strings.TrimSpace(payload.Data.User.ProfilePicURL)
	}
	if image == "" {
		return nil, fmt.Errorf("%w: web_profile_info returned no profile image", errPreviewFetchFailed)
	}

	username := strings.TrimSpace(payload.Data.User.Username)
	if username == "" {
		username = target.username
	}

	return &PreviewResponse{
		IGURL:      target.url,
		IGImageURL: image,
		IGUsername: username,
	}, nil
}

func (c *previewClient) primeInstagramCookies(ctx context.Context, profileURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return err
	}
	setInstagramNavigationHeaders(req)

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: profile cookie request failed", errPreviewFetchFailed)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("%w: profile cookie request returned %d", errPreviewFetchFailed, res.StatusCode)
	}
	return nil
}

func (c *previewClient) previewFromMobileProfileInfo(ctx context.Context, target instagramTarget) (*PreviewResponse, error) {
	endpoint, err := neturl.Parse("https://i.instagram.com/api/v1/users/web_profile_info/")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("username", target.username)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", "Instagram 337.0.0.35.102 Android (30/11; 420dpi; 1080x1920; samsung; SM-G973F; beyond1; exynos9820; en_US; 602430320)")
	req.Header.Set("X-IG-App-ID", "936619743392459")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: mobile_profile_info request failed", errPreviewFetchFailed)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: mobile_profile_info returned %d", errPreviewFetchFailed, res.StatusCode)
	}

	var payload any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: mobile_profile_info decode failed", errPreviewFetchFailed)
	}

	username, image, ok := findProfilePreview(payload, target.username)
	if !ok || image == "" {
		return nil, fmt.Errorf("%w: mobile_profile_info returned no profile image", errPreviewFetchFailed)
	}
	if username == "" {
		username = target.username
	}

	return &PreviewResponse{
		IGURL:      target.url,
		IGImageURL: image,
		IGUsername: username,
	}, nil
}

func (c *previewClient) previewFromProfilePageJSON(ctx context.Context, target instagramTarget) (*PreviewResponse, error) {
	endpoint, err := neturl.Parse(target.url)
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("__a", "1")
	query.Set("__d", "dis")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	setInstagramWebHeaders(req, target.url, "application/json,text/javascript,*/*;q=0.8")
	if c.webCookie != "" {
		req.Header.Set("Cookie", c.webCookie)
	}
	if csrfToken := c.instagramCSRFToken(); csrfToken != "" {
		req.Header.Set("X-CSRFToken", csrfToken)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: profile_json request failed", errPreviewFetchFailed)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: profile_json returned %d", errPreviewFetchFailed, res.StatusCode)
	}

	var payload any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: profile_json decode failed", errPreviewFetchFailed)
	}

	username, image, ok := findProfilePreview(payload, target.username)
	if !ok || image == "" {
		return nil, fmt.Errorf("%w: profile_json returned no profile image", errPreviewFetchFailed)
	}
	if username == "" {
		username = target.username
	}

	return &PreviewResponse{
		IGURL:      target.url,
		IGImageURL: image,
		IGUsername: username,
	}, nil
}

func findProfilePreview(value any, expectedUsername string) (string, string, bool) {
	switch current := value.(type) {
	case map[string]any:
		username := stringValue(current["username"])
		image := firstStringValue(current["profile_pic_url_hd"], current["profile_pic_url"])
		if image != "" && (username == "" || strings.EqualFold(username, expectedUsername)) {
			return username, image, true
		}
		for _, child := range current {
			if username, image, ok := findProfilePreview(child, expectedUsername); ok {
				return username, image, true
			}
		}
	case []any:
		for _, child := range current {
			if username, image, ok := findProfilePreview(child, expectedUsername); ok {
				return username, image, true
			}
		}
	}

	return "", "", false
}

func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func firstStringValue(values ...any) string {
	for _, value := range values {
		if text := stringValue(value); text != "" {
			return text
		}
	}
	return ""
}

func (c *previewClient) instagramCSRFToken() string {
	if token := cookieValue(c.webCookie, "csrftoken"); token != "" {
		return token
	}
	if c.http == nil || c.http.Jar == nil {
		return ""
	}

	instagramURL, err := neturl.Parse("https://www.instagram.com/")
	if err != nil {
		return ""
	}
	for _, cookie := range c.http.Jar.Cookies(instagramURL) {
		if strings.EqualFold(cookie.Name, "csrftoken") {
			return cookie.Value
		}
	}
	return ""
}

func cookieValue(rawCookie string, name string) string {
	for _, part := range strings.Split(rawCookie, ";") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && strings.EqualFold(key, name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isInstagramAuthError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "returned 401")
}

func setInstagramNavigationHeaders(req *http.Request) {
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="148", "Microsoft Edge";v="148", "Not A(Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("User-Agent", instagramWebUserAgent())
}

func setInstagramWebHeaders(req *http.Request, referer string, accept string) {
	req.Header.Set("Accept", accept)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", referer)
	req.Header.Set("Sec-Ch-Ua", `"Chromium";v="148", "Microsoft Edge";v="148", "Not A(Brand";v="99"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", instagramWebUserAgent())
	req.Header.Set("X-ASBD-ID", "359341")
	req.Header.Set("X-IG-App-ID", "936619743392459")
	req.Header.Set("X-IG-WWW-Claim", "0")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
}

func instagramWebUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36 Edg/148.0.0.0"
}

func isAllowedInstagramImageURL(u *neturl.URL) bool {
	if u == nil || u.Scheme != "https" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	if host == "cdninstagram.com" ||
		strings.HasSuffix(host, ".cdninstagram.com") ||
		host == "fbcdn.net" ||
		strings.HasSuffix(host, ".fbcdn.net") {
		return true
	}

	return (host == "instagram.com" || host == "www.instagram.com") && isInstagramMediaImagePath(u.Path)
}

func isInstagramMediaImagePath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[1] == "" || parts[2] != "media" {
		return false
	}
	switch parts[0] {
	case "p", "reel", "tv":
		return true
	default:
		return false
	}
}

func previewFromMediaEndpoint(target instagramTarget) (*PreviewResponse, error) {
	if target.kind != instagramKindMedia {
		return nil, errInvalidInstagramURL
	}
	return &PreviewResponse{
		IGURL:      target.url,
		IGImageURL: strings.TrimRight(target.url, "/") + "/media/?size=l",
	}, nil
}

func (c *previewClient) previewFromOEmbed(ctx context.Context, igURL string) (*PreviewResponse, error) {
	endpoint, err := neturl.Parse("https://graph.facebook.com/" + strings.Trim(c.oembedVersion, "/") + "/instagram_oembed")
	if err != nil {
		return nil, err
	}
	query := endpoint.Query()
	query.Set("url", igURL)
	query.Set("maxwidth", "640")
	query.Set("fields", "thumbnail_url,author_name,provider_name,provider_url")
	query.Set("access_token", c.oembedToken)
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, errPreviewFetchFailed
	}

	var payload struct {
		ThumbnailURL string `json:"thumbnail_url"`
		AuthorName   string `json:"author_name"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.ThumbnailURL) == "" {
		return nil, errPreviewFetchFailed
	}

	return &PreviewResponse{
		IGURL:      igURL,
		IGImageURL: strings.TrimSpace(payload.ThumbnailURL),
		IGUsername: strings.TrimSpace(payload.AuthorName),
	}, nil
}

func (c *previewClient) previewFromOpenGraph(ctx context.Context, igURL string) (*PreviewResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, igURL, nil)
	if err != nil {
		return nil, err
	}
	setInstagramNavigationHeaders(req)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: open_graph request failed", errPreviewFetchFailed)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: open_graph returned %d", errPreviewFetchFailed, res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: open_graph read body failed", errPreviewFetchFailed)
	}

	htmlStr := string(body)
	username := ""
	if u, err := neturl.Parse(igURL); err == nil {
		_, _, username, _ = normalizeInstagramPath(u.Path)
	}

	// For posts/reels, try the embedded image_versions2 matched against og:image —
	// this returns the uncropped original while ensuring it belongs to the requested post.
	// Also extract video URL for reels.
	target, _ := parseInstagramURL(igURL)
	if target.kind == instagramKindMedia {
		image := extractBestImageFromHTML(htmlStr)
		video := extractVideoFromHTML(htmlStr)
		if image != "" || video != "" {
			return &PreviewResponse{
				IGURL:      igURL,
				IGImageURL: image,
				IGUsername: username,
				VideoURL:   video,
			}, nil
		}
	}

	// Try 1: Standard OpenGraph meta tags
	meta, err := extractPreviewMetaFromBytes(body)
	if err == nil && meta.image != "" {
		return &PreviewResponse{
			IGURL:      igURL,
			IGImageURL: meta.image,
			IGUsername: meta.username,
		}, nil
	}

	// For profiles, try image_versions2 as fallback (some profile pages embed it too)
	if target.kind == instagramKindProfile {
		image := extractImageVersionsFromHTML(htmlStr)
		if image != "" {
			return &PreviewResponse{
				IGURL:      igURL,
				IGImageURL: image,
				IGUsername: username,
			}, nil
		}
	}

	// Try 3: Extract from embedded JSON in HTML (Instagram still embeds profile data in script tags)
	image, jsonUsername := extractProfileFromEmbeddedJSON(body)
	if image != "" {
		if jsonUsername != "" {
			username = jsonUsername
		}
		return &PreviewResponse{
			IGURL:      igURL,
			IGImageURL: image,
			IGUsername: username,
		}, nil
	}

	// Try 3: Direct regex search for profile_pic_url in HTML
	image = extractProfilePicFromHTML(string(body))
	if image != "" {
		return &PreviewResponse{
			IGURL:      igURL,
			IGImageURL: image,
			IGUsername: username,
		}, nil
	}

	return nil, fmt.Errorf("%w: open_graph returned no image", errPreviewFetchFailed)
}

func isInstagramURL(raw string) bool {
	_, ok := normalizeInstagramURL(raw)
	return ok
}

func normalizeInstagramURL(raw string) (string, bool) {
	target, ok := parseInstagramURL(raw)
	if !ok {
		return "", false
	}
	return target.url, true
}

type instagramKind string

const (
	instagramKindMedia   instagramKind = "media"
	instagramKindProfile instagramKind = "profile"
)

type instagramTarget struct {
	url      string
	kind     instagramKind
	username string
}

func parseInstagramURL(raw string) (instagramTarget, bool) {
	u, err := neturl.Parse(raw)
	if err != nil {
		return instagramTarget{}, false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return instagramTarget{}, false
	}
	host := strings.ToLower(u.Hostname())
	if host != "instagram.com" && host != "www.instagram.com" {
		return instagramTarget{}, false
	}
	normalizedPath, kind, username, ok := normalizeInstagramPath(u.Path)
	if !ok {
		return instagramTarget{}, false
	}

	u.Scheme = "https"
	u.Host = "www.instagram.com"
	u.Path = normalizedPath
	u.Fragment = ""
	u.RawQuery = ""
	return instagramTarget{
		url:      u.String(),
		kind:     kind,
		username: username,
	}, true
}

func normalizeInstagramPath(path string) (string, instagramKind, string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		return "", "", "", false
	}

	first, err := neturl.PathUnescape(parts[0])
	if err != nil {
		return "", "", "", false
	}

	if len(parts) >= 2 && parts[1] != "" {
		second, err := neturl.PathUnescape(parts[1])
		if err != nil {
			return "", "", "", false
		}
		switch first {
		case "p", "reel", "tv":
			return "/" + first + "/" + second + "/", instagramKindMedia, "", true
		case "reels":
			return "/reel/" + second + "/", instagramKindMedia, "", true
		}
	}

	switch first {
	case "p", "reel", "tv":
		return "", "", "", false
	case "reels":
		return "", "", "", false
	case "accounts", "direct", "explore", "stories":
		return "", "", "", false
	default:
		if !instagramUsernameRe.MatchString(first) {
			return "", "", "", false
		}
		return "/" + first + "/", instagramKindProfile, first, true
	}
}

type previewMeta struct {
	image    string
	username string
}

func extractPreviewMeta(res *http.Response) (*previewMeta, error) {
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return extractPreviewMetaFromBytes(body)
}

func extractPreviewMetaFromBytes(body []byte) (*previewMeta, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var image string
	var title string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var property, name, content string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "property":
					property = strings.ToLower(attr.Val)
				case "name":
					name = strings.ToLower(attr.Val)
				case "content":
					content = strings.TrimSpace(attr.Val)
				}
			}

			switch {
			case property == "og:image" || property == "og:image:secure_url" || name == "twitter:image":
				if image == "" {
					image = content
				}
			case property == "og:title" || name == "title" || name == "twitter:title":
				if title == "" {
					title = content
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	// Fallback: use regex for minified HTML where the parser sometimes misses meta tags
	if image == "" {
		image = extractMetaByRegex(string(body), "og:image")
	}
	if title == "" {
		title = extractMetaByRegex(string(body), "og:title")
	}

	// Decode HTML entities in image URL (Instagram sometimes encodes & as &amp;)
	image = strings.ReplaceAll(image, "&amp;", "&")

	username := ""
	if match := usernameFromTitleRe.FindStringSubmatch(title); len(match) == 2 {
		username = strings.TrimSpace(match[1])
	}

	return &previewMeta{
		image:    image,
		username: username,
	}, nil
}

var metaRegexCache = map[string]*regexp.Regexp{}

func extractMetaByRegex(html, property string) string {
	re, ok := metaRegexCache[property]
	if !ok {
		// Matches both orderings: property="..." content="..." and content="..." property="..."
		pattern := fmt.Sprintf(`<meta[^>]*(?:property=["']%s["'][^>]*content=["']([^"']+)["']|content=["']([^"']+)["'][^>]*property=["']%s["'])`, regexp.QuoteMeta(property), regexp.QuoteMeta(property))
		re = regexp.MustCompile(pattern)
		metaRegexCache[property] = re
	}
	matches := re.FindStringSubmatch(html)
	if len(matches) >= 3 {
		if matches[1] != "" {
			return matches[1]
		}
		return matches[2]
	}
	return ""
}

// extractProfileFromEmbeddedJSON searches for Instagram's embedded JSON data in HTML script tags.
// Instagram embeds user profile data in <script type="application/json"> tags.
func extractProfileFromEmbeddedJSON(body []byte) (image string, username string) {
	// Try to find any <script> tags with JSON content and search for profile_pic_url
	scriptRe := regexp.MustCompile(`<script[^>]*>([\s\S]*?)</script>`)
	matches := scriptRe.FindAllSubmatch(body, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		scriptContent := match[1]

		// Skip non-JSON scripts quickly
		if !bytes.Contains(scriptContent, []byte("profile_pic_url")) {
			continue
		}

		var payload any
		if err := json.Unmarshal(scriptContent, &payload); err != nil {
			// Try to find JSON in window._sharedData format
			continue
		}

		// Use recursive search to find profile picture and username
		img, uname := findProfilePicInAnyJSON(payload)
		if img != "" {
			return img, uname
		}
	}

	return "", ""
}

func findProfilePicInAnyJSON(value any) (image string, username string) {
	switch current := value.(type) {
	case map[string]any:
		// Check for profile picture and username at this level
		if img, ok := current["profile_pic_url_hd"].(string); ok && img != "" {
			if uname, ok := current["username"].(string); ok {
				return img, uname
			}
			return img, ""
		}
		if img, ok := current["profile_pic_url"].(string); ok && img != "" {
			if uname, ok := current["username"].(string); ok {
				return img, uname
			}
			return img, ""
		}
		if img, ok := current["profile_pic_url"].(string); ok && img != "" {
			return img, ""
		}

		// Recurse into child values
		for _, child := range current {
			if img, uname := findProfilePicInAnyJSON(child); img != "" {
				return img, uname
			}
		}
	case []any:
		for _, child := range current {
			if img, uname := findProfilePicInAnyJSON(child); img != "" {
				return img, uname
			}
		}
	}
	return "", ""
}

// extractOGImageID extracts the image identifier from an Instagram CDN URL.
// e.g. "https://.../587631845_..._n.jpg?..." → "587631845_..."
func extractOGImageID(url string) string {
	parts := strings.Split(url, "/")
	for _, p := range parts {
		if idx := strings.Index(p, "_n.jpg"); idx >= 0 {
			return p[:idx]
		}
		if idx := strings.Index(p, "_n.webp"); idx >= 0 {
			return p[:idx]
		}
	}
	return ""
}

// extractBestImageFromHTML finds the image_versions2 block whose candidate URL matches
// the og:image identifier. This avoids picking a random carousel/related-post image.
func extractBestImageFromHTML(html string) string {
	// Step 1: Find og:image URL to use as the anchor
	idx := strings.Index(html, `property="og:image"`)
	if idx < 0 {
		return ""
	}
	contentStart := strings.Index(html[idx:], `content="`)
	if contentStart < 0 {
		return ""
	}
	contentStart += idx + len(`content="`)
	contentEnd := strings.Index(html[contentStart:], `"`)
	if contentEnd < 0 {
		return ""
	}
	ogURL := strings.ReplaceAll(html[contentStart:contentStart+contentEnd], "&amp;", "&")
	ogID := extractOGImageID(ogURL)
	if ogID == "" {
		return ogURL // fallback to the og:image itself
	}

	// Step 2: Scan all image_versions2 blocks and pick the one whose candidate URL
	// contains the same image ID as the og:image.
	searchStart := 0
	for {
		idx := strings.Index(html[searchStart:], `"image_versions2":`)
		if idx < 0 {
			break
		}
		idx += searchStart

		start := idx + len(`"image_versions2":`)
		depth := 0
		end := start
		for i := start; i < len(html); i++ {
			if html[i] == '{' {
				depth++
			} else if html[i] == '}' {
				depth--
				if depth == 0 {
					end = i + 1
					break
				}
			}
		}
		if end <= start {
			break
		}

		jsonStr := html[start:end]
		jsonStr = strings.ReplaceAll(jsonStr, `\\u0025`, "%")
		jsonStr = strings.ReplaceAll(jsonStr, `\\u0026`, "&")
		jsonStr = strings.ReplaceAll(jsonStr, `\\u002F`, "/")
		jsonStr = strings.ReplaceAll(jsonStr, `\\/`, "/")
		jsonStr = strings.ReplaceAll(jsonStr, `&amp;`, "&")

		var payload struct {
			Candidates []struct {
				URL string `json:"url"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
			searchStart = end
			continue
		}
		if len(payload.Candidates) > 0 {
			url := payload.Candidates[0].URL
			if strings.Contains(url, ogID) {
				return strings.TrimSpace(url)
			}
		}
		searchStart = end
	}

	// Fallback to og:image if no matching image_versions2 found
	return ogURL
}

// extractImageVersionsFromHTML extracts the highest-quality image URL from Instagram's embedded
// image_versions2 JSON inside the HTML. This works for posts/reels and usually returns an uncropped
// original image (e.g. 1440x1800) instead of the square-cropped og:image thumbnail.
func extractImageVersionsFromHTML(html string) string {
	idx := strings.Index(html, `"image_versions2":`)
	if idx < 0 {
		return ""
	}

	start := idx + len(`"image_versions2":`)
	depth := 0
	end := start
	for i := start; i < len(html); i++ {
		if html[i] == '{' {
			depth++
		} else if html[i] == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}
	if end <= start {
		return ""
	}

	jsonStr := html[start:end]
	// Fix common JSON escape sequences found in Instagram's inline JSON
	jsonStr = strings.ReplaceAll(jsonStr, `\\u0025`, "%")
	jsonStr = strings.ReplaceAll(jsonStr, `\\u0026`, "&")
	jsonStr = strings.ReplaceAll(jsonStr, `\\u002F`, "/")
	jsonStr = strings.ReplaceAll(jsonStr, `\\/`, "/")
	jsonStr = strings.ReplaceAll(jsonStr, `&amp;`, "&")

	var payload struct {
		Candidates []struct {
			URL string `json:"url"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		return ""
	}
	if len(payload.Candidates) == 0 {
		return ""
	}
	return strings.TrimSpace(payload.Candidates[0].URL)
}

// extractVideoFromHTML searches for a video URL in Instagram HTML.
// It checks og:video meta tags first, then falls back to video_url in embedded JSON.
func extractVideoFromHTML(html string) string {
	// Try 1: og:video meta tag
	idx := strings.Index(html, `property="og:video"`)
	if idx >= 0 {
		contentStart := strings.Index(html[idx:], `content="`)
		if contentStart >= 0 {
			contentStart += idx + len(`content="`)
			contentEnd := strings.Index(html[contentStart:], `"`)
			if contentEnd >= 0 {
				return strings.ReplaceAll(html[contentStart:contentStart+contentEnd], "&amp;", "&")
			}
		}
	}

	// Try 2: og:video:url (alternate property name)
	idx = strings.Index(html, `property="og:video:url"`)
	if idx >= 0 {
		contentStart := strings.Index(html[idx:], `content="`)
		if contentStart >= 0 {
			contentStart += idx + len(`content="`)
			contentEnd := strings.Index(html[contentStart:], `"`)
			if contentEnd >= 0 {
				return strings.ReplaceAll(html[contentStart:contentStart+contentEnd], "&amp;", "&")
			}
		}
	}

	// Try 3: video_url in embedded JSON
	videoRe := regexp.MustCompile(`"video_url"\s*:\s*"([^"]+)"`)
	if match := videoRe.FindStringSubmatch(html); len(match) > 1 {
		url := match[1]
		url = strings.ReplaceAll(url, `\\u0026`, "&")
		url = strings.ReplaceAll(url, `\u0026`, "&")
		url = strings.ReplaceAll(url, `&amp;`, "&")
		return url
	}

	return ""
}

// extractProfilePicFromHTML searches the raw HTML for profile_pic_url using regex as a last resort.
func extractProfilePicFromHTML(html string) string {
	// Look for profile_pic_url_hd first (higher quality)
	hdRe := regexp.MustCompile(`"profile_pic_url_hd"\s*:\s*"([^"]+)"`)
	if match := hdRe.FindStringSubmatch(html); len(match) > 1 {
		return match[1]
	}

	// Fallback to profile_pic_url
	picRe := regexp.MustCompile(`"profile_pic_url"\s*:\s*"([^"]+)"`)
	if match := picRe.FindStringSubmatch(html); len(match) > 1 {
		return match[1]
	}

	return ""
}
