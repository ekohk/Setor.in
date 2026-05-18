import { redirect } from 'next/navigation';
import { api } from '@/lib/api';
import type { UserProfile } from '@/types/api';
import ProfileClient from './ProfileClient';

export default async function ProfilePage() {
  let profile: UserProfile;
  try {
    profile = await api<UserProfile>('/v1/users/me');
  } catch {
    // If sync not done yet, redirect to home which will handle it gracefully
    redirect('/home');
  }

  return <ProfileClient initialProfile={profile} />;
}
