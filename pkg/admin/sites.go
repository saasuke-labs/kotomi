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

// SitesHandler handles site-related requests
type SitesHandler struct {
	db        *sql.DB
	templates *template.Template
}

// NewSitesHandler creates a new sites handler
func NewSitesHandler(db *sql.DB, templates *template.Template) *SitesHandler {
	return &SitesHandler{
		db:        db,
		templates: templates,
	}
}

// ListSites handles GET /admin/sites
func (h *SitesHandler) ListSites(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	sites, err := siteStore.GetByOwner(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch sites", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// Return partial HTML for HTMX
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "sites/list.html", map[string]interface{}{
				"Sites": sites,
			})
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		}
		return
	}

	// Return JSON for API requests
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}

// GetSite handles GET /admin/sites/{siteId}
func (h *SitesHandler) GetSite(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Check ownership
	if site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Get pages for this site
	pageStore := models.NewPageStore(h.db)
	pages, err := pageStore.GetBySite(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Failed to fetch pages", http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		if h.templates != nil {
			err = h.templates.ExecuteTemplate(w, "sites/detail.html", map[string]interface{}{
				"Site":  site,
				"Pages": pages,
			})
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		}
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"site":  site,
		"pages": pages,
	})
}

// CreateSite handles POST /admin/sites
func (h *SitesHandler) CreateSite(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	domain := r.FormValue("domain")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.Create(r.Context(), userID, name, domain, description)
	if err != nil {
		http.Error(w, "Failed to create site", http.StatusInternalServerError)
		return
	}

	// For HTMX requests, redirect to the site detail page
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/sites/"+site.ID)
		w.WriteHeader(http.StatusOK)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(site)
}

// UpdateSite handles PUT /admin/sites/{siteId}
func (h *SitesHandler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	if site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	domain := r.FormValue("domain")
	description := r.FormValue("description")

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	err = siteStore.Update(r.Context(), siteID, name, domain, description)
	if err != nil {
		http.Error(w, "Failed to update site", http.StatusInternalServerError)
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

// DeleteSite handles DELETE /admin/sites/{siteId}
func (h *SitesHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Verify ownership
	siteStore := models.NewSiteStore(h.db)
	site, err := siteStore.GetByID(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	if site.OwnerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	err = siteStore.Delete(r.Context(), siteID)
	if err != nil {
		http.Error(w, "Failed to delete site", http.StatusInternalServerError)
		return
	}

	// For HTMX requests
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/sites")
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ShowSiteForm handles GET /admin/sites/new and GET /admin/sites/{siteId}/edit
func (h *SitesHandler) ShowSiteForm(w http.ResponseWriter, r *http.Request) {
	if h.templates == nil {
		http.Error(w, "Templates not available", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	siteID := vars["siteId"]

	data := map[string]interface{}{
		"IsEdit": siteID != "",
	}

	if siteID != "" {
		userID := auth.GetUserIDFromContext(r.Context())
		siteStore := models.NewSiteStore(h.db)
		site, err := siteStore.GetByID(r.Context(), siteID)
		if err != nil || site.OwnerID != userID {
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}
		data["Site"] = site
	}

	err := h.templates.ExecuteTemplate(w, "sites/form.html", data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
