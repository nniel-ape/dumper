package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
	io.Copy(w, reader)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r.Context())

	vault, err := s.stores.GetVault(userID)
	if err != nil {
		jsonError(w, "failed to access vault", http.StatusInternalServerError)
		return
	}

	count, _ := vault.ItemCount()
	tags, _ := vault.GetAllTags()

	jsonResponse(w, map[string]interface{}{
		"items": count,
		"tags":  len(tags),
	})
}
