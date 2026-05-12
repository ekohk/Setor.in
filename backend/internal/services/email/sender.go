// Package email is the shared transactional-email service for Setor.in.
//
// One Service instance is constructed in main.go and injected into any usecase
// that needs to notify users. All sends are async: the public methods return
// immediately and the actual SMTP call runs in a goroutine, so HTTP requests
// are not slowed down by mail server latency.
//
// Templates are rendered from `templates/*.html` (embedded at build time).
// The `base` template provides the shell (header, footer, brand styles); each
// event template defines `content` and includes `base`.
package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	gomail "github.com/wneessen/go-mail"

	"github.com/setorin/setorin/backend/internal/shared/config"
)

// Service sends transactional email. Safe for concurrent use.
type Service struct {
	cfg    config.SMTPConfig
	client *gomail.Client
	dryRun bool // true if SMTP_HOST is empty — logs only, never connects
}

// NewService builds the sender. Returns ErrDisabled if SMTP_HOST is unset
// (caller should treat the service as a no-op in that case).
var ErrDisabled = errors.New("email: disabled (SMTP_HOST not configured)")

func NewService(cfg config.SMTPConfig) (*Service, error) {
	if cfg.Host == "" {
		slog.Warn("email service running in dry-run mode (SMTP_HOST empty); will log instead of send")
		return &Service{cfg: cfg, dryRun: true}, nil
	}

	opts := []gomail.Option{
		gomail.WithPort(cfg.Port),
		gomail.WithTimeout(10 * time.Second),
	}

	// Local catchers like MailHog don't support STARTTLS / auth.
	// Production providers (Resend, SendGrid) need auth + TLS.
	if cfg.Username == "" && cfg.Password == "" {
		opts = append(opts, gomail.WithTLSPolicy(gomail.NoTLS), gomail.WithSMTPAuth(gomail.SMTPAuthNoAuth))
	} else {
		opts = append(opts,
			gomail.WithTLSPolicy(gomail.TLSMandatory),
			gomail.WithSMTPAuth(gomail.SMTPAuthPlain),
			gomail.WithUsername(cfg.Username),
			gomail.WithPassword(cfg.Password),
		)
	}

	client, err := gomail.NewClient(cfg.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("email: build client: %w", err)
	}

	slog.Info("email service ready", "host", cfg.Host, "port", cfg.Port, "from", cfg.From)
	return &Service{cfg: cfg, client: client}, nil
}

// send is the lowest-level helper — synchronous SMTP send.
// All Send* helpers wrap this in a goroutine.
func (s *Service) send(ctx context.Context, to, subject, htmlBody string) error {
	if s.dryRun {
		slog.Info("email (dry-run)", "to", to, "subject", subject, "body_len", len(htmlBody))
		return nil
	}

	m := gomail.NewMsg()
	if err := m.FromFormat(s.cfg.FromName, s.cfg.From); err != nil {
		return fmt.Errorf("set From: %w", err)
	}
	if err := m.To(to); err != nil {
		return fmt.Errorf("set To: %w", err)
	}
	m.Subject(subject)
	m.SetBodyString(gomail.TypeTextHTML, htmlBody)

	if err := s.client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

// sendAsync renders + sends in a goroutine; logs but does not return errors.
// Use for fire-and-forget notifications where business flow shouldn't block
// or fail because of mail server hiccups.
//
// Caller is expected to populate `data` with branding fields (AccentColor,
// BadgeBg, BadgeColor, SecurityFooter) via the helpers in notifications.go.
// We always set Subject + Category here so individual templates don't have to.
func (s *Service) sendAsync(to, subject, templateName, category string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	data["Subject"] = subject
	data["Category"] = category

	body, err := render(templateName, data)
	if err != nil {
		slog.Error("email render failed", "template", templateName, "err", err, "to", to)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.send(ctx, to, subject, body); err != nil {
			slog.Error("email send failed", "template", templateName, "to", to, "err", err)
			return
		}
		slog.Info("email sent", "template", templateName, "to", redactEmail(to))
	}()
}

func redactEmail(e string) string {
	for i, c := range e {
		if c == '@' {
			if i <= 1 {
				return "***" + e[i:]
			}
			return e[:1] + "***" + e[i:]
		}
	}
	return "***"
}
