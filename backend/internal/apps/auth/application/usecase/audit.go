package usecase

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/setorin/setorin/backend/internal/apps/auth/application/ports"
	"github.com/setorin/setorin/backend/internal/apps/auth/domain/model"
)

// AuditService is a thin convenience wrapper over the AuthEventRepository.
//
// Failures to write an audit row are logged but never returned to the caller,
// so business actions don't fail when audit storage is degraded. Trade-off:
// audit gaps possible during DB outage.
type AuditService struct {
	repo ports.AuthEventRepository
}

func NewAuditService(repo ports.AuthEventRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Record writes an event. ip + ua + metadata may be empty.
// userID may be nil for events that occur before sync (e.g. failed_auth).
func (s *AuditService) Record(
	ctx context.Context,
	eventType model.AuthEventType,
	userID *uuid.UUID,
	keycloakID *uuid.UUID,
	ip, ua string,
	metadata map[string]any,
) {
	var raw []byte
	if len(metadata) > 0 {
		b, err := json.Marshal(metadata)
		if err != nil {
			slog.Warn("audit: marshal metadata failed", "err", err, "type", eventType)
			// Continue — better to log without metadata than not at all.
		} else {
			raw = b
		}
	}

	e := &model.AuthEvent{
		UserID:     userID,
		KeycloakID: keycloakID,
		EventType:  eventType,
		Metadata:   raw,
	}
	if ip != "" {
		e.IPAddress = &ip
	}
	if ua != "" {
		e.UserAgent = &ua
	}

	if err := s.repo.Insert(ctx, e); err != nil {
		slog.Warn("audit: insert failed",
			"err", err,
			"type", eventType,
			"user_id", userID,
		)
	}
}
