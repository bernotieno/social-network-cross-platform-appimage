package utils

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Allowed image MIME types
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
}

// MaxImageSize is the maximum allowed image size in bytes (5MB)
const MaxImageSize = 5 * 1024 * 1024

// SaveImage saves an uploaded image to the uploads directory
func SaveImage(file multipart.File, header *multipart.FileHeader, directory string) (string, error) {
	// Check file size
	if header.Size > MaxImageSize {
		return "", errors.New("file size exceeds the limit")
	}

	// Check file type
	contentType := header.Header.Get("Content-Type")
	extension, ok := allowedImageTypes[contentType]
	if !ok {
		return "", errors.New("invalid file type, only JPEG, PNG, and GIF are allowed")
	}

	// Create uploads directory if it doesn't exist
	uploadDir := filepath.Join("uploads", directory)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate a unique filename
	filename := uuid.New().String() + extension
	filePath := filepath.Join(uploadDir, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the file content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Return the relative path to the file
	return filepath.Join("/uploads", directory, filename), nil
}

// DeleteImage deletes an image file
func DeleteImage(imagePath string) error {
	// Ensure the path is within the uploads directory
	if !strings.HasPrefix(imagePath, "/uploads/") {
		return errors.New("invalid image path")
	}

	// Remove the leading slash
	localPath := imagePath[1:]

	// Check if the file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	// Delete the file
	if err := os.Remove(localPath); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// ValidateImageType validates the content type of an image
func ValidateImageType(contentType string) bool {
	_, ok := allowedImageTypes[contentType]
	return ok
}

// DetectContentType detects the content type of a file
func DetectContentType(file multipart.File) (string, error) {
	// Read the first 512 bytes to detect the content type
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Reset the file pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}
