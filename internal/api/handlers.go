package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nerdneilsfield/dumper/internal/export"
)

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// Check for tag filter
	tag := r.URL.Query().Get("tag")
	if tag != "" {
		items, err := vault.ListItemsByTag(tag, limit, offset)
		if err != nil {
			jsonError(w, "failed to list items", http.StatusInternalServerError)
			return
		}
		jsonResponse(w, items)
		return
	}

	items, err := vault.ListItems(limit, offset)
	if err != nil {
		jsonError(w, "failed to list items", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, items)
}

func (s *Server) handleGetItem(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	itemID := r.PathValue("id")

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	item, err := vault.GetItem(itemID)
	if err != nil {
		jsonError(w, "failed to get item", http.StatusInternalServerError)
		return
	}
	if item == nil {
		jsonError(w, "item not found", http.StatusNotFound)
		return
	}

	jsonResponse(w, item)
}

func (s *Server) handleGetItemImage(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	itemID := r.PathValue("id")

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		http.Error(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	// Get item to verify ownership and get image path
	item, err := vault.GetItem(itemID)
	if err != nil {
		http.Error(w, "failed to get item", http.StatusInternalServerError)
		return
	}
	if item == nil {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	// Check if item has an image
	if item.ImagePath == "" {
		http.Error(w, "item has no image", http.StatusNotFound)
		return
	}

	// Construct full path to image file
	userDir := s.stores.UserDir(userID)

	// Security: Validate ImagePath doesn't contain path traversal attempts
	// Check BEFORE joining to prevent bypass via filepath.Clean normalization
	if item.ImagePath == "" || strings.Contains(item.ImagePath, "..") || filepath.IsAbs(item.ImagePath) {
		http.Error(w, "invalid image path", http.StatusForbidden)
		return
	}

	imagePath := filepath.Join(userDir, item.ImagePath)

	// Additional safety check: ensure final path stays within user directory
	cleanPath := filepath.Clean(imagePath)
	cleanUserDir := filepath.Clean(userDir)
	if cleanPath == cleanUserDir || !strings.HasPrefix(cleanPath, cleanUserDir+string(filepath.Separator)) {
		http.Error(w, "invalid image path", http.StatusForbidden)
		return
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		http.Error(w, "image file not found", http.StatusNotFound)
		return
	}

	// Detect content type from extension
	ext := strings.ToLower(filepath.Ext(item.ImagePath))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	// Serve the file with caching
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, imagePath)
}

func (s *Server) handleGetItemImageByIndex(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	itemID := r.PathValue("id")
	indexStr := r.PathValue("index")

	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		http.Error(w, "invalid image index", http.StatusBadRequest)
		return
	}

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		http.Error(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	item, err := vault.GetItem(itemID)
	if err != nil {
		http.Error(w, "failed to get item", http.StatusInternalServerError)
		return
	}
	if item == nil {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}

	if len(item.ImagePaths) == 0 {
		http.Error(w, "item has no images", http.StatusNotFound)
		return
	}

	if index >= len(item.ImagePaths) {
		http.Error(w, "image index out of range", http.StatusNotFound)
		return
	}

	imagePath := item.ImagePaths[index]
	userDir := s.stores.UserDir(userID)

	// Security: validate path doesn't contain traversal attempts
	if imagePath == "" || strings.Contains(imagePath, "..") || filepath.IsAbs(imagePath) {
		http.Error(w, "invalid image path", http.StatusForbidden)
		return
	}

	fullPath := filepath.Join(userDir, imagePath)

	cleanPath := filepath.Clean(fullPath)
	cleanUserDir := filepath.Clean(userDir)
	if cleanPath == cleanUserDir || !strings.HasPrefix(cleanPath, cleanUserDir+string(filepath.Separator)) {
		http.Error(w, "invalid image path", http.StatusForbidden)
		return
	}

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "image file not found", http.StatusNotFound)
		return
	}

	ext := strings.ToLower(filepath.Ext(imagePath))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, fullPath)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	itemID := r.PathValue("id")

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	if err := vault.DeleteItem(itemID); err != nil {
		jsonError(w, "failed to delete item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())
	query := r.URL.Query().Get("q")
	if query == "" {
		jsonError(w, "query required", http.StatusBadRequest)
		return
	}

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	results, err := vault.Search(query, 20)
	if err != nil {
		jsonError(w, "search failed", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, results)
}

func (s *Server) handleGetTags(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	tags, err := vault.GetAllTags()
	if err != nil {
		jsonError(w, "failed to get tags", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, tags)
}

func (s *Server) handleGetGraph(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	items, relationships, err := vault.GetGraph()
	if err != nil {
		jsonError(w, "failed to get graph", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"nodes": items,
		"edges": relationships,
	})
}

func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	var req struct {
		Question string `json:"question"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		jsonError(w, "question required", http.StatusBadRequest)
		return
	}

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	// Search for relevant items
	results, err := vault.Search(req.Question, 5)
	if err != nil {
		jsonError(w, "search failed", http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		jsonResponse(w, map[string]string{
			"answer": "I couldn't find any relevant items in your vault to answer this question.",
		})
		return
	}

	// Format items for LLM
	var itemsStr []string
	for _, r := range results {
		itemsStr = append(itemsStr, fmt.Sprintf("Title: %s\nSummary: %s\nContent: %s",
			r.Item.Title, r.Item.Summary, r.Item.Content))
	}

	answer, err := s.llmClient.AnswerQuestion(context.Background(), req.Question, itemsStr)
	if err != nil {
		jsonError(w, "failed to generate answer", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"answer":  answer,
		"sources": results,
	})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	exporter := export.NewObsidianExporter()
	reader, err := exporter.Export(vault)
	if err != nil {
		jsonError(w, "failed to export", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=dumper-export.zip")
	if _, err := io.Copy(w, reader); err != nil {
		slog.Error("failed to write export to response", "error", err)
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	count, err := vault.ItemCount()
	if err != nil {
		jsonError(w, "failed to get item count", http.StatusInternalServerError)
		return
	}

	tags, err := vault.GetAllTags()
	if err != nil {
		jsonError(w, "failed to get tags", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]interface{}{
		"items": count,
		"tags":  len(tags),
	})
}
