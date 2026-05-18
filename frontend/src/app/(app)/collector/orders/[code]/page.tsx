import { notFound } from 'next/navigation';
import { api, ApiError } from '@/lib/api';
import type { Order } from '@/types/api';
import CollectorOrderDetail from './CollectorOrderDetail';

interface Props {
  params: Promise<{ code: string }>;
}

export default async function CollectorOrderDetailPage({ params }: Props) {
  const { code } = await params;

  let order: Order;
  try {
    order = await api<Order>(`/v1/collector/orders/${code}`);
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) notFound();
    throw e;
  }

  return <CollectorOrderDetail order={order} />;
}
