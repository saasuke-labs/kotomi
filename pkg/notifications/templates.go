package notifications

import (
	"bytes"
	"fmt"
	"html/template"
)

// EmailTemplate manages email template rendering
type EmailTemplate struct {
	templates *template.Template
}

// NewEmailTemplate creates a new email template manager
func NewEmailTemplate() *EmailTemplate {
	tmpl := template.New("email")

	// New comment template
	template.Must(tmpl.New("new_comment").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Comment</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 20px; margin: 20px 0; border-left: 4px solid #4CAF50; }
        .comment { background-color: white; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .author { font-weight: bold; color: #4CAF50; }
        .footer { text-align: center; color: #777; font-size: 12px; padding: 20px; }
        .button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>New Comment on {{ .SiteName }}</h1>
    </div>
    <div class="content">
        <p>A new comment has been posted on <strong>{{ .PageTitle }}</strong>:</p>
        <div class="comment">
            <p class="author">{{ .AuthorName }}</p>
            <p>{{ .CommentText }}</p>
        </div>
        <a href="{{ .CommentURL }}" class="button">View Comment</a>
    </div>
    <div class="footer">
        <p>You're receiving this because you're the owner of {{ .SiteName }}.</p>
        <p><a href="{{ .UnsubscribeURL }}">Unsubscribe</a> from these notifications</p>
    </div>
</body>
</html>
`))

	// Comment reply template
	template.Must(tmpl.New("comment_reply").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Reply</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 20px; margin: 20px 0; border-left: 4px solid #2196F3; }
        .comment { background-color: white; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .original { border-left: 3px solid #ddd; padding-left: 15px; margin: 10px 0; color: #666; }
        .author { font-weight: bold; color: #2196F3; }
        .footer { text-align: center; color: #777; font-size: 12px; padding: 20px; }
        .button { display: inline-block; padding: 10px 20px; background-color: #2196F3; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Someone Replied to Your Comment</h1>
    </div>
    <div class="content">
        <p><strong>{{ .AuthorName }}</strong> replied to your comment on <strong>{{ .PageTitle }}</strong>:</p>
        <div class="comment">
            <p>{{ .ReplyText }}</p>
        </div>
        <div class="original">
            <p><em>Your original comment:</em></p>
            <p>{{ .OriginalText }}</p>
        </div>
        <a href="{{ .CommentURL }}" class="button">View Reply</a>
    </div>
    <div class="footer">
        <p>You're receiving this because someone replied to your comment.</p>
        <p><a href="{{ .UnsubscribeURL }}">Unsubscribe</a> from these notifications</p>
    </div>
</body>
</html>
`))

	// Moderation update template
	template.Must(tmpl.New("moderation_update").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Comment Moderation Update</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #FF9800; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 20px; margin: 20px 0; border-left: 4px solid #FF9800; }
        .comment { background-color: white; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .status { font-weight: bold; padding: 5px 10px; border-radius: 3px; display: inline-block; }
        .status.approved { background-color: #4CAF50; color: white; }
        .status.rejected { background-color: #f44336; color: white; }
        .footer { text-align: center; color: #777; font-size: 12px; padding: 20px; }
        .button { display: inline-block; padding: 10px 20px; background-color: #FF9800; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Comment Moderation Update</h1>
    </div>
    <div class="content">
        <p>Your comment on <strong>{{ .PageTitle }}</strong> has been:</p>
        <p><span class="status {{ .Status }}">{{ .Status }}</span></p>
        <div class="comment">
            <p>{{ .CommentText }}</p>
        </div>
        {{ if .Reason }}
        <p><strong>Reason:</strong> {{ .Reason }}</p>
        {{ end }}
        <a href="{{ .CommentURL }}" class="button">View Comment</a>
    </div>
    <div class="footer">
        <p>You're receiving this because this is your comment.</p>
        <p><a href="{{ .UnsubscribeURL }}">Unsubscribe</a> from these notifications</p>
    </div>
</body>
</html>
`))

	return &EmailTemplate{templates: tmpl}
}

// RenderNewComment renders the new comment email template
func (e *EmailTemplate) RenderNewComment(data map[string]string) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, "new_comment", data); err != nil {
		return "", fmt.Errorf("failed to render new_comment template: %w", err)
	}
	return buf.String(), nil
}

// RenderCommentReply renders the comment reply email template
func (e *EmailTemplate) RenderCommentReply(data map[string]string) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, "comment_reply", data); err != nil {
		return "", fmt.Errorf("failed to render comment_reply template: %w", err)
	}
	return buf.String(), nil
}

// RenderModerationUpdate renders the moderation update email template
func (e *EmailTemplate) RenderModerationUpdate(data map[string]string) (string, error) {
	var buf bytes.Buffer
	if err := e.templates.ExecuteTemplate(&buf, "moderation_update", data); err != nil {
		return "", fmt.Errorf("failed to render moderation_update template: %w", err)
	}
	return buf.String(), nil
}
