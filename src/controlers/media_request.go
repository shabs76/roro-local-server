package controlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shabs76/roro-local-server/constants"
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
		c.JSON(http.StatusBadRequest, gin.H{"state": constants.ErrorState, "data": "No file uploaded"})
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

	// Append UNIX timestamp to ensure uniqueness
	ext := filepath.Ext(name)
	nameWithoutExt := strings.ReplaceAll(strings.TrimSuffix(name, ext), " ", "_")
	name = fmt.Sprintf("%s_%d%s", nameWithoutExt, time.Now().Unix(), ext)

	mediaType := getMediaType(file.Filename)
	saveDir := filepath.Join(constants.MediaBaseDir, mediaType)

	if err := ensureDir(saveDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to create directory"})
		return
	}

	dst := filepath.Join(saveDir, name) // Use 'name' here, not filepath.Base(name) again
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"state": constants.ErrorState, "data": "Failed to save file"})
		return
	}
	url := fmt.Sprintf("%s/%s", mediaType, name)
	c.JSON(http.StatusOK, gin.H{
		"state":   constants.SuccessState,
		"message": "File uploaded successfully",
		"info": gin.H{
			"url":      url,
			"type":     mediaType,
			"filename": name,
			"path":     dst,
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
