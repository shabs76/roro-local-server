package specials

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/chai2010/webp"
	"github.com/shabs76/roro-local-server/constants"
	"golang.org/x/crypto/bcrypt"
)

func RandomString(length int, suffix string) string {
	// Seed the random number generator with the current time
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Characters to choose from for the random string
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create the random string
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	// Add the suffix to the random string
	result = append(result, []byte(suffix)...)

	return string(result)
}

func RandomStringNoSuffix(length int) string {
	// Seed the random number generator with the current time
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Characters to choose from for the random string
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create the random string
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	// Add the suffix to the random string
	return string(result)
}

func RandomNumb(length int) string {
	// Seed the random number generator with the current time
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Characters to choose from for the random string
	charset := "01234567890987654321"

	// Create the random string
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	// Add the suffix to the random string
	sttr := string(result)

	return sttr
}

func GenerateRandomInt32(length int) int32 {
	// Set up a seed for the random number generator
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Define the minimum and maximum values for the specified length
	min := int32(1)
	max := int32(10)

	// Adjust the minimum and maximum values based on the specified length
	for i := 1; i < length; i++ {
		min *= 10
		max *= 10
	}

	// Generate a random int32 number within the specified range
	return min + rand.Int31n(max-min)
}

func GeneratePasswordHash(password string) (string, error) {
	// Generate a salted hash for the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePasswordHash compares a password against a bcrypt hash
func ComparePasswordHash(password, hash string) bool {
	// Compare the given password with the hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func IsValidPhoneNumber(phone string) bool {
	// Define the regex pattern for the phone number
	var validPhonePattern = `^255[76]\d{8}$`
	re := regexp.MustCompile(validPhonePattern)
	return re.MatchString(phone)
}

func IsValidEmail(email string) bool {
	// Define the regex pattern for the email address
	var validEmailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(validEmailPattern)
	return re.MatchString(email)
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func ExtractBucketAndKey(s3Link string) (bucket, key string, err error) {
	// Parse the S3 URL
	// Find the start and end positions for bucket and key
	start := strings.Index(s3Link, "https://") + len("https://")
	endBucket := strings.Index(s3Link, ".s3.amazonaws.com")
	startKey := strings.Index(s3Link, ".amazonaws.com/") + len(".amazonaws.com/")

	// Check if the required substrings were found
	if start < len("https://") || endBucket == -1 || startKey < len(".amazonaws.com/") {
		return "", "", fmt.Errorf("invalid S3 URL format")
	}

	// Extract bucket and key using the identified positions
	bucket = s3Link[start:endBucket]
	key = s3Link[startKey:]

	return bucket, key, nil
}

func GetObjectBytes(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic("failed to fetch image: -" + url + "--" + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("non-200 response from image URL: -" + url + "--" + resp.Status)
	}

	webpData, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("failed to read webp bytes: -" + url + "--" + err.Error())
	}

	img, err := webp.Decode(bytes.NewReader(webpData))
	if err != nil {
		panic("failed to decode webp: -" + url + "--" + err.Error())
	}

	var pngBuf bytes.Buffer
	err = png.Encode(&pngBuf, img)
	if err != nil {
		panic("failed to encode to png: -" + url + "--" + err.Error())
	}

	return pngBuf.Bytes()
}

func GetObjectBytesNoneConvert(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic("failed to fetch image: -" + url + "--" + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("non-200 response from image URL: -" + url + "--" + resp.Status)
	}

	webpData, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("failed to read webp bytes: -" + url + "--" + err.Error())
	}

	return webpData
}

func ShortTimeDate(inputTime string) (*string, error) {
	layoutIn := "2006-01-02 15:04:05"
	layoutOut := "Jan 02, 15:04"

	t, err := time.Parse(layoutIn, inputTime)
	if err != nil {
		return nil, err
	}

	// Format the time to the desired output format
	formattedTime := t.Format(layoutOut)
	return &formattedTime, nil
}

func CapitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func GetEnvVariable(envKey, defaultVal string) string {
	value := os.Getenv(envKey)
	if value == "" {
		return defaultVal
	}
	return value
}

// GetImageBytes gets the bytes of an image from a local file or URL and returns it as WebP format.
func GetImageBytes(imagePath string) ([]byte, error) {
	var inputBytes []byte
	var err error

	// 1. Fetch Bytes
	if strings.HasPrefix(imagePath, "http") {
		// Remote URL
		resp, err := http.Get(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch image from URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
		}

		inputBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
	} else {
		// Local File
		// Normalize path
		localPath := constants.MediaBaseDir + "/" + imagePath

		// Read file
		inputBytes, err = os.ReadFile(localPath)
		if err != nil {
			// Fallback: try reading the original path if modification failed
			if localPath != imagePath {
				inputBytes, err = os.ReadFile(imagePath)
				if err != nil {
					return nil, fmt.Errorf("failed to read local file '%s': %v", localPath, err)
				}
			}
		}
	}

	// 2. Check if already WebP (RIFF....WEBP)
	if len(inputBytes) > 12 && string(inputBytes[0:4]) == "RIFF" && string(inputBytes[8:12]) == "WEBP" {
		return inputBytes, nil
	}

	// 3. Convert to WebP
	// Identify and decode image
	img, _, err := image.Decode(bytes.NewReader(inputBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Encode to WebP
	var buf bytes.Buffer
	// Lossless: false for better compression usually, or true if requirement implies exactness.
	// Default quality ~80 if not specified.
	if err := webp.Encode(&buf, img, nil); err != nil {
		return nil, fmt.Errorf("failed to encode to WebP: %v", err)
	}

	return buf.Bytes(), nil
}

// GetImageBytesNoneWebp gets the bytes of an image from a local file or URL and returns it as PNG format.
// This is used for PDF generation which does not support WebP.
func GetImageBytesNoneWebp(imagePath string) ([]byte, error) {
	var inputBytes []byte
	var err error

	// 1. Fetch Bytes
	if strings.HasPrefix(imagePath, "http") {
		// Remote URL
		resp, err := http.Get(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch image from URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
		}

		inputBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
	} else {
		// Local File
		// Normalize path
		localPath := constants.MediaBaseDir + "/" + imagePath

		// Read file
		inputBytes, err = os.ReadFile(localPath)
		if err != nil {
			// Fallback: try reading the original path if modification failed
			if localPath != imagePath {
				inputBytes, err = os.ReadFile(imagePath)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to read local file '%s': %v", localPath, err)
			}
		}
	}

	// 2. Decode image (any supported format, including WebP if imported)
	// We need to decode it to encode it as PNG/JPEG later.
	// We imported github.com/chai2010/webp so image.Decode should handle WebP if registered.
	// But standard image.Decode relies on registrations.
	// "image/png" and "image/jpeg" are imported.
	// "golang.org/x/image/webp" or "github.com/chai2010/webp" registers itself usually?
	// Let's check imports in special_functions.go.

	// If it is WebP, we might need special handling if image.Decode doesn't pick it up.
	// But let's try standard decode first.
	img, _, err := image.Decode(bytes.NewReader(inputBytes))
	if err != nil {
		// If standard decode failed, maybe it's WebP and the decoder wasn't registered automatically or correctly?
		// Try decoding explicitly with webp library
		img, err = webp.Decode(bytes.NewReader(inputBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %v", err)
		}
	}

	// 3. Convert to PNG (Safe for Maroto/PDF)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode to PNG: %v", err)
	}

	return buf.Bytes(), nil
}

// GetImageBytesForPdf fetches image and returns it as PNG bytes.
// This is necessary because Maroto/gofpdf typically does not support WebP.
func GetImageBytesForPdf(imagePath string) ([]byte, error) {
	var inputBytes []byte
	var err error

	// 1. Fetch Bytes
	if strings.HasPrefix(imagePath, "http") {
		// Remote URL
		resp, err := http.Get(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch image from URL: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
		}

		inputBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}
	} else {
		// Local File Logic
		localPath := constants.MediaBaseDir + "/" + imagePath

		inputBytes, err = os.ReadFile(localPath)
		if err != nil {
			if localPath != imagePath {
				inputBytes, err = os.ReadFile(imagePath)
				if err != nil {
					return nil, fmt.Errorf("failed to read local file '%s': %v", localPath, err)
				}
			} else {
				return nil, fmt.Errorf("failed to read local file '%s': %v", localPath, err)
			}
		}
	}

	// 2. Decode (try standard, then webp explicitly)
	img, _, err := image.Decode(bytes.NewReader(inputBytes))
	if err != nil {
		img, err = webp.Decode(bytes.NewReader(inputBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to decode image: %v", err)
		}
	}

	// Ensure 8-bit depth (convert to RGBA)
	// This fixes "16-bit depth not supported in PNG file" errors in PDF generation
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	img = rgba

	// 3. Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode to PNG: %v", err)
	}

	return buf.Bytes(), nil
}
