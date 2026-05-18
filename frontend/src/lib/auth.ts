/**
 * next-auth v5 (Auth.js) configuration.
 *
 * We use the Credentials provider to wrap Keycloak's Direct Access Grant —
 * the user submits email+password to OUR /login page, our backend (Auth.js
 * route handler) calls Keycloak's token endpoint, and on success Auth.js
 * sets an httpOnly session cookie. The actual access_token is stored in
 * the encrypted JWT session so it can be attached to backend API calls
 * without exposing it to client-side JS (XSS-safe).
 *
 * Why Credentials over OAuth provider:
 *   - We control the login UI (branded EcoCycle), not Keycloak's hosted page.
 *   - Token storage stays httpOnly cookie — same as the OAuth flow would give.
 *   - Forgot-password / register still link out to Keycloak's hosted pages
 *     (those are infrequent and not worth re-implementing).
 *
 * Token refresh: handled in jwt() callback. When access_token nears expiry
 * we call Keycloak's refresh endpoint and update the session.
 */

import NextAuth, { type NextAuthConfig } from 'next-auth';
import Credentials from 'next-auth/providers/credentials';

const KC_BASE = process.env.KEYCLOAK_BASE_URL!;
const KC_REALM = process.env.KEYCLOAK_REALM!;
const KC_CLIENT_ID = process.env.KEYCLOAK_CLIENT_ID!;

const TOKEN_URL = `${KC_BASE}/realms/${KC_REALM}/protocol/openid-connect/token`;

export const authConfig: NextAuthConfig = {
  trustHost: true,
  session: { strategy: 'jwt' },
  pages: {
    signIn: '/login',
    error: '/login', // surface errors via ?error= on /login
  },
  providers: [
    Credentials({
      id: 'keycloak-credentials',
      name: 'Setor.in',
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' },
      },
      async authorize(creds) {
        const email = String(creds?.email ?? '').trim();
        const password = String(creds?.password ?? '');
        if (!email || !password) return null;

        const body = new URLSearchParams({
          grant_type: 'password',
          client_id: KC_CLIENT_ID,
          username: email,
          password,
          scope: 'openid profile email',
        });

        const res = await fetch(TOKEN_URL, {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body,
          cache: 'no-store',
        });

        if (!res.ok) {
          const err = (await res.json().catch(() => ({}))) as { error?: string };
          // Throw to surface as next-auth error code on the login page
          throw new Error(err.error ?? 'invalid_grant');
        }

        const tok = (await res.json()) as TokenResponse;
        const claims = decodeJwt(tok.access_token);

        // The "user" we return is the seed for the JWT cookie.
        return {
          id: claims.sub,
          email: claims.email ?? email,
          name: claims.name ?? '',
          // Custom fields piped into the JWT via callbacks.jwt
          accessToken: tok.access_token,
          refreshToken: tok.refresh_token,
          accessTokenExpires: Date.now() + tok.expires_in * 1000,
          roles: claims.realm_access?.roles ?? [],
          emailVerified: claims.email_verified ?? false,
        } as never;
      },
    }),
  ],
  callbacks: {
    async jwt({ token, user, trigger }) {
      // First sign-in: copy fields from authorize() return into the JWT,
      // then call /v1/auth/sync so the backend creates a local user row.
      // sync is idempotent — safe to call on every fresh login.
      if (user) {
        const accessToken = (user as Record<string, unknown>).accessToken as string;
        // Sync once at login — runs in background, does not block token return.
        fetch(`${process.env.BACKEND_URL ?? 'http://localhost:8000'}/v1/auth/sync`, {
          method: 'POST',
          headers: { Authorization: `Bearer ${accessToken}` },
          cache: 'no-store',
        }).catch(() => {});
        return {
          ...token,
          accessToken,
          refreshToken: (user as Record<string, unknown>).refreshToken,
          accessTokenExpires: (user as Record<string, unknown>).accessTokenExpires,
          roles: (user as Record<string, unknown>).roles,
          emailVerified: (user as Record<string, unknown>).emailVerified,
        };
      }

      // Subsequent calls: refresh access_token if it's about to expire (60s buffer).
      const exp = token.accessTokenExpires as number | undefined;
      if (exp && Date.now() < exp - 60_000) return token;

      // Refresh
      try {
        const refreshed = await refreshAccessToken(token.refreshToken as string);
        return {
          ...token,
          accessToken: refreshed.access_token,
          refreshToken: refreshed.refresh_token,
          accessTokenExpires: Date.now() + refreshed.expires_in * 1000,
        };
      } catch {
        // Force re-login by marking the session as errored.
        return { ...token, error: 'RefreshAccessTokenError' };
      }
    },
    async session({ session, token }) {
      // Expose what client + server pages need. NEVER expose refreshToken to client.
      session.user = {
        ...session.user,
        id: (token.sub as string) ?? '',
        roles: (token.roles as string[]) ?? [],
        emailVerified: Boolean(token.emailVerified),
      } as typeof session.user;
      // accessToken is needed server-side to call backend; it is augmented
      // onto Session via the module declaration in `src/types/next-auth.d.ts`.
      session.accessToken = token.accessToken as string | undefined;
      session.error = token.error as string | undefined;
      return session;
    },
  },
};

export const { handlers, auth, signIn, signOut } = NextAuth(authConfig);

// ─── Helpers ────────────────────────────────────────────────────────────

interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  refresh_expires_in: number;
  token_type: string;
}

interface JwtClaims {
  sub: string;
  email?: string;
  name?: string;
  email_verified?: boolean;
  realm_access?: { roles: string[] };
  exp?: number;
}

function decodeJwt(token: string): JwtClaims {
  const [, payload] = token.split('.');
  const json = Buffer.from(payload.replace(/-/g, '+').replace(/_/g, '/'), 'base64').toString('utf-8');
  return JSON.parse(json);
}

async function refreshAccessToken(refreshToken: string): Promise<TokenResponse> {
  const body = new URLSearchParams({
    grant_type: 'refresh_token',
    client_id: KC_CLIENT_ID,
    refresh_token: refreshToken,
  });
  const res = await fetch(TOKEN_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
    cache: 'no-store',
  });
  if (!res.ok) throw new Error('refresh failed');
  return (await res.json()) as TokenResponse;
}
