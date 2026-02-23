package http_adapter

import (
	"encoding/json"
	"net/http"

	"vnp-network-backend/mcp/internal/core/tools"
)

type Handler struct {
	registry *tools.Registry
}

func NewHandler(registry *tools.Registry) *Handler {
	return &Handler{
		registry: registry,
	}
}

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/tools", h.ListTools)
	mux.HandleFunc("/tools/execute", h.ExecuteTool)
	return mux
}

func (h *Handler) ListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := h.registry.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

func (h *Handler) ExecuteTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string                 `json:"name"`
		Args map[string]interface{} `json:"args"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tool, ok := h.registry.Get(req.Name)
	if !ok {
		http.Error(w, "Tool not found", http.StatusNotFound)
		return
	}

	// In a real implementation, we would validate args against schema here.
	if tool.Handler == nil {
		http.Error(w, "Tool handler not implemented", http.StatusInternalServerError)
		return
	}

	result, err := tool.Handler(req.Args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
