// Shared API response types — mirrors backend DTOs exactly.
// Keep in sync with backend/internal/apps/*/application/dto/*.go

// ─── User ───────────────────────────────────────────────────────────────────

export interface UserProfile {
  id: string;
  keycloak_id: string;
  email: string;
  phone?: string;
  full_name: string;
  avatar_url?: string;
  primary_role: 'user' | 'collector' | 'admin' | 'super_admin';
  status: 'pending_verification' | 'active' | 'suspended' | 'deleted';
  email_verified: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
  roles?: string[]; // from JWT — present in /auth/me response
}

export interface UpdateProfilePayload {
  full_name?: string;
  phone?: string;
}

// ─── Catalog ────────────────────────────────────────────────────────────────

export interface Material {
  id: string;
  slug: string;
  name: string;
  icon?: string;
  unit: string;
  description?: string;
  is_active: boolean;
  sort_order: number;
  current_price?: number;   // rupiah/kg, nil if not yet set
  price_valid_from?: string; // ISO timestamp
}

// ─── Order ──────────────────────────────────────────────────────────────────

export type OrderStatus =
  | 'received'
  | 'accepted'
  | 'enroute'
  | 'arrived'
  | 'weighing'
  | 'quality'
  | 'cash_handover'
  | 'done'
  | 'cancelled'
  | 'disputed';

export type OrderMethod = 'pickup' | 'dropoff';

export interface Order {
  id: string;
  order_code: string;
  user_id: string;
  collector_id?: string;
  material_id: string;

  estimated_weight_kg: string;
  actual_weight_kg?: string;

  unit_price_at_order: number;
  estimated_payout: number;
  final_payout?: number;

  quality_grade?: string;
  quality_bonus_pct: number;

  method: OrderMethod;
  payment_method: string;
  payment_status: string;
  paid_at?: string;

  status: OrderStatus;
  otp_code?: string;
  otp_expires_at?: string;

  address_text: string;
  latitude?: number;
  longitude?: number;
  notes?: string;

  created_at: string;
  accepted_at?: string;
  arrived_at?: string;
  completed_at?: string;
  cancelled_at?: string;
  cancellation_reason?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  page: number;
  page_size: number;
  total: number;
}

// ─── Order status helpers ────────────────────────────────────────────────────

export const ORDER_STATUS_LABEL: Record<OrderStatus, string> = {
  received:      'Menunggu Collector',
  accepted:      'Diterima Collector',
  enroute:       'Dalam Perjalanan',
  arrived:       'Driver Tiba',
  weighing:      'Penimbangan',
  quality:       'Cek Kualitas',
  cash_handover: 'Serah Terima Uang',
  done:          'Selesai',
  cancelled:     'Dibatalkan',
  disputed:      'Dalam Sengketa',
};

export const ORDER_STATUS_STEP: Record<OrderStatus, number> = {
  received:      0,
  accepted:      1,
  enroute:       2,
  arrived:       3,
  weighing:      4,
  quality:       5,
  cash_handover: 6,
  done:          7,
  cancelled:    -1,
  disputed:     -1,
};

export function isActiveOrder(status: OrderStatus): boolean {
  return !['done', 'cancelled', 'disputed'].includes(status);
}

// ─── Admin ───────────────────────────────────────────────────────────────────

export type UserStatus = 'pending_verification' | 'active' | 'suspended' | 'deleted';
export type UserRole = 'user' | 'collector' | 'admin' | 'super_admin';
export type ApplicationStatus = 'pending' | 'approved' | 'rejected';

export interface AdminUser {
  id: string;
  keycloak_id: string;
  email: string;
  phone?: string;
  full_name: string;
  primary_role: UserRole;
  status: UserStatus;
  email_verified: boolean;
  last_login_at?: string;
  created_at: string;
}

export interface CollectorApplication {
  id: string;
  user_id: string;
  business_name: string;
  license_no?: string;
  ktp_url: string;
  siup_url?: string;
  address: string;
  status: ApplicationStatus;
  rejection_reason?: string;
  reviewed_by?: string;
  submitted_at: string;
  reviewed_at?: string;
}

export const USER_STATUS_LABEL: Record<UserStatus, string> = {
  pending_verification: 'Belum Verifikasi',
  active:               'Aktif',
  suspended:            'Disuspend',
  deleted:              'Dihapus',
};

export const USER_ROLE_LABEL: Record<UserRole, string> = {
  user:        'Pengguna',
  collector:   'Collector',
  admin:       'Admin',
  super_admin: 'Super Admin',
};

export const APPLICATION_STATUS_LABEL: Record<ApplicationStatus, string> = {
  pending:  'Menunggu Review',
  approved: 'Disetujui',
  rejected: 'Ditolak',
};
