import { notFound } from 'next/navigation';
import { api } from '@/lib/api';
import type { Order } from '@/types/api';
import OrderDetail from './OrderDetail';

interface Props {
  params: Promise<{ code: string }>;
}

export default async function OrderDetailPage({ params }: Props) {
  const { code } = await params;

  let order: Order;
  try {
    order = await api<Order>(`/v1/orders/${code}`);
  } catch {
    notFound();
  }

  return <OrderDetail initialOrder={order} />;
}
