package router

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/luskaner/ageLANServer/common/game"
	i "github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/models"
)

// newMatchContentGame builds a game whose CustomFiles instance is enabled and
// rooted at a per-test temp folder. The router under test is built separately
// via gameHandler.
func newMatchContentGame(t *testing.T, gameId string, opts *models.CustomFilesOpts) models.Game {
	t.Helper()
	origAuth := i.Authentication
	i.Authentication = "disabled"
	t.Cleanup(func() { i.Authentication = origAuth })
	origGen := i.GeneratePlatformUserId
	i.GeneratePlatformUserId = false
	t.Cleanup(func() { i.GeneratePlatformUserId = origGen })
	i.InitializeRng(42)

	fake := &fakeResources{}
	fake.Initialize(gameId, nil)
	if opts == nil {
		opts = &models.CustomFilesOpts{}
	}
	opts.Enabled = true
	opts.Folder = t.TempDir()
	if opts.Expiry <= 0 {
		opts.Expiry = time.Hour
	}
	if opts.MaxFileSize <= 0 {
		opts.MaxFileSize = 1 << 20
	}
	if opts.MaxTotalSize <= 0 {
		opts.MaxTotalSize = 4 << 20
	}
	g := models.CreateMainGame(gameId, &models.CreateMainGameOpts{
		Instances: &models.InstanceOpts{
			Resources:       fake,
			CustomFilesOpts: opts,
		},
	})
	t.Cleanup(g.CustomFiles().Close)
	return g
}

func doRawRequest(handler http.Handler, method, path string, g models.Game, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	ctx := context.WithValue(req.Context(), "game", g)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// buildUploadBody reproduces the multipart capture: fileContent first, then
// matchId.
func buildUploadBody(t *testing.T, filename string, content []byte, matchId string, withFile, withMatch bool) (io.Reader, string) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if withFile {
		fw, err := mw.CreateFormFile("fileContent", filename)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = fw.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if withMatch {
		if err := mw.WriteField("matchId", matchId); err != nil {
			t.Fatal(err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, mw.FormDataContentType()
}

func TestMatchContentUploadDownload_httptest(t *testing.T) {
	for _, gid := range []string{game.AoE1, game.AoE2, game.AoE3, game.AoE4, game.AoM} {
		t.Run(gid, func(t *testing.T) {
			g := newMatchContentGame(t, gid, nil)
			h := gameHandler(gid)
			content := []byte("AOE2 multiplayer savegame bytes")

			body, contentType := buildUploadBody(t, "TEST2.aoe2mpgame", content, "234876462", true, true)
			rr := doRawRequest(h, "POST", "/api/MatchContent/", g, body, map[string]string{"Content-Type": contentType})
			if rr.Code != http.StatusOK {
				t.Fatalf("upload code = %d want 200 body=%q", rr.Code, rr.Body.String())
			}
			if g.CustomFiles().Len() != 1 {
				t.Fatalf("expected one stored file, got %d", g.CustomFiles().Len())
			}

			// Download must work without a session (the game does not send one).
			dl := doRawRequest(h, "GET", "/api/MatchContent/234876462", g, nil, nil)
			if dl.Code != http.StatusOK {
				t.Fatalf("download code = %d want 200 body=%q", dl.Code, dl.Body.String())
			}
			if !bytes.Equal(dl.Body.Bytes(), content) {
				t.Fatalf("downloaded %q want %q", dl.Body.Bytes(), content)
			}
			if got := dl.Header().Get("Content-Length"); got != strconv.Itoa(len(content)) {
				t.Fatalf("Content-Length = %q want %d", got, len(content))
			}
			if got := dl.Header().Get("Content-Type"); got != "application/octet-stream" {
				t.Fatalf("Content-Type = %q", got)
			}
			if dl.Header().Get("Content-MD5") == "" || dl.Header().Get("ETag") == "" {
				t.Fatal("download is missing MD5/ETag headers")
			}

			// Unknown id => 404.
			missing := doRawRequest(h, "GET", "/api/MatchContent/999", g, nil, nil)
			if missing.Code != http.StatusNotFound {
				t.Fatalf("missing id code = %d want 404", missing.Code)
			}
		})
	}
}

func TestMatchContentUploadInvalid_httptest(t *testing.T) {
	g := newMatchContentGame(t, game.AoE2, nil)
	h := gameHandler(game.AoE2)

	cases := []struct {
		name      string
		filename  string
		content   []byte
		matchId   string
		withFile  bool
		withMatch bool
	}{
		{"no_match_id", "TEST2.aoe2mpgame", []byte("data"), "", true, false},
		{"zero_match_id", "TEST2.aoe2mpgame", []byte("data"), "0", true, true},
		{"invalid_match_id", "TEST2.aoe2mpgame", []byte("data"), "abc", true, true},
		{"no_file", "", nil, "55", false, true},
		{"empty_file", "TEST2.aoe2mpgame", nil, "55", true, true},
		{"unknown_extension", "save.zip", []byte("data"), "55", true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, contentType := buildUploadBody(t, tc.filename, tc.content, tc.matchId, tc.withFile, tc.withMatch)
			rr := doRawRequest(h, "POST", "/api/MatchContent/", g, body, map[string]string{"Content-Type": contentType})
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("code = %d want 400 body=%q", rr.Code, rr.Body.String())
			}
			if g.CustomFiles().Len() != 0 {
				t.Fatalf("rejected upload was stored (%d entries)", g.CustomFiles().Len())
			}
		})
	}
}

func TestMatchContentUploadTooLarge_httptest(t *testing.T) {
	g := newMatchContentGame(t, game.AoE2, &models.CustomFilesOpts{MaxFileSize: 16, MaxTotalSize: 1024})
	h := gameHandler(game.AoE2)
	content := bytes.Repeat([]byte("A"), 32)
	body, contentType := buildUploadBody(t, "TEST2.aoe2mpgame", content, "1", true, true)
	rr := doRawRequest(h, "POST", "/api/MatchContent/", g, body, map[string]string{"Content-Type": contentType})
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("code = %d want 413 body=%q", rr.Code, rr.Body.String())
	}
	if g.CustomFiles().Len() != 0 {
		t.Fatal("oversized upload was stored")
	}
}

// TestMatchContentUploadOverGlobalCap_httptest is the regression guard for the
// body-cap exemption: a savegame bigger than the server-wide 1 MiB cap has to
// reach the handler and be judged by CustomFiles.MaxFileSize, and being turned
// away by the global reader must still be reported as 413 rather than 500.
func TestMatchContentUploadOverGlobalCap_httptest(t *testing.T) {
	g := newMatchContentGame(t, game.AoE2, &models.CustomFilesOpts{
		MaxFileSize:  8 << 20,
		MaxTotalSize: 16 << 20,
	})
	h := MaxBodySizeMiddleware(1<<20, gameHandler(game.AoE2))

	// Bigger than the 1 MiB global cap, smaller than MaxFileSize: accepted.
	content := bytes.Repeat([]byte{0x11}, 3<<20)
	body, contentType := buildUploadBody(t, "TEST2.aoe2mpgame", content, "9", true, true)
	rr := doRawRequest(h, "POST", "/api/MatchContent/", g, body, map[string]string{"Content-Type": contentType})
	if rr.Code != http.StatusOK {
		t.Fatalf("code = %d want 200 body=%q", rr.Code, rr.Body.String())
	}
	dl := doRawRequest(h, "GET", "/api/MatchContent/9", g, nil, nil)
	if dl.Code != http.StatusOK || !bytes.Equal(dl.Body.Bytes(), content) {
		t.Fatalf("download code=%d len=%d want %d", dl.Code, dl.Body.Len(), len(content))
	}

	// Bigger than MaxFileSize: rejected, and as 413 whatever tripped first.
	tooBig := bytes.Repeat([]byte{0x22}, 9<<20)
	body2, contentType2 := buildUploadBody(t, "TEST2.aoe2mpgame", tooBig, "10", true, true)
	rr2 := doRawRequest(h, "POST", "/api/MatchContent/", g, body2, map[string]string{"Content-Type": contentType2})
	if rr2.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("code = %d want 413 body=%q", rr2.Code, rr2.Body.String())
	}
	if _, _, ok := g.CustomFiles().Open(10); ok {
		t.Fatal("rejected oversized upload was stored")
	}
}

func TestMatchContentOverwrite_httptest(t *testing.T) {
	g := newMatchContentGame(t, game.AoE2, nil)
	h := gameHandler(game.AoE2)
	first := []byte("first version")
	second := []byte("second version longer")
	for _, payload := range [][]byte{first, second} {
		body, contentType := buildUploadBody(t, "TEST2.aoe2mpgame", payload, "42", true, true)
		rr := doRawRequest(h, "POST", "/api/MatchContent/", g, body, map[string]string{"Content-Type": contentType})
		if rr.Code != http.StatusOK {
			t.Fatalf("upload code = %d", rr.Code)
		}
	}
	if g.CustomFiles().Len() != 1 {
		t.Fatalf("len = %d want 1", g.CustomFiles().Len())
	}
	dl := doRawRequest(h, "GET", "/api/MatchContent/42", g, nil, nil)
	if !bytes.Equal(dl.Body.Bytes(), second) {
		t.Fatalf("content = %q want %q", dl.Body.Bytes(), second)
	}
}

// TestMatchContentApiAgeOfEmpiresRoute checks the routes are registered on the
// api.ageofempires.com mux (so they win over the reverse proxy) and that the
// handler is reached.
func TestMatchContentApiAgeOfEmpiresRoute_httptest(t *testing.T) {
	g := newMatchContentGame(t, game.AoE2, nil)
	h := NewApiAgeOfEmpires().InitializeRoutes(game.AoE2, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("request for %s reached the upstream fallback", r.URL.Path)
		w.WriteHeader(http.StatusBadGateway)
	}))

	body, contentType := buildUploadBody(t, "TEST2.aoe2mpgame", []byte("data"), "77", true, true)
	req := httptest.NewRequest("POST", "/api/MatchContent/", body)
	req.Header.Set("Content-Type", contentType)
	req = req.WithContext(context.WithValue(req.Context(), "game", g))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code == http.StatusNotFound || rr.Code == http.StatusBadGateway {
		t.Fatalf("route not handled locally: code=%d", rr.Code)
	}

	dl := httptest.NewRequest("GET", "/api/MatchContent/77", nil)
	dl = dl.WithContext(context.WithValue(dl.Context(), "game", g))
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, dl)
	if rr2.Code != http.StatusOK {
		t.Fatalf("download via api mux code = %d", rr2.Code)
	}
}
