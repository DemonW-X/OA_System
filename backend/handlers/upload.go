package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const uploadDir = "./uploads"

var allowedImageTypes = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

var allowedFileTypes = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true,
	".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
	".txt": true, ".zip": true, ".rar": true,
}

const maxImageSize = 5 << 20  // 5MB
const maxFileSize = 20 << 20  // 20MB

func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "获取文件失败"})
		return
	}
	if file.Size > maxImageSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "图片不能超过5MB"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "不支持的图片格式"})
		return
	}
	url, err := saveUploadedFile(c, file, "images")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
		return
	}
	// wangEditor 图片上传要求的返回格式
	c.JSON(http.StatusOK, gin.H{
		"errno": 0,
		"data":  []gin.H{{"url": url, "alt": file.Filename, "href": ""}},
	})
}

func UploadAttachment(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "获取文件失败"})
		return
	}
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "附件不能超过20MB"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedFileTypes[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "不支持的附件格式"})
		return
	}
	url, err := saveUploadedFile(c, file, "attachments")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{"url": url, "name": file.Filename},
	})
}

func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, subDir string) (string, error) {
	dir := filepath.Join(uploadDir, subDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", err
	}
	return "/uploads/" + subDir + "/" + filename, nil
}
