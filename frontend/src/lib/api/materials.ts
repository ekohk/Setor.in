import { api } from '@/lib/api';
import type { Material } from '@/types/api';

export async function getMaterials(): Promise<Material[]> {
  return api<Material[]>('/v1/materials', { noAuth: true });
}

export async function getMaterialBySlug(slug: string): Promise<Material> {
  return api<Material>(`/v1/materials/${slug}`, { noAuth: true });
}
