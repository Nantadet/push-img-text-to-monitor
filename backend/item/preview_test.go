package item

import "testing"

func TestDetectPlatformTikTokHosts(t *testing.T) {
	cases := []string{
		"https://www.tiktok.com/@user/video/123",
		"https://tiktok.com/@user/video/123",
		"https://m.tiktok.com/v/123",
		"https://vm.tiktok.com/ZExample/",
		"https://vt.tiktok.com/ZExample/",
	}

	for _, rawURL := range cases {
		if got := detectPlatform(rawURL); got != platformTiktok {
			t.Fatalf("detectPlatform(%q) = %q, want %q", rawURL, got, platformTiktok)
		}
	}
}

func TestTikTokURLHelpers(t *testing.T) {
	rawURL := "https://www.tiktok.com/@user/video/7345678901234567890"

	if got := tiktokUsernameFromURL(rawURL); got != "@user" {
		t.Fatalf("tiktokUsernameFromURL() = %q, want @user", got)
	}
	if got := tiktokVideoIDFromURL(rawURL); got != "7345678901234567890" {
		t.Fatalf("tiktokVideoIDFromURL() = %q, want 7345678901234567890", got)
	}
	if got := tiktokEmbedURL(rawURL); got != "https://www.tiktok.com/player/v1/7345678901234567890?autoplay=1&muted=1&loop=1&controls=0&progress_bar=0&play_button=0&volume_control=0&fullscreen_button=0&rel=0" {
		t.Fatalf("tiktokEmbedURL() = %q", got)
	}
}

func TestTikTokEmbedURLFromHTML(t *testing.T) {
	rawHTML := `<iframe src="https://www.tiktok.com/embed/v2/7345678901234567890"></iframe>`

	if got := tiktokEmbedURLFromHTML(rawHTML); got != "https://www.tiktok.com/player/v1/7345678901234567890?autoplay=1&muted=1&loop=1&controls=0&progress_bar=0&play_button=0&volume_control=0&fullscreen_button=0&rel=0" {
		t.Fatalf("tiktokEmbedURLFromHTML() = %q", got)
	}
}
