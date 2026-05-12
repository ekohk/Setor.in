package email

// Public notification methods. Each maps 1:1 to an event in the auth flow.
// Categories drive header badge color & accent stripe so each email type has
// distinct visual identity:
//
//   AKUN     → green (default brand) — welcome, verified, unsuspended
//   KEAMANAN → blue                  — password / security events
//   PERHATIAN → orange               — reset / new-device alerts
//   COLLECTOR → green muted          — application events
//   PEMBLOKIRAN → red                — suspended, rejected
//   PERAN    → purple                — role mutations

// palette returns (accentColor, badgeBg, badgeColor) per category.
func palette(category string) (string, string, string) {
	switch category {
	case "KEAMANAN":
		return "#1e5a96", "#e6f0fb", "#1e5a96"
	case "PERHATIAN":
		return "#a05a2c", "#fdf2e9", "#a05a2c"
	case "PEMBLOKIRAN":
		return "#9c2a1f", "#fbe9e7", "#9c2a1f"
	case "PERAN":
		return "#5b3aa6", "#efeafa", "#5b3aa6"
	case "COLLECTOR":
		return "#2f7d52", "#eaf3ec", "#1f5d3a"
	case "ORDER":
		// teal — distinct from green (akun/collector) and red (pemblokiran)
		return "#0f766e", "#ccfbf1", "#0f766e"
	case "AKUN":
		fallthrough
	default:
		return "#2f7d52", "#eaf3ec", "#1f5d3a"
	}
}

// withBranding adds standard fields all templates need (palette + flags).
// `data` may be nil.
func withBranding(category string, securityFooter bool, data map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	accent, badgeBg, badgeColor := palette(category)
	data["AccentColor"] = accent
	data["BadgeBg"] = badgeBg
	data["BadgeColor"] = badgeColor
	data["SecurityFooter"] = securityFooter
	return data
}

// ────────── Onboarding ──────────

// IMPORTANT: the templateName arg passed to sendAsync below is the name
// declared by the matching {{define "<name>"}} block — NOT the filename.
// All templates are parsed into one global namespace, so each must use a
// unique define name (e.g. "welcome", "collector_approved"). If two
// templates declare the same name, only one wins → all emails look identical.

// Welcome — sent ONLY on first sync (when local DB row is created).
// NOT sent on subsequent logins. Idempotency lives in user_usecase.go.
func (s *Service) Welcome(to, fullName string) {
	s.sendAsync(to, "Selamat datang di Setor.in", "welcome", "AKUN",
		withBranding("AKUN", false, map[string]any{
			"FullName": fullName,
		}),
	)
}

func (s *Service) EmailVerified(to, fullName, email string) {
	s.sendAsync(to, "Email Anda terverifikasi", "email_verified", "AKUN",
		withBranding("AKUN", false, map[string]any{
			"FullName": fullName,
			"Email":    email,
		}),
	)
}

// ────────── Security ──────────

func (s *Service) PasswordChanged(to, fullName, changedAt, ip string) {
	s.sendAsync(to, "Password akun Anda berhasil diubah", "password_changed", "KEAMANAN",
		withBranding("KEAMANAN", true, map[string]any{
			"FullName":  fullName,
			"ChangedAt": changedAt,
			"IPAddress": ip,
		}),
	)
}

func (s *Service) PasswordResetRequested(to, fullName, resetURL string) {
	s.sendAsync(to, "Permintaan reset password", "password_reset_requested", "PERHATIAN",
		withBranding("PERHATIAN", true, map[string]any{
			"FullName": fullName,
			"ResetURL": resetURL,
		}),
	)
}

func (s *Service) LoginNewDevice(to, fullName, loginAt, device, ip, location string) {
	s.sendAsync(to, "Login baru terdeteksi di akun Anda", "login_new_device", "PERHATIAN",
		withBranding("PERHATIAN", true, map[string]any{
			"FullName":  fullName,
			"LoginAt":   loginAt,
			"Device":    device,
			"IPAddress": ip,
			"Location":  location,
		}),
	)
}

// ────────── Collector application ──────────

func (s *Service) CollectorApplicationSubmitted(to, fullName, businessName, address, licenseNo string) {
	s.sendAsync(to, "Pengajuan collector terkirim", "collector_submitted", "COLLECTOR",
		withBranding("COLLECTOR", false, map[string]any{
			"FullName":     fullName,
			"BusinessName": businessName,
			"Address":      address,
			"LicenseNo":    licenseNo,
		}),
	)
}

func (s *Service) CollectorApplicationApproved(to, fullName, businessName string) {
	s.sendAsync(to, "Anda kini collector terverifikasi", "collector_approved", "COLLECTOR",
		withBranding("COLLECTOR", false, map[string]any{
			"FullName":     fullName,
			"BusinessName": businessName,
		}),
	)
}

func (s *Service) CollectorApplicationRejected(to, fullName, businessName, reason string) {
	s.sendAsync(to, "Pengajuan collector belum disetujui", "collector_rejected", "PEMBLOKIRAN",
		withBranding("PEMBLOKIRAN", false, map[string]any{
			"FullName":     fullName,
			"BusinessName": businessName,
			"Reason":       reason,
		}),
	)
}

// ────────── Account moderation ──────────

func (s *Service) AccountSuspended(to, fullName, reason string) {
	s.sendAsync(to, "Akun Anda dinonaktifkan sementara", "account_suspended", "PEMBLOKIRAN",
		withBranding("PEMBLOKIRAN", false, map[string]any{
			"FullName": fullName,
			"Reason":   reason,
		}),
	)
}

func (s *Service) AccountUnsuspended(to, fullName string) {
	s.sendAsync(to, "Akun Anda diaktifkan kembali", "account_unsuspended", "AKUN",
		withBranding("AKUN", false, map[string]any{
			"FullName": fullName,
		}),
	)
}

// ────────── Order lifecycle (cash flow) ──────────

func (s *Service) OrderCreated(to, fullName, orderCode, materialName string, estimatedPayout int64) {
	s.sendAsync(to, "Order Anda dibuat — "+orderCode, "order_created", "ORDER",
		withBranding("ORDER", false, map[string]any{
			"FullName":        fullName,
			"OrderCode":       orderCode,
			"MaterialName":    materialName,
			"EstimatedPayout": estimatedPayout,
		}),
	)
}

func (s *Service) OrderAccepted(to, fullName, orderCode, otpCode string) {
	s.sendAsync(to, "Collector menerima order Anda — "+orderCode, "order_accepted", "ORDER",
		withBranding("ORDER", false, map[string]any{
			"FullName":  fullName,
			"OrderCode": orderCode,
			"OTPCode":   otpCode,
		}),
	)
}

func (s *Service) OrderEnroute(to, fullName, orderCode string) {
	s.sendAsync(to, "Driver dalam perjalanan — "+orderCode, "order_enroute", "ORDER",
		withBranding("ORDER", false, map[string]any{
			"FullName":  fullName,
			"OrderCode": orderCode,
		}),
	)
}

func (s *Service) OrderArrived(to, fullName, orderCode string) {
	s.sendAsync(to, "Driver sudah tiba — "+orderCode, "order_arrived", "ORDER",
		withBranding("ORDER", false, map[string]any{
			"FullName":  fullName,
			"OrderCode": orderCode,
		}),
	)
}

func (s *Service) OrderCompleted(to, fullName, orderCode string, finalPayout int64) {
	s.sendAsync(to, "Order selesai — "+orderCode, "order_completed", "ORDER",
		withBranding("ORDER", false, map[string]any{
			"FullName":    fullName,
			"OrderCode":   orderCode,
			"FinalPayout": finalPayout,
		}),
	)
}

func (s *Service) OrderCancelled(to, fullName, orderCode, reason string) {
	s.sendAsync(to, "Order dibatalkan — "+orderCode, "order_cancelled", "PEMBLOKIRAN",
		withBranding("PEMBLOKIRAN", false, map[string]any{
			"FullName":  fullName,
			"OrderCode": orderCode,
			"Reason":    reason,
		}),
	)
}

// ────────── Role / privileges ──────────

func (s *Service) RoleChanged(to, fullName, oldRole, newRole, reason string) {
	s.sendAsync(to, "Peran akun Anda diperbarui", "role_changed", "PERAN",
		withBranding("PERAN", false, map[string]any{
			"FullName": fullName,
			"OldRole":  oldRole,
			"NewRole":  newRole,
			"Reason":   reason,
		}),
	)
}
