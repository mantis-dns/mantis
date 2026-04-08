package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/gravity"
)

// RulesHandler handles custom rule CRUD.
type RulesHandler struct {
	repo    domain.CustomRuleRepository
	gravity *gravity.Engine
}

// List returns all custom rules.
func (h *RulesHandler) List(w http.ResponseWriter, r *http.Request) {
	rules, err := h.repo.List(r.Context())
	if err != nil {
		Error(w, "INTERNAL_ERROR", "failed to list rules", http.StatusInternalServerError)
		return
	}
	if rules == nil {
		rules = []domain.CustomRule{}
	}
	Success(w, rules)
}

// Create adds a new custom rule.
func (h *RulesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var rule domain.CustomRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}
	if rule.Domain == "" {
		Error(w, "VALIDATION_ERROR", "domain required", http.StatusBadRequest)
		return
	}
	if rule.Type != domain.RuleBlock && rule.Type != domain.RuleAllow {
		Error(w, "VALIDATION_ERROR", "type must be 'block' or 'allow'", http.StatusBadRequest)
		return
	}

	rule.ID = uuid.New().String()
	rule.Created = time.Now()

	if err := h.repo.Create(r.Context(), &rule); err != nil {
		Error(w, "INTERNAL_ERROR", "failed to create rule", http.StatusInternalServerError)
		return
	}

	// Update live gravity tree.
	if rule.Type == domain.RuleBlock {
		h.gravity.AddBlockRule(rule.Domain)
	} else {
		h.gravity.AddAllowRule(rule.Domain)
	}

	writeJSON(w, http.StatusCreated, successResponse{
		Data: rule,
		Meta: meta{RequestID: requestID()},
	})
}

// Delete removes a custom rule.
func (h *RulesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Find rule first to update gravity.
	rules, _ := h.repo.List(r.Context())
	for _, rule := range rules {
		if rule.ID == id {
			if rule.Type == domain.RuleBlock {
				h.gravity.RemoveBlockRule(rule.Domain)
			} else {
				h.gravity.RemoveAllowRule(rule.Domain)
			}
			break
		}
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		Error(w, "INTERNAL_ERROR", "failed to delete rule", http.StatusInternalServerError)
		return
	}
	Success(w, map[string]string{"status": "deleted"})
}
