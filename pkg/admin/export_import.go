package admin

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/auth"
	"github.com/saasuke-labs/kotomi/pkg/export"
	importpkg "github.com/saasuke-labs/kotomi/pkg/import"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// ExportImportHandler handles export and import operations
type ExportImportHandler struct {
	db        *sql.DB
	templates *template.Template
}

// NewExportImportHandler creates a new export/import handler
func NewExportImportHandler(db *sql.DB, templates *template.Template) *ExportImportHandler {
	return &ExportImportHandler{
		db:        db,
		templates: templates,
	}
}

// ShowExportForm displays the export form for a site
func (h *ExportImportHandler) ShowExportForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r.Context())

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

	data := map[string]interface{}{
		"Site": site,
	}

	if err := h.templates.ExecuteTemplate(w, "admin/export_import/export_form.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// ExportData handles the export operation
func (h *ExportImportHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r.Context())

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

	// Get format from query parameter (default to JSON)
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	exporter := export.NewExporter(h.db)

	switch format {
	case "json":
		exportData, err := exporter.ExportToJSON(siteID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
			return
		}

		filename := export.GetExportFilename(site.Name, "json")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		if err := exporter.WriteJSON(w, exportData); err != nil {
			http.Error(w, fmt.Sprintf("Failed to write export: %v", err), http.StatusInternalServerError)
			return
		}

	case "csv-comments":
		var buf bytes.Buffer
		if err := exporter.ExportToCSV(&buf, siteID); err != nil {
			http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
			return
		}

		filename := export.GetExportFilename(site.Name+"_comments", "csv")
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Write(buf.Bytes())

	case "csv-reactions":
		var buf bytes.Buffer
		if err := exporter.ExportReactionsToCSV(&buf, siteID); err != nil {
			http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
			return
		}

		filename := export.GetExportFilename(site.Name+"_reactions", "csv")
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		w.Write(buf.Bytes())

	default:
		http.Error(w, "Invalid format", http.StatusBadRequest)
		return
	}
}

// ShowImportForm displays the import form for a site
func (h *ExportImportHandler) ShowImportForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r.Context())

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

	data := map[string]interface{}{
		"Site": site,
	}

	if err := h.templates.ExecuteTemplate(w, "admin/export_import/import_form.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// ImportData handles the import operation
func (h *ExportImportHandler) ImportData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r.Context())

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

	// Parse multipart form (limit to 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get duplicate strategy
	strategy := r.FormValue("strategy")
	if strategy == "" {
		strategy = string(importpkg.StrategySkip)
	}

	// Create importer
	importer := importpkg.NewImporter(h.db, importpkg.DuplicateStrategy(strategy))

	// Determine format from file extension
	var result *importpkg.ImportResult
	if header.Filename[len(header.Filename)-5:] == ".json" {
		result, err = importer.ImportFromJSON(file, siteID)
	} else if header.Filename[len(header.Filename)-4:] == ".csv" {
		result, err = importer.ImportFromCSV(file, siteID)
	} else {
		http.Error(w, "Unsupported file format (must be .json or .csv)", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Import failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
		"message": fmt.Sprintf("Import completed: %d comments imported, %d skipped, %d updated",
			result.CommentsImported, result.CommentsSkipped, result.CommentsUpdated),
	})
}

// ExportDataAPI provides API endpoint for exports (returns JSON in response)
func (h *ExportImportHandler) ExportDataAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r.Context())

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

	exporter := export.NewExporter(h.db)
	exportData, err := exporter.ExportToJSON(siteID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exportData)
}

// ImportDataAPI provides API endpoint for imports (accepts JSON directly)
func (h *ExportImportHandler) ImportDataAPI(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get user ID from context
	userID := auth.GetUserIDFromContext(r.Context())

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

	// Get duplicate strategy from query parameter
	strategy := r.URL.Query().Get("strategy")
	if strategy == "" {
		strategy = string(importpkg.StrategySkip)
	}

	// Create importer
	importer := importpkg.NewImporter(h.db, importpkg.DuplicateStrategy(strategy))

	// Import from request body
	result, err := importer.ImportFromJSON(r.Body, siteID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Import failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return result as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
	})
}

// DownloadExport generates and downloads an export file
func (h *ExportImportHandler) DownloadExport(w http.ResponseWriter, r *http.Request) {
	// This is an alias for ExportData for clearer routing
	h.ExportData(w, r)
}
