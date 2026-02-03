package admin

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/saasuke-labs/kotomi/pkg/notifications"
)

// NotificationsHandler handles notification configuration requests
type NotificationsHandler struct {
	db        *sql.DB
	templates *template.Template
	store     *notifications.Store
}

// NewNotificationsHandler creates a new notifications handler
func NewNotificationsHandler(db *sql.DB, templates *template.Template) *NotificationsHandler {
	return &NotificationsHandler{
		db:        db,
		templates: templates,
		store:     notifications.NewStore(db),
	}
}

// HandleNotificationsForm displays the notification configuration form
func (h *NotificationsHandler) HandleNotificationsForm(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get notification settings or use defaults
	settings, err := h.store.GetSettings(siteID)
	if err != nil {
		log.Printf("Error getting notification settings: %v", err)
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	// If no settings exist, create default
	if settings == nil {
		settings = &notifications.NotificationSettings{
			SiteID:           siteID,
			Enabled:          false,
			Provider:         "smtp",
			FromEmail:        "",
			FromName:         "",
			SMTPEncryption:   "tls",
			SMTPPort:         587,
			NotifyNewComment: true,
			NotifyReply:      true,
			NotifyModeration: true,
		}
	}

	data := map[string]interface{}{
		"SiteID":   siteID,
		"Settings": settings,
	}

	if err := h.templates.ExecuteTemplate(w, "admin/notifications/form.html", data); err != nil {
		log.Printf("Error rendering notifications form: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// HandleNotificationsUpdate saves the notification configuration
func (h *NotificationsHandler) HandleNotificationsUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Parse form values
	enabled := r.FormValue("enabled") == "on"
	provider := r.FormValue("provider")
	fromEmail := r.FormValue("from_email")
	fromName := r.FormValue("from_name")
	replyTo := r.FormValue("reply_to")
	ownerEmail := r.FormValue("owner_email")
	
	// SMTP settings
	smtpHost := r.FormValue("smtp_host")
	smtpPortStr := r.FormValue("smtp_port")
	smtpUser := r.FormValue("smtp_user")
	smtpPassword := r.FormValue("smtp_password")
	smtpEncryption := r.FormValue("smtp_encryption")
	
	// SendGrid settings
	sendGridAPIKey := r.FormValue("sendgrid_api_key")
	
	// Notification types
	notifyNewComment := r.FormValue("notify_new_comment") == "on"
	notifyReply := r.FormValue("notify_reply") == "on"
	notifyModeration := r.FormValue("notify_moderation") == "on"

	// Parse SMTP port
	smtpPort := 587
	if smtpPortStr != "" {
		var err error
		smtpPort, err = strconv.Atoi(smtpPortStr)
		if err != nil {
			http.Error(w, "Invalid SMTP port", http.StatusBadRequest)
			return
		}
	}

	// Get existing settings or create new
	settings, err := h.store.GetSettings(siteID)
	if err != nil {
		log.Printf("Error getting settings: %v", err)
		http.Error(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	if settings == nil {
		settings = &notifications.NotificationSettings{
			SiteID: siteID,
		}
	}

	// Update settings
	settings.Enabled = enabled
	settings.Provider = provider
	settings.FromEmail = fromEmail
	settings.FromName = fromName
	settings.ReplyTo = replyTo
	settings.OwnerEmail = ownerEmail
	settings.SMTPHost = smtpHost
	settings.SMTPPort = smtpPort
	settings.SMTPUser = smtpUser
	if smtpPassword != "" {
		settings.SMTPPassword = smtpPassword
	}
	settings.SMTPEncryption = smtpEncryption
	if sendGridAPIKey != "" {
		settings.SendGridAPIKey = sendGridAPIKey
	}
	settings.NotifyNewComment = notifyNewComment
	settings.NotifyReply = notifyReply
	settings.NotifyModeration = notifyModeration

	// Save settings
	if err := h.store.SaveSettings(settings); err != nil {
		log.Printf("Error saving notification settings: %v", err)
		http.Error(w, "Failed to save settings", http.StatusInternalServerError)
		return
	}

	// Redirect back to the form with success message
	http.Redirect(w, r, "/admin/sites/"+siteID+"/notifications?success=1", http.StatusSeeOther)
}

// HandleTestEmail sends a test email
func (h *NotificationsHandler) HandleTestEmail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	siteID := vars["siteId"]

	// Get notification settings
	settings, err := h.store.GetSettings(siteID)
	if err != nil || settings == nil {
		http.Error(w, "Settings not configured", http.StatusBadRequest)
		return
	}

	if !settings.Enabled {
		http.Error(w, "Notifications are not enabled", http.StatusBadRequest)
		return
	}

	// Create a test notification
	testEmail := r.URL.Query().Get("email")
	if testEmail == "" {
		testEmail = settings.OwnerEmail
	}

	// Create email provider
	var provider notifications.EmailProvider
	switch settings.Provider {
	case "smtp":
		provider = notifications.NewSMTPProvider(
			settings.SMTPHost,
			settings.SMTPPort,
			settings.SMTPUser,
			settings.SMTPPassword,
			settings.FromEmail,
			settings.FromName,
			settings.SMTPEncryption,
		)
	case "sendgrid":
		provider = notifications.NewSendGridProvider(
			settings.SendGridAPIKey,
			settings.FromEmail,
			settings.FromName,
		)
	default:
		http.Error(w, "Unknown provider", http.StatusBadRequest)
		return
	}

	sender := notifications.NewEmailSender(provider)

	// Send test email
	testBody := `
		<html>
		<body>
			<h1>Test Email</h1>
			<p>This is a test email from Kotomi notification system.</p>
			<p>If you received this email, your email configuration is working correctly!</p>
		</body>
		</html>
	`

	err = sender.Send(r.Context(), testEmail, "Kotomi Test Email", testBody)
	if err != nil {
		log.Printf("Test email failed: %v", err)
		http.Error(w, "Failed to send test email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Test email sent successfully!"))
}
