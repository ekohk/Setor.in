package email

import (
	"strings"
	"testing"
)

// TestEachTemplateRendersUniqueContent guards against the "all emails look the
// same" bug — caused when multiple templates declare the same `{{define ...}}`
// name and one overwrites the others in the global namespace.
//
// We render every event template and assert each output contains a string
// that ONLY appears in that specific template's body. If two templates render
// identical-looking output, this test will fail and point to the culprit.
func TestEachTemplateRendersUniqueContent(t *testing.T) {
	cases := []struct {
		name        string
		data        map[string]any
		mustContain string // text that should appear ONLY in this template
	}{
		{
			name: "welcome",
			data: map[string]any{"FullName": "Tika"},
			// Distinguishing copy: the welcome onboarding heading
			mustContain: "Akun Anda aktif",
		},
		{
			name:        "email_verified",
			data:        map[string]any{"FullName": "Tika", "Email": "tika@test.local"},
			mustContain: "Email Anda terverifikasi",
		},
		{
			name:        "password_changed",
			data:        map[string]any{"FullName": "Tika", "ChangedAt": "2026-05-11"},
			mustContain: "Password berhasil diubah",
		},
		{
			name:        "password_reset_requested",
			data:        map[string]any{"FullName": "Tika", "ResetURL": "https://reset.example/abc"},
			mustContain: "Reset password",
		},
		{
			name:        "login_new_device",
			data:        map[string]any{"FullName": "Tika", "LoginAt": "2026-05-11"},
			mustContain: "Login dari perangkat baru",
		},
		{
			name: "collector_submitted",
			data: map[string]any{
				"FullName":     "Tika",
				"BusinessName": "Tika Recycling",
				"Address":      "Jl. Mawar 1",
			},
			mustContain: "Pengajuan collector terkirim",
		},
		{
			name: "collector_approved",
			data: map[string]any{
				"FullName":     "Tika",
				"BusinessName": "Tika Recycling",
			},
			mustContain: "Anda kini collector terverifikasi",
		},
		{
			name: "collector_rejected",
			data: map[string]any{
				"FullName":     "Tika",
				"BusinessName": "Tika Recycling",
				"Reason":       "KTP buram",
			},
			mustContain: "Aplikasi collector belum bisa kami setujui",
		},
		{
			name:        "account_suspended",
			data:        map[string]any{"FullName": "Tika", "Reason": "Investigasi"},
			mustContain: "Akun dinonaktifkan sementara",
		},
		{
			name:        "account_unsuspended",
			data:        map[string]any{"FullName": "Tika"},
			mustContain: "Akun aktif kembali",
		},
		{
			name: "role_changed",
			data: map[string]any{
				"FullName": "Tika",
				"OldRole":  "user",
				"NewRole":  "admin",
			},
			mustContain: "Peran akun diperbarui",
		},
		{
			name: "order_created",
			data: map[string]any{
				"FullName": "Tika", "OrderCode": "ECC-01001",
				"MaterialName": "Plastik", "EstimatedPayout": int64(15000),
			},
			mustContain: "Order baru kami terima",
		},
		{
			name: "order_accepted",
			data: map[string]any{
				"FullName": "Tika", "OrderCode": "ECC-01001", "OTPCode": "4729",
			},
			mustContain: "Collector siap menerima order Anda",
		},
		{
			name:        "order_enroute",
			data:        map[string]any{"FullName": "Tika", "OrderCode": "ECC-01001"},
			mustContain: "Driver sedang menuju lokasi",
		},
		{
			name:        "order_arrived",
			data:        map[string]any{"FullName": "Tika", "OrderCode": "ECC-01001"},
			mustContain: "Driver sudah tiba",
		},
		{
			name: "order_completed",
			data: map[string]any{
				"FullName": "Tika", "OrderCode": "ECC-01001", "FinalPayout": int64(20000),
			},
			mustContain: "Terima kasih telah daur ulang",
		},
		{
			name: "order_cancelled",
			data: map[string]any{
				"FullName": "Tika", "OrderCode": "ECC-01001", "Reason": "berubah pikiran",
			},
			mustContain: "Order dibatalkan",
		},
	}

	// Cross-check: collect outputs and assert all are mutually distinct.
	outputs := make(map[string]string, len(cases))

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			data := withBranding("AKUN", false, c.data)
			data["Subject"] = "test"
			out, err := render(c.name, data)
			if err != nil {
				t.Fatalf("render %q: %v", c.name, err)
			}
			if !strings.Contains(out, c.mustContain) {
				t.Errorf("template %q missing distinguishing text %q", c.name, c.mustContain)
			}
			outputs[c.name] = out
		})
	}

	// Final cross-check: extract the BODY (between header & footer markers)
	// and assert no two templates produced identical bodies.
	t.Run("all bodies distinct", func(t *testing.T) {
		seen := make(map[string]string)
		for name, full := range outputs {
			body := bodyOnly(full)
			if other, dup := seen[body]; dup {
				t.Errorf("templates %q and %q produced identical body content", name, other)
			}
			seen[body] = name
		}
	})
}

// bodyOnly strips the shared header (everything before the first <h1) and
// shared footer (everything after the last </h1...><table or "Email ini
// dikirim otomatis"), leaving the unique middle.
func bodyOnly(full string) string {
	const startMark = "<h1"
	start := strings.Index(full, startMark)
	if start < 0 {
		return full
	}
	const endMark = "Email ini dikirim otomatis"
	end := strings.Index(full, endMark)
	if end < 0 {
		end = len(full)
	}
	return full[start:end]
}
