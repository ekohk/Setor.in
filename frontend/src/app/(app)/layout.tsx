import { auth } from '@/lib/auth';
import BottomNav from '@/components/ui/BottomNav';

export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const session = await auth();
  const roles = (session?.user as Record<string, unknown> | undefined)?.roles as string[] ?? [];

  return (
    <div className="flex justify-center min-h-screen bg-paper">
      <div className="w-full max-w-md flex flex-col relative">
        <main className="flex-1 pb-[72px]">
          {children}
        </main>
        <BottomNav roles={roles} />
      </div>
    </div>
  );
}
