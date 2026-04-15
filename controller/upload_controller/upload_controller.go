package uploadcontroller

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const uploadRootDir = "upload"

var allowedImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
	".gif":  {},
}

func UploadAvatar(c *fiber.Ctx) error {
	return uploadImage(c, "avatars", "Avatar uploaded successfully")
}

func UploadItemImage(c *fiber.Ctx) error {
	return uploadImage(c, "items", "Item image uploaded successfully")
}

func uploadImage(c *fiber.Ctx, subDir string, successMessage string) error {
	context := fiber.Map{
		"statusText": "Ok",
		"msg":        successMessage,
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Image file is required"
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	if err := validateImageFile(fileHeader); err != nil {
		context["statusText"] = "bad"
		context["msg"] = err.Error()
		return c.Status(fiber.StatusBadRequest).JSON(context)
	}

	storageDir := filepath.Join(uploadRootDir, subDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Error preparing upload directory"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	fileName := buildUploadFileName(fileHeader.Filename)
	storagePath := filepath.Join(storageDir, fileName)
	if err := c.SaveFile(fileHeader, storagePath); err != nil {
		context["statusText"] = "bad"
		context["msg"] = "Error saving image"
		return c.Status(fiber.StatusInternalServerError).JSON(context)
	}

	publicPath := "/upload/" + subDir + "/" + fileName
	context["path"] = publicPath
	context["file_name"] = fileName
	return c.Status(fiber.StatusCreated).JSON(context)
}

func validateImageFile(fileHeader *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := allowedImageExtensions[ext]; !ok {
		return fmt.Errorf("Only jpg, jpeg, png, webp and gif files are allowed")
	}

	contentType := strings.ToLower(strings.TrimSpace(fileHeader.Header.Get("Content-Type")))
	if contentType != "" && !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("Only image files are allowed")
	}

	return nil
}

func buildUploadFileName(originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	return strconv.FormatInt(time.Now().UnixNano(), 10) + ext
}
