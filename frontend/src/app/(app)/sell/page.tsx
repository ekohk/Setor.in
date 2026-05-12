import { getMaterials } from '@/lib/api/materials';
import SellForm from './SellForm';

export default async function SellPage() {
  let materials = await getMaterials().catch(() => []);
  // Sort by sort_order asc
  materials = materials.sort((a, b) => a.sort_order - b.sort_order);

  return <SellForm materials={materials} />;
}
