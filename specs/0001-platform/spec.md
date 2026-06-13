# Spec 0001 — Platform Overview (Setor.in)

**Status:** Living document
**Last updated:** 2026-05-05
**Owner:** @ekohk

## 1. Problem

Indonesia menghasilkan ~64 juta ton sampah/tahun, ~15% adalah anorganik bernilai
ekonomi (plastik, logam, kertas, e-waste). Saat ini:

- Konsumen rumah tangga **tidak tahu kemana menjual** sampah daur ulang dengan
  harga adil.
- Pengepul informal **tidak transparan** soal harga & berat → konsumen sering rugi.
- Bank sampah pemerintah & startup eksisting **UX-nya kompleks**, jam terbatas,
  cakupan sempit.
- **Tidak ada data terpusat** untuk pemerintah/industri daur ulang memetakan
  supply.

## 2. Solusi

**Setor.in** — marketplace mobile-first yang menghubungkan penjual sampah daur
ulang (rumah tangga) dengan pengepul tersertifikasi, dengan:

- Estimasi harga marketplace per material (rentang harga per wilayah)
- Verifikasi digital: OTP serah terima + live weighing + inspection result + foto
- Cash-first transaction flow (wallet/payment gateway di phase berikutnya)
- Satu app untuk seller, collector, dan CV partner

## 3. Goals (12 bulan)

| # | Goal | Metric |
|---|---|---|
| G1 | PMF di 1 kota (Jakarta Selatan) | 1.000 active sellers/bulan, 30% retention bulanan |
| G2 | Onboard collector tervalidasi | 50 collector verified |
| G3 | Volume transaksi | 10 ton/bulan total dari semua material |
| G4 | Trust tinggi | <2% dispute rate, NPS ≥ 50 |
| G5 | Unit economics positif | Commission cover operasional di 1 kota |

## 4. Non-Goals (v1)

- ❌ Multi-bahasa (ID-only di MVP)
- ❌ Native iOS/Android — PWA dulu
- ❌ Social login (email/phone OTP cukup)
- ❌ Live chat in-app — pakai WhatsApp deeplink
- ❌ Real-time GPS tracking driver — ETA static cukup
- ❌ AI weight estimation — phase 3
- ❌ Marketplace barang preloved (out of scope total)

## 5. Personas

### Aria — Urban household (25–40)
Tinggal di apartemen Kemang, punya botol PET & kardus menumpuk tiap minggu.
Mau ekstra pemasukan tapi malas bawa ke pengepul. Punya GoPay & rekening BCA.

### Budi — Independent collector (30–55)
Punya motor + timbangan + kontrak ke pabrik daur ulang. Ambil 5–15 order/hari.
Butuh dashboard simpel, bukan aplikasi desktop kompleks.

### Ibu Sari — RT/PKK coordinator (40+)
Kumpul sampah skala blok (10–30 KK), jual collective tiap 2 minggu.
Butuh fitur multi-source dalam 1 order.

### Admin internal Setor.in
Role operasional internal untuk verifikasi, moderasi, dan maintenance platform.
Tidak menjadi persona produk utama untuk phase 1.

## 6. Phase 1 Core Operational Contract

### Roles

- `user` — upload foto barang, input estimasi berat, request pickup, setuju/tolak harga final.
- `cv` — partner penerima material dari collector; input berat diterima, nominal pembelian, dan catatan kualitas.
- `collector` — terima order, pickup, timbang ulang, tetapkan harga final per kg + total, isi catatan kondisi.
- `super_admin` — operasional internal: suspend akun, promote role, monitoring, recovery.

### API surface minimum

- Public catalog: `GET /v1/materials`, `GET /v1/materials/:slug`
- Auth sync/profile: `POST /v1/auth/sync`, `GET /v1/auth/me`, `GET /v1/users/me`, `PATCH /v1/users/me`
- Seller flow: `POST /v1/orders`, `GET /v1/orders/me`, `GET /v1/orders/:code`, `POST /v1/orders/:code/cancel`, `POST /v1/orders/:code/confirm-cash`
- CV flow: dashboard/incoming-material endpoints akan dipisah dari seller flow ketika task phase 2 dikerjakan
- Collector flow: `GET /v1/collector/orders/incoming`, `GET /v1/collector/orders/me`, `GET /v1/collector/orders/:code`, `POST /v1/collector/orders/:code/accept`, `POST /v1/collector/orders/:code/start-pickup`, `POST /v1/collector/orders/:code/arrive`, `POST /v1/collector/orders/:code/verify-otp`, `POST /v1/collector/orders/:code/weigh`, `POST /v1/collector/orders/:code/inspection-result`
- Internal admin: user moderation, collector application review, material pricing admin, and RBAC smoke test

### DB anchors

- `users`, `user_role_history` for identity and role change audit
- `auth_events` for auth/audit trail
- `collector_applications` for approval flow
- `materials`, `material_price_estimates` for catalog and market-based estimation history
- `orders`, `order_status_history`, `order_photos` for pickup and cash transaction flow

### Rule of thumb

Jika sebuah endpoint belum dipakai oleh phase 1 user/collector flow, jangan dulu dibawa ke kontrak publik. Simpan sebagai internal/admin atau phase 2+.

## 6. Sukses kriteria platform

- 95% order selesai dalam <24 jam (untuk pickup)
- 99% order final offer mendapat respons user (setuju/tolak) dalam <2 jam setelah inspection result
- 0 incident kebocoran data PII
- Withdrawal sukses rate >98%

## 7. Stakeholder

- **Owner / CTO:** @ekohk
- **Product:** TBD
- **Backend:** @ekohk
- **Frontend:** TBD
- **Designer:** Claude Design (handoff bundle di [project/](../../project/))

## 8. Glossary singkat

| Term | Arti |
|---|---|
| **User / Seller** | Orang yang menjual sampah |
| **Collector** | Pengepul yang membeli & memproses |
| **Order** | Transaksi dari pembuatan sampai payout |
| **Pickup** | Method: collector datang ke rumah user |
| **Drop-off** | Method: user antar ke lokasi collector (+5% bonus) |
| **Wallet** | Saldo internal user, bisa di-withdraw |
| **Material** | Jenis bahan (8 kategori: plastic, cardboard, paper, aluminum, copper, steel, glass, e-waste) |
| **Inspection Result** | Hasil penilaian collector (berat aktual, harga per kg, total harga, catatan kondisi) |
| **Payout** | Uang yang diterima user setelah menyetujui harga final |
| **OTP** | Kode 4-digit untuk verifikasi serah terima |

## 9. Cross-references

- Design mock: [project/EcoCycle.html](../../project/EcoCycle.html)
- Architecture: [adr/0002-hexagonal-architecture.md](../adr/0002-hexagonal-architecture.md)
- First feature: [0002-auth/spec.md](../0002-auth/spec.md)
