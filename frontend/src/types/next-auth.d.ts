/**
 * Type augmentation for next-auth so TypeScript knows about our custom
 * session/user fields (roles, emailVerified, etc.).
 */
import 'next-auth';

declare module 'next-auth' {
  interface Session {
    user: {
      id: string;
      email?: string | null;
      name?: string | null;
      image?: string | null;
      roles: string[];
      emailVerified: boolean;
    };
    accessToken?: string;
    error?: string;
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    accessToken?: string;
    refreshToken?: string;
    accessTokenExpires?: number;
    roles?: string[];
    emailVerified?: boolean;
    error?: string;
  }
}
