'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';

type NavItem = { href: string; label: string; icon: (active: boolean) => React.ReactNode };

const USER_NAV: NavItem[] = [
  {
    href: '/home',
    label: 'Beranda',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
        <polyline points="9 22 9 12 15 12 15 22" />
      </svg>
    ),
  },
  {
    href: '/sell',
    label: 'Jual',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="16" /><line x1="8" y1="12" x2="16" y2="12" />
      </svg>
    ),
  },
  {
    href: '/tracking',
    label: 'Tracking',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
      </svg>
    ),
  },
  {
    href: '/profile',
    label: 'Profil',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
      </svg>
    ),
  },
];

const COLLECTOR_NAV: NavItem[] = [
  {
    href: '/collector/orders',
    label: 'Order',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <rect x="1" y="3" width="15" height="13" /><polygon points="16 8 20 8 23 11 23 16 16 16 16 8" /><circle cx="5.5" cy="18.5" r="2.5" /><circle cx="18.5" cy="18.5" r="2.5" />
      </svg>
    ),
  },
  {
    href: '/profile',
    label: 'Profil',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
      </svg>
    ),
  },
];

const CV_NAV: NavItem[] = [
  {
    href: '/cv',
    label: 'Dashboard',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="4" width="18" height="16" rx="2" />
        <path d="M7 8h10" />
        <path d="M7 12h6" />
        <path d="M7 16h4" />
      </svg>
    ),
  },
  {
    href: '/profile',
    label: 'Profil',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
      </svg>
    ),
  },
];

const ADMIN_NAV: NavItem[] = [
  {
    href: '/admin/users',
    label: 'Users',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" />
      </svg>
    ),
  },
  {
    href: '/admin/applications',
    label: 'Aplikasi',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /><line x1="16" y1="13" x2="8" y2="13" /><line x1="16" y1="17" x2="8" y2="17" /><polyline points="10 9 9 9 8 9" />
      </svg>
    ),
  },
  {
    href: '/profile',
    label: 'Profil',
    icon: (a) => (
      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={a ? 2.5 : 2} strokeLinecap="round" strokeLinejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" /><circle cx="12" cy="7" r="4" />
      </svg>
    ),
  },
];

interface Props { roles: string[] }

export default function BottomNav({ roles }: Props) {
  const pathname = usePathname();

  const items =
    roles.includes('super_admin') || roles.includes('admin') ? ADMIN_NAV :
    roles.includes('cv') ? CV_NAV :
    roles.includes('collector') ? COLLECTOR_NAV :
    USER_NAV;

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-40 flex justify-center">
      <div className="w-full max-w-md bg-surface border-t border-line">
        <div className="flex items-center justify-around px-2 py-2 pb-[env(safe-area-inset-bottom)]">
          {items.map(({ href, label, icon }) => {
            const active = pathname.startsWith(href);
            return (
              <Link
                key={href}
                href={href}
                className="flex flex-col items-center gap-0.5 min-w-[60px] py-1 px-2 rounded-sm transition-colors"
              >
                <span className={active ? 'text-accent' : 'text-ink-4'}>{icon(active)}</span>
                <span className={`text-[10px] font-semibold tracking-tight transition-colors ${active ? 'text-accent' : 'text-ink-4'}`}>
                  {label}
                </span>
              </Link>
            );
          })}
        </div>
      </div>
    </nav>
  );
}
