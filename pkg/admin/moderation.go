package admin

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/moderation"
)

// ModerationHandler handles moderation configuration requests
type ModerationHandler struct {
	db        *sql.DB
	templates *template.Template
	store     *moderation.ConfigStore
}

// NewModerationHandler creates a new moderation handler
func NewModerationHandler(db *sql.DB, templates *template.Template) *ModerationHandler {
	return &ModerationHandler{
		db:        db,
		templates: templates,
		store:     moderation.NewConfigStore(db),
	}
}

// HandleModerationForm displays the moderation configuration form
func (h *ModerationHandler) HandleModerationForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get moderation config or use defaults
	config, err := h.store.GetBySiteID(siteID)
	if err != nil {
		// Config doesn't exist yet, use defaults
		defaultConfig := moderation.DefaultModerationConfig()
		config = &defaultConfig
	}

	data := map[string]interface{}{
		"SiteID": siteID,
		"Config": config,
	}

	if err := h.templates.ExecuteTemplate(w, "moderation/form.html", data); err != nil {
		log.Printf("Error rendering moderation form: %v", err)
		http.Error(w, "Failed to render form", http.StatusInternalServerError)
	}
}

// HandleModerationUpdate updates moderation configuration
func (h *ModerationHandler) HandleModerationUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Parse form values
	config := moderation.ModerationConfig{
		Enabled:              r.FormValue("enabled") == "on",
		CheckSpam:            r.FormValue("check_spam") == "on",
		CheckOffensive:       r.FormValue("check_offensive") == "on",
		CheckAggressive:      r.FormValue("check_aggressive") == "on",
		CheckOffTopic:        r.FormValue("check_off_topic") == "on",
	}

	// Parse thresholds
	autoRejectThreshold, err := strconv.ParseFloat(r.FormValue("auto_reject_threshold"), 64)
	if err != nil {
		autoRejectThreshold = 0.85
	}
	config.AutoRejectThreshold = autoRejectThreshold

	autoApproveThreshold, err := strconv.ParseFloat(r.FormValue("auto_approve_threshold"), 64)
	if err != nil {
		autoApproveThreshold = 0.30
	}
	config.AutoApproveThreshold = autoApproveThreshold

	// Check if config exists
	_, err = h.store.GetBySiteID(siteID)
	if err != nil {
		// Config doesn't exist, create it
		if err := h.store.Create(siteID, config); err != nil {
			log.Printf("Error creating moderation config: %v", err)
			http.Error(w, "Failed to create configuration", http.StatusInternalServerError)
			return
		}
	} else {
		// Config exists, update it
		if err := h.store.Update(siteID, config); err != nil {
			log.Printf("Error updating moderation config: %v", err)
			http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
			return
		}
	}

	// Return success response for HTMX
	w.Header().Set("HX-Trigger", "configUpdated")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Configuration updated successfully")
}
