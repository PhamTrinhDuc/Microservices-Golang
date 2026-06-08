package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"backend/internal/storage"
	"backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UploadController struct {
	storageProvider storage.StorageProvider
}

func NewUploadController(sp storage.StorageProvider) *UploadController {
	return &UploadController{storageProvider: sp}
}

func (ctl *UploadController) UploadImage(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		utils.RespondBadRequest(ctx, "file parameter is required")
		return
	}
	defer file.Close()

	if header.Size > 5*1024*1024 {
		utils.RespondBadRequest(ctx, "file size exceeds maximum limit of 5MB")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		utils.RespondBadRequest(ctx, "only image file uploads are allowed")
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".svg": true}
	if !allowed[ext] {
		utils.RespondBadRequest(ctx, fmt.Sprintf("unsupported image format: %s", ext))
		return
	}

	url, err := ctl.storageProvider.UploadFile(ctx.Request.Context(), header.Filename, file, header.Size, contentType)
	if err != nil {
		utils.RespondInternalError(ctx, fmt.Sprintf("upload failed: %v", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}
