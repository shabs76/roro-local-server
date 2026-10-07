package controlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/constants"
)

func TestStoreUniqueNeverReplacesAnExistingFile(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "photo_1.jpg")
	if err := os.WriteFile(existing, []byte("first tablet"), 0o644); err != nil {
		t.Fatal(err)
	}
	tmp := filepath.Join(dir, ".upload-1")
	if err := os.WriteFile(tmp, []byte("second tablet"), 0o644); err != nil {
		t.Fatal(err)
	}

	name, err := storeUnique(tmp, dir, "photo_1", ".jpg")
	if err != nil {
		t.Fatal(err)
	}
	if name != "photo_1_2.jpg" {
		t.Fatalf("stored as %q, want photo_1_2.jpg", name)
	}
	if got, _ := os.ReadFile(existing); string(got) != "first tablet" {
		t.Fatalf("existing file was changed to %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, name)); string(got) != "second tablet" {
		t.Fatalf("new file holds %q", got)
	}
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatal("temporary file left behind")
	}
}

// Two tablets send different photos under the same file name within one second.
// Both photos must be stored, each under its own path.
func TestUploadsWithTheSameNameKeepBothPhotos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	wd, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(wd)

	router := gin.New()
	router.POST("/upload", UploadMedia)

	upload := func(content string) string {
		t.Helper()
		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		part, _ := w.CreateFormFile("file", "Kh8p_20260528_123722.jpg")
		part.Write([]byte(content))
		w.WriteField("name", "Kh8p_20260528_123722jpg")
		w.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload", &body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("upload answered %d: %s", rec.Code, rec.Body.String())
		}
		var res struct{ Data string }
		json.Unmarshal(rec.Body.Bytes(), &res)
		return res.Data
	}

	urlA := upload("photo taken on tablet A")
	urlB := upload("photo taken on tablet B")
	if urlA == urlB {
		t.Fatalf("both photos stored at %s", urlA)
	}
	for url, want := range map[string]string{urlA: "photo taken on tablet A", urlB: "photo taken on tablet B"} {
		got, err := os.ReadFile(filepath.Join(constants.MediaBaseDir, url))
		if err != nil || string(got) != want {
			t.Fatalf("%s holds %q (%v), want %q", url, got, err, want)
		}
	}
}
