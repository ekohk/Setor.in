import { auth } from '@/lib/auth';
import { redirect } from 'next/navigation';

export default async function CvPage() {
  const session = await auth();
  if (!session) redirect('/login');

  const name = session.user?.name?.split(' ')[0] ?? 'Partner';

  return (
    <main className="min-h-screen animate-pageIn">
      <div className="bg-emerald-600 px-6 pt-12 pb-8">
        <p className="text-emerald-50/80 text-sm font-medium">CV Dashboard</p>
        <h1 className="text-white text-2xl font-bold tracking-tight mt-0.5">Selamat datang, {name}</h1>
        <p className="text-emerald-50/80 text-sm mt-2 max-w-sm">
          Ruang kerja untuk partner CV yang menerima material dari collector dan menindaklanjuti proses berikutnya.
        </p>
      </div>

      <div className="px-4 py-5 space-y-4">
        <section className="glass px-4 py-4 space-y-2">
          <p className="text-[12px] font-bold text-ink-3 uppercase tracking-wider">Status Role</p>
          <p className="text-[14px] text-ink leading-relaxed">
            Role `cv` sekarang diperlakukan sebagai role terpisah, bukan alias seller user. Dashboard ini menjadi titik awal untuk alur penerimaan dari collector.
          </p>
        </section>

        <section className="glass px-4 py-4 space-y-3">
          <p className="text-[12px] font-bold text-ink-3 uppercase tracking-wider">Langkah Berikutnya</p>
          <ul className="space-y-2 text-[13px] text-ink-2 leading-relaxed">
            <li>1. Tambahkan flow self-service `become-cv` jika CV perlu pendaftaran mandiri.</li>
            <li>2. Tambahkan halaman incoming material dari collector.</li>
            <li>3. Hubungkan approval admin atau verifikasi partner sesuai proses bisnis.</li>
          </ul>
        </section>
      </div>
    </main>
  );
}