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

- Harga transparan per material per kg (real-time)
- Verifikasi digital: OTP serah terima + live weighing + quality grading + foto
- Wallet internal yang bisa di-withdraw ke bank/e-wallet
- Satu app untuk seller & collector (mode toggle)

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
Update harga harian, resolve dispute, verify collector application.

## 6. Sukses kriteria platform

- 95% order selesai dalam <24 jam (untuk pickup)
- 99% transaksi wallet ter-credit dalam <60 detik setelah quality check
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
| **Payout** | Uang yang diterima user setelah quality check |
| **OTP** | Kode 4-digit untuk verifikasi serah terima |

## 9. Cross-references

- Design mock: [project/EcoCycle.html](../../project/EcoCycle.html)
- Architecture: [adr/0002-hexagonal-architecture.md](../adr/0002-hexagonal-architecture.md)
- First feature: [0002-auth/spec.md](../0002-auth/spec.md)
