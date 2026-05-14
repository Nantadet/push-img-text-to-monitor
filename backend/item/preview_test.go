package item

import (
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"testing"
)

func TestNormalizeInstagramURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{
			name: "post",
			raw:  "https://www.instagram.com/p/abc123/?igsh=xyz",
			want: "https://www.instagram.com/p/abc123/",
			ok:   true,
		},
		{
			name: "reel",
			raw:  "http://instagram.com/reel/abc123/",
			want: "https://www.instagram.com/reel/abc123/",
			ok:   true,
		},
		{
			name: "reels alias",
			raw:  "https://instagram.com/reels/abc123/?utm_source=ig_web_copy_link",
			want: "https://www.instagram.com/reel/abc123/",
			ok:   true,
		},
		{
			name: "tv",
			raw:  "https://www.instagram.com/tv/abc123/",
			want: "https://www.instagram.com/tv/abc123/",
			ok:   true,
		},
		{
			name: "profile",
			raw:  "https://www.instagram.com/example/",
			want: "https://www.instagram.com/example/",
			ok:   true,
		},
		{
			name: "profile subpage",
			raw:  "https://www.instagram.com/example/tagged/",
			want: "https://www.instagram.com/example/",
			ok:   true,
		},
		{
			name: "missing shortcode",
			raw:  "https://www.instagram.com/p/",
			ok:   false,
		},
		{
			name: "other host",
			raw:  "https://example.com/p/abc123/",
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeInstagramURL(tt.raw)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("url = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseInstagramProfileURL(t *testing.T) {
	target, ok := parseInstagramURL("https://www.instagram.com/ntd.wcy/")
	if !ok {
		t.Fatal("expected profile URL to parse")
	}
	if target.kind != instagramKindProfile {
		t.Fatalf("kind = %q, want %q", target.kind, instagramKindProfile)
	}
	if target.username != "ntd.wcy" {
		t.Fatalf("username = %q", target.username)
	}
}

func TestExtractPreviewMeta(t *testing.T) {
	res := &http.Response{
		Body: io.NopCloser(strings.NewReader(`
			<html>
				<head>
					<meta property="og:image" content="https://cdn.example/image.jpg">
					<meta property="og:title" content="demo_user on Instagram: test post">
				</head>
			</html>
		`)),
	}

	meta, err := extractPreviewMeta(res)
	if err != nil {
		t.Fatalf("extractPreviewMeta returned error: %v", err)
	}
	if meta.image != "https://cdn.example/image.jpg" {
		t.Fatalf("image = %q", meta.image)
	}
	if meta.username != "demo_user" {
		t.Fatalf("username = %q", meta.username)
	}
}

func TestFindProfilePreview(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"user": map[string]any{
				"username":           "ntd.wcy",
				"profile_pic_url_hd": "https://scontent.example/profile.jpg",
			},
		},
	}

	username, image, ok := findProfilePreview(payload, "ntd.wcy")
	if !ok {
		t.Fatal("expected profile preview to be found")
	}
	if username != "ntd.wcy" {
		t.Fatalf("username = %q", username)
	}
	if image != "https://scontent.example/profile.jpg" {
		t.Fatalf("image = %q", image)
	}
}

func TestIsAllowedInstagramImageURL(t *testing.T) {
	tests := []struct {
		raw string
		ok  bool
	}{
		{raw: "https://scontent-bkk1-1.cdninstagram.com/v/example.jpg", ok: true},
		{raw: "https://example.fbcdn.net/v/example.jpg", ok: true},
		{raw: "https://www.instagram.com/p/abc123/media/?size=l", ok: true},
		{raw: "https://www.instagram.com/reel/abc123/media/?size=l", ok: true},
		{raw: "http://scontent-bkk1-1.cdninstagram.com/v/example.jpg", ok: false},
		{raw: "https://cdninstagram.com.evil.test/v/example.jpg", ok: false},
		{raw: "https://www.instagram.com/accounts/login/", ok: false},
		{raw: "https://example.com/v/example.jpg", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			u, err := neturl.Parse(tt.raw)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if got := isAllowedInstagramImageURL(u); got != tt.ok {
				t.Fatalf("isAllowedInstagramImageURL = %v, want %v", got, tt.ok)
			}
		})
	}
}

func TestExtractPreviewMetaFromBytes(t *testing.T) {
	html := []byte(`<html><head>
		<meta property="og:image" content="https://cdn.example/image.jpg">
		<meta property="og:title" content="demo_user on Instagram: test post">
	</head></html>`)

	meta, err := extractPreviewMetaFromBytes(html)
	if err != nil {
		t.Fatalf("extractPreviewMetaFromBytes returned error: %v", err)
	}
	if meta.image != "https://cdn.example/image.jpg" {
		t.Fatalf("image = %q", meta.image)
	}
	if meta.username != "demo_user" {
		t.Fatalf("username = %q", meta.username)
	}
}

func TestExtractProfileFromEmbeddedJSON(t *testing.T) {
	html := []byte(`<html><head><script type="application/json">
		{"data":{"user":{"username":"testuser","profile_pic_url_hd":"https://cdn.example/hd.jpg"}}}
	</script></head></html>`)

	image, username := extractProfileFromEmbeddedJSON(html)
	if image != "https://cdn.example/hd.jpg" {
		t.Fatalf("image = %q", image)
	}
	if username != "testuser" {
		t.Fatalf("username = %q", username)
	}
}

func TestExtractProfilePicFromHTML(t *testing.T) {
	html := `<html><body>"profile_pic_url_hd":"https://cdn.example/hd.jpg"</body></html>`

	image := extractProfilePicFromHTML(html)
	if image != "https://cdn.example/hd.jpg" {
		t.Fatalf("image = %q", image)
	}

	htmlNoHD := `<html><body>"profile_pic_url":"https://cdn.example/normal.jpg"</body></html>`
	image = extractProfilePicFromHTML(htmlNoHD)
	if image != "https://cdn.example/normal.jpg" {
		t.Fatalf("image = %q", image)
	}
}
