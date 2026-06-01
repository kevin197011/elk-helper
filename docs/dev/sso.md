# OIDC SSO

ELK Helper supports OpenID Connect (OIDC) single sign-on, aligned with [dvr-manager](https://github.com/kevin197011/dvr-manager).

## Flow

1. Login page loads enabled providers from `GET /api/v1/auth/sso/providers`.
2. User clicks **Sign in with {name}** → browser opens `GET /api/v1/auth/sso/oidc/:id/login`.
3. Backend redirects to the IdP; after auth, IdP calls `GET /api/v1/auth/sso/oidc/:id/callback`.
4. Backend issues a JWT and redirects to `{SSO_FRONTEND_BASE_URL}/sso-callback?token=...&username=...&role=...`.
5. Frontend stores the token and loads `/auth/me`.

## Admin setup

1. Sign in as admin → **SSO 配置**.
2. Add an OIDC provider (issuer, client ID/secret, redirect URL, scopes).
3. Register the callback URL at the IdP (path shown in the table), for example:
   - Docker / nginx: `http://localhost:3000/api/v1/auth/sso/oidc/1/callback`
   - Direct backend: `http://localhost:8080/api/v1/auth/sso/oidc/1/callback`
4. **Redirect URL** in the provider config must match the IdP registration exactly.

New SSO users are created with role `user`. Promote to admin under **用户管理** if needed. SSO-only accounts cannot use password login.

## Environment

| Variable | Description |
|----------|-------------|
| `SSO_FRONTEND_BASE_URL` | Frontend origin for post-login redirect (no trailing slash). Default: first `CORS_ORIGINS` entry or `http://localhost:3000`. |
| `CORS_ORIGINS` | Must include the frontend origin used for SSO callback. |

## API summary

| Method | Path | Auth |
|--------|------|------|
| GET | `/api/v1/auth/sso/providers` | Public |
| GET | `/api/v1/auth/sso/oidc/:id/login` | Public |
| GET | `/api/v1/auth/sso/oidc/:id/callback` | Public |
| GET/POST/PUT/DELETE | `/api/v1/sso/providers` | Admin |
| POST | `/api/v1/sso/providers/:id/toggle` | Admin |
