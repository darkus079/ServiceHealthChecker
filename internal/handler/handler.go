package handler

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"servicehealthchecker/internal/models"
	"servicehealthchecker/internal/pdf"
	"servicehealthchecker/internal/service"
	"servicehealthchecker/internal/storage"
)

type Handler struct {
	checker     *service.Checker
	storage     *storage.Storage
	pdfGen      *pdf.Generator
	isShutdown  *atomic.Bool
}

func New(checker *service.Checker, storage *storage.Storage, pdfGen *pdf.Generator, isShutdown *atomic.Bool) *Handler {
	return &Handler{
		checker:    checker,
		storage:    storage,
		pdfGen:     pdfGen,
		isShutdown: isShutdown,
	}
}

func (h *Handler) CheckLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Links) == 0 {
		http.Error(w, "No links provided", http.StatusBadRequest)
		return
	}

	if h.isShutdown.Load() {
		id, err := h.storage.CreateLinkSet(req.Links)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := h.storage.AddPendingTask(id, req.Links); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		resp := models.CheckResponse{
			Links:    make(map[string]string),
			LinksNum: id,
		}
		for _, link := range req.Links {
			resp.Links[link] = "pending"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	statuses, id, err := h.checker.CheckLinks(req.Links)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := models.CheckResponse{
		Links:    statuses,
		LinksNum: id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.LinksList) == 0 {
		http.Error(w, "No link set IDs provided", http.StatusBadRequest)
		return
	}

	linkSets := h.storage.GetLinkSets(req.LinksList)
	if len(linkSets) == 0 {
		http.Error(w, "No completed link sets found", http.StatusNotFound)
		return
	}

	pdfData, err := h.pdfGen.Generate(linkSets)
	if err != nil {
		http.Error(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=report.pdf")
	w.Write(pdfData)
}

