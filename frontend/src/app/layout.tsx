import type { Metadata, Viewport } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Setor.in — Marketplace daur ulang Indonesia',
  description:
    'Jual sampah Anda jadi penghasilan. Plastik, logam, kertas, dan lainnya — pickup di rumah atau drop-off ke collector terdekat.',
};

// Mobile-first: lock the viewport scaling, prevent iOS auto-zoom on inputs.
export const viewport: Viewport = {
  themeColor: '#2f7d52',
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"
          rel="stylesheet"
        />
      </head>
      <body>{children}</body>
    </html>
  );
}
