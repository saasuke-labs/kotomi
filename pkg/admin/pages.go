package admin

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// PagesHandler handles page-related requests
type PagesHandler struct {
	db        *sql.DB
	templates *template.Template
}

// NewPagesHandler creates a new pages handler
func NewPagesHandler(db *sql.DB, templates *template.Template) *PagesHandler {
	return &PagesHandler{
		db:        db,
		templates: templates,
	}
}

// ListPages handles GET /admin/sites/{siteId}/pages
func (h *PagesHandler) ListPages(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	if site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	pageStore := models.NewPageStore(h.db)
	pages, err := pageStore.GetBySite(siteID)
	if err != nil {
		http.Error(w, "Failed to fetch pages", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "pages/list.html", map[string]interface{}{
				"Pages":  pages,
				"SiteID": siteID,
			})
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		}
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pages)
}

// GetPage handles GET /admin/sites/{siteId}/pages/{pageId}
func (h *PagesHandler) GetPage(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	pageID := vars["pageId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	pageStore := models.NewPageStore(h.db)
	page, err := pageStore.GetByID(pageID)
	if err != nil || page.SiteID != siteID {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(page)
}

// CreatePage handles POST /admin/sites/{siteId}/pages
func (h *PagesHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	title := r.FormValue("title")

	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	pageStore := models.NewPageStore(h.db)
	page, err := pageStore.Create(siteID, path, title)
	if err != nil {
		http.Error(w, "Failed to create page", http.StatusInternalServerError)
		return
	}

	// For HTMX requests
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/sites/"+siteID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(page)
}

// UpdatePage handles PUT /admin/sites/{siteId}/pages/{pageId}
func (h *PagesHandler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	pageID := vars["pageId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Verify page belongs to site
	pageStore := models.NewPageStore(h.db)
	page, err := pageStore.GetByID(pageID)
	if err != nil || page.SiteID != siteID {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	path := r.FormValue("path")
	title := r.FormValue("title")

	if path == "" {
		http.Error(w, "Path is required", http.StatusBadRequest)
		return
	}

	err = pageStore.Update(pageID, path, title)
	if err != nil {
		http.Error(w, "Failed to update page", http.StatusInternalServerError)
		return
	}

	// For HTMX requests
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/sites/"+siteID)
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletePage handles DELETE /admin/sites/{siteId}/pages/{pageId}
func (h *PagesHandler) DeletePage(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	pageID := vars["pageId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(siteID)
	if err != nil || site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Verify page belongs to site
	pageStore := models.NewPageStore(h.db)
	page, err := pageStore.GetByID(pageID)
	if err != nil || page.SiteID != siteID {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	err = pageStore.Delete(pageID)
	if err != nil {
		http.Error(w, "Failed to delete page", http.StatusInternalServerError)
		return
	}

	// For HTMX requests
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/sites/"+siteID)
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ShowPageForm handles GET /admin/sites/{siteId}/pages/new and GET /admin/sites/{siteId}/pages/{pageId}/edit
func (h *PagesHandler) ShowPageForm(w http.ResponseWriter, r *http.Request) {
	if h.templates == nil {
		http.Error(w, "Templates not available", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]
	pageID := vars["pageId"]

	data := map[string]interface{}{
		"IsEdit": pageID != "",
		"SiteID": siteID,
	}

	if pageID != "" {
		userID := auth.GetUserIDFromContext(r.Context())
		
		// Verify site ownership
		siteStore := models.NewSiteStore(h.db)
		site, err := siteStore.GetByID(siteID)
		if err != nil || site.OwnerID != userID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		pageStore := models.NewPageStore(h.db)
		page, err := pageStore.GetByID(pageID)
		if err != nil || page.SiteID != siteID {
			http.Error(w, "Page not found", http.StatusNotFound)
			return
		}
		data["Page"] = page
	}

	err := h.templates.ExecuteTemplate(w, "pages/form.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
