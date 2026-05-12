import BottomNav from '@/components/ui/BottomNav';

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex justify-center min-h-screen bg-paper">
      <div className="w-full max-w-md flex flex-col relative">
        <main className="flex-1 pb-[72px]">
          {children}
        </main>
        <BottomNav />
      </div>
    </div>
  );
}
