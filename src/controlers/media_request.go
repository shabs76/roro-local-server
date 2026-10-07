package controlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/constants"
	manifestdataservices "github.com/shabs76/roro-local-server/database/manifest_data_services"
)

// ensureDir creates the directory if it doesn't exist
func ensureDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// getMediaType determines the sub-folder based on file extension
func getMediaType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "images"
	case ".mp4", ".avi", ".mov", ".mkv", ".webm":
		return "videos"
	case ".mp3", ".wav", ".ogg":
		return "audio"
	case ".pdf", ".doc", ".docx", ".txt":
		return "documents"
	default:
		return "others"
	}
}

// UploadMedia handles file upload.
// Expects multipart/form-data:
// - file: The file to upload
// - name: (Optional) Rename the file
func UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		slog.Error("File is missing or could not get it from the upload.")
		slog.Error(err.Error())
		if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
			c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "No file uploaded"})
			return
		}
		// The body stopped arriving (slow Wi-Fi, connection dropped, read deadline).
		// A retry can succeed, so this is not reported as a bad request.
		c.JSON(http.StatusRequestTimeout, gin.H{"state": constants.ErrorState, "data": "The upload was interrupted. Please try again."})
		return
	}

	// Use provided name or original filename
	name := c.PostForm("name")
	if name == "" {
		name = file.Filename
	} else {
		// If name is provided but has no extension, append the original extension
		if filepath.Ext(name) == "" {
			name = name + filepath.Ext(file.Filename)
		}
	}

	// Sanitize filename
	name = filepath.Base(name)

	ext := filepath.Ext(name)
	nameWithoutExt := strings.ReplaceAll(strings.TrimSuffix(name, ext), " ", "_")

	mediaType := getMediaType(file.Filename)
	saveDir := filepath.Join(constants.MediaBaseDir, mediaType)

	if err := ensureDir(saveDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to create directory"})
		return
	}

	if file.Size == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"state": constants.ErrorState, "data": "The uploaded file is empty"})
		return
	}

	// Copy into a temporary file next to the destination while hashing it.
	src, err := file.Open()
	if err != nil {
		slog.Error("Failed to open uploaded file", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "The uploaded file could not be read"})
		return
	}
	defer src.Close()

	tmp, err := os.CreateTemp(saveDir, ".upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to save file"})
		return
	}
	hash := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(tmp, hash), src)
	closeErr := tmp.Close()
	if copyErr != nil || closeErr != nil {
		os.Remove(tmp.Name())
		slog.Error("Failed to store uploaded file", "copyError", copyErr, "closeError", closeErr)
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to save file"})
		return
	}
	sum := hex.EncodeToString(hash.Sum(nil))

	// The same content was stored before (older app builds upload a photo again on
	// every retry): answer with that file instead of keeping another copy.
	if url, ok := manifestdataservices.FindMediaFile(sum); ok {
		existing := filepath.Join(constants.MediaBaseDir, url)
		if info, err := os.Stat(existing); err == nil && !info.IsDir() {
			os.Remove(tmp.Name())
			respondUploaded(c, url, filepath.Dir(url), filepath.Base(url), existing)
			return
		}
	}

	// Different tablets can send different photos under the same name in the same
	// second. The stored name therefore carries part of the content hash, and the file
	// is placed without ever replacing an existing one.
	name, err = storeUnique(tmp.Name(), saveDir,
		fmt.Sprintf("%s_%d_%s", nameWithoutExt, time.Now().Unix(), sum[:12]), ext)
	if err != nil {
		os.Remove(tmp.Name())
		slog.Error("Failed to store uploaded file", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to save file"})
		return
	}
	dst := filepath.Join(saveDir, name)
	url := fmt.Sprintf("%s/%s", mediaType, name)
	manifestdataservices.SaveMediaFile(sum, url, file.Size)
	respondUploaded(c, url, mediaType, name, dst)
}

// storeUnique moves the temporary file into dir as base+ext, or base_2+ext,
// base_3+ext ... when that name is taken. os.Link fails when the target exists, so a
// file stored by another upload is never replaced, even by an upload running at the
// same moment. It returns the stored file name.
func storeUnique(tmpPath, dir, base, ext string) (string, error) {
	for i := 1; i <= 100; i++ {
		name := base + ext
		if i > 1 {
			name = fmt.Sprintf("%s_%d%s", base, i, ext)
		}
		dst := filepath.Join(dir, name)
		err := os.Link(tmpPath, dst)
		if err == nil {
			os.Remove(tmpPath)
			return name, nil
		}
		if errors.Is(err, os.ErrExist) {
			continue
		}
		// The file system cannot hard-link: fall back to a rename, still only onto a
		// name that is free.
		if _, statErr := os.Stat(dst); errors.Is(statErr, os.ErrNotExist) {
			if err := os.Rename(tmpPath, dst); err != nil {
				return "", err
			}
			return name, nil
		}
	}
	return "", fmt.Errorf("no free file name for %s%s", base, ext)
}

func respondUploaded(c *gin.Context, url, mediaType, filename, path string) {
	c.JSON(http.StatusOK, gin.H{
		"state":   constants.SuccessState,
		"message": "File uploaded successfully",
		"info": gin.H{
			"url":      url,
			"type":     mediaType,
			"filename": filename,
			"path":     path,
		},
		"data": url,
	})
}

// GetFile streams a file by its relative path or media type/filename key.
// Expects 'url' as a query parameter (e.g. ?url=images/my_image_12345.png or ?url=videos/movie.mp4)
func GetFile(c *gin.Context) {
	// The url parameter is expected to be relative to MediaBaseDir, e.g., "images/photo.jpg" or "videos/demo.mp4"
	// It matches the output format of UploadMedia: /media/{type}/{name}, but we probably just need {type}/{name}
	// Let's assume the client sends the relative path stored in the DB or returned by upload.
	// If the client sends "/media/images/file.jpg", we need to strip "/media/".
	// If the client sends just "images/file.jpg", we use it directly.

	rawPath := c.Query("url")
	if rawPath == "" {
		rawPath = c.Param("url") // fallback if mapped as path param
	}

	if rawPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "File URL is required"})
		return
	}

	// Clean up path found in URL
	// Ensure no directory traversal hacking
	relativePath := filepath.Clean(rawPath)
	if strings.Contains(relativePath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid file path"})
		return
	}

	fullPath := filepath.Join(constants.MediaBaseDir, relativePath)

	fileInfo, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"state": constants.ErrorState, "data": "File not found"})
		return
	}

	// Open file
	file, err := os.Open(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to open file"})
		return
	}
	defer file.Close()

	// Determine mime type
	// We can use http.DetectContentType for first 512 bytes, or rely on extension.
	// Gin's c.File() does this automatically, but since the user wants specific headers and byte streaming,
	// we will set them manually.
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to read file preview"})
		return
	}
	contentType := http.DetectContentType(buffer)
	// Reset file pointer
	file.Seek(0, 0)

	filename := filepath.Base(fullPath)

	c.Header("Access-Control-Expose-Headers", "Content-Disposition, File-Name")
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("File-Name", filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Stream the file
	http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), file)
}

// PlayMedia streams a file specifically for video/audio players (like Flutter video_player).
// It supports HTTP byte-range requests and serves the file inline instead of as an attachment.
func PlayMedia(c *gin.Context) {
	rawPath := c.Query("url")
	if rawPath == "" {
		rawPath = c.Param("url") // fallback if mapped as path param
	}

	if rawPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "File URL is required"})
		return
	}

	// Clean up path found in URL
	relativePath := filepath.Clean(rawPath)
	if strings.Contains(relativePath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "Invalid file path"})
		return
	}

	fullPath := filepath.Join(constants.MediaBaseDir, relativePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"state": constants.ErrorState, "data": "File not found"})
		return
	}

	// Gin's c.File automatically handles 'Accept-Ranges', 'Content-Range',
	// and 'Content-Type' headers which are required by the Flutter video_player.
	c.File(fullPath)
}
