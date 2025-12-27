# OAuth Provider Configuration Guide

Bingo has built-in support for 5 major OAuth platforms: Google, Apple, GitHub, Discord, and Twitter. This document details the configuration for each platform.

## Platform Requirements Overview

| Platform | Cost | Requires App/Website | Review Required | Barrier |
|----------|------|---------------------|-----------------|---------|
| GitHub | Free | No | None | 🟢 Lowest |
| Discord | Free | No | None | 🟢 Lowest |
| Google | Free | Production requires | Verification needed | 🟡 Medium |
| Twitter/X | Free tier | No | Developer account | 🟡 Medium |
| Apple | **$99/year** | Yes | Developer account | 🔴 Highest |

::: tip Quick Start Recommendation
For quick OAuth testing, start with **GitHub** or **Discord**—no review required, instant setup.
:::

## Common Configuration

All OAuth platforms share the following configuration fields:

| Field | Description | Example |
|-------|-------------|---------|
| `client_id` | OAuth App Client ID | `xxx.apps.googleusercontent.com` |
| `client_secret` | OAuth App Client Secret | `GOCSPX-xxx` |
| `redirect_url` | Authorization callback URL | `https://api.example.com/v1/auth/callback` |
| `scopes` | Requested permissions (space-separated) | `openid email profile` |
| `pkce_enabled` | Enable PKCE security mechanism | `true` |

## Google

### Platform Requirements

- **Cost**: Free
- **Testing phase**: No review required, but limited to 100 test users
- **Production**: Requires [brand verification](https://support.google.com/cloud/answer/13464321) (2-3 business days), must provide:
  - Public homepage (verified domain)
  - Privacy policy link
  - Terms of service link
- Sensitive/restricted scopes require additional security assessment

### Creating an OAuth App

1. Visit [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Navigate to **APIs & Services** → **Credentials**
4. Click **Create Credentials** → **OAuth client ID**
5. Select application type (Web application)
6. Configure authorized redirect URIs

### Configuration Parameters

```json
{
  "name": "google",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID.apps.googleusercontent.com",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
  "token_url": "https://oauth2.googleapis.com/token",
  "user_info_url": "https://www.googleapis.com/oauth2/v3/userinfo",
  "scopes": "openid email profile",
  "pkce_enabled": true,
  "field_mapping": {
    "account_id": "sub",
    "email": "email",
    "nickname": "name",
    "avatar": "picture"
  }
}
```

### Obtaining Credentials

After creating an OAuth 2.0 Client ID in Google Cloud Console, you'll receive:
- **Client ID**: Similar to `123456789.apps.googleusercontent.com`
- **Client Secret**: Similar to `GOCSPX-xxxxxxx`

## Apple

### Platform Requirements

- **Cost**: $99/year ([Apple Developer Program](https://developer.apple.com/programs/enroll/))
- **Identity verification**: Passport or government-issued ID required
- **Organization requirements**: D-U-N-S number required (for business accounts)
- **Key management**: Private key can only be downloaded once—store securely

::: warning Note
Apple is the only platform that requires payment. For OAuth testing, consider using other free platforms first.
:::

### Creating an OAuth App

1. Visit [Apple Developer Portal](https://developer.apple.com/)
2. Navigate to **Certificates, Identifiers & Profiles**
3. Create an **App ID** (enable Sign In with Apple)
4. Create a **Services ID** (used as client_id)
5. Create a **Key** (used to generate client_secret)

### Configuration Parameters

```json
{
  "name": "apple",
  "status": "enabled",
  "client_id": "com.example.app.service",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://appleid.apple.com/auth/authorize",
  "token_url": "https://appleid.apple.com/auth/token",
  "scopes": "name email",
  "pkce_enabled": true,
  "field_mapping": {
    "account_id": "sub",
    "email": "email"
  },
  "info": {
    "team_id": "YOUR_TEAM_ID",
    "key_id": "YOUR_KEY_ID",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
  }
}
```

### Obtaining Credentials

| Credential | Description | Location |
|------------|-------------|----------|
| `client_id` | Services ID | Identifiers → Services IDs |
| `team_id` | 10-character team identifier | Top right of developer account |
| `key_id` | Sign In with Apple Key ID | Keys page |
| `private_key` | ECDSA P-256 private key | Downloaded when creating Key (one-time only) |

::: warning Note
The Apple private key can only be downloaded once when created. Store it securely. If lost, you'll need to create a new Key.
:::

### Apple's Special Mechanism

Apple doesn't use a static `client_secret`. Instead, it requires dynamically generating a JWT using the private key. Bingo handles this automatically—you only need to provide the necessary key information in the `info` field.

## GitHub

### Platform Requirements

- **Cost**: Free
- **Review**: None required, instant setup
- **Limits**: 2,000 token requests per hour, max 10 active tokens per user

::: tip Recommended
GitHub is the easiest platform to get started—just need a GitHub account to create an OAuth app.
:::

### Creating an OAuth App

1. Visit [GitHub Developer Settings](https://github.com/settings/developers)
2. Click **New OAuth App**
3. Fill in application information and callback URL
4. After registration, obtain Client ID and Client Secret

### Configuration Parameters

```json
{
  "name": "github",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://github.com/login/oauth/authorize",
  "token_url": "https://github.com/login/oauth/access_token",
  "user_info_url": "https://api.github.com/user",
  "scopes": "read:user user:email",
  "pkce_enabled": false,
  "field_mapping": {
    "account_id": "id",
    "username": "login",
    "nickname": "name",
    "email": "email",
    "avatar": "avatar_url",
    "bio": "bio"
  }
}
```

::: tip Note
GitHub currently doesn't support PKCE, so `pkce_enabled` is set to `false`.
:::

### Obtaining Credentials

From the OAuth App settings page:
- **Client ID**: 20-character string
- **Client Secret**: Click "Generate a new client secret"

## Discord

### Platform Requirements

- **Cost**: Free
- **Review**: None required, instant setup
- **Limits**: Some scopes require Discord approval (e.g., `bot`, `guilds.join`)

::: tip Recommended
Discord is as simple as GitHub—just need a Discord account.
:::

### Creating an OAuth App

1. Visit [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **New Application**
3. Navigate to **OAuth2** page
4. Add redirect URLs
5. Copy Client ID and Client Secret

### Configuration Parameters

```json
{
  "name": "discord",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://discord.com/api/oauth2/authorize",
  "token_url": "https://discord.com/api/oauth2/token",
  "user_info_url": "https://discord.com/api/users/@me",
  "scopes": "identify email",
  "pkce_enabled": true,
  "field_mapping": {
    "account_id": "id",
    "username": "username",
    "nickname": "global_name",
    "email": "email",
    "avatar": "avatar"
  }
}
```

### Obtaining Credentials

From the Application's OAuth2 page:
- **Client ID**: Application ID
- **Client Secret**: Click "Reset Secret" to generate

::: warning Note
Discord's avatar field returns an avatar hash. The full URL format is:
`https://cdn.discordapp.com/avatars/{user_id}/{avatar}.png`
:::

## Twitter

### Platform Requirements

- **Cost**: Free tier available, premium features require payment
- **Developer account**: Must apply at [Developer Portal](https://developer.twitter.com/)
- **Access token validity**: Default 2 hours, use `offline.access` scope for refresh tokens
- **Credential security**: Client ID and Secret shown only once—save immediately

### Creating an OAuth App

1. Visit [Twitter Developer Portal](https://developer.twitter.com/en/portal/dashboard)
2. Create a Project and App
3. Enable **OAuth 2.0** in App settings
4. Configure callback URLs
5. Obtain Client ID and Client Secret

### Configuration Parameters

```json
{
  "name": "twitter",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://twitter.com/i/oauth2/authorize",
  "token_url": "https://api.twitter.com/2/oauth2/token",
  "user_info_url": "https://api.twitter.com/2/users/me",
  "scopes": "users.read tweet.read",
  "pkce_enabled": true,
  "extra_headers": {
    "User-Agent": "BingoApp/1.0"
  },
  "field_mapping": {
    "account_id": "data.id",
    "username": "data.username",
    "nickname": "data.name"
  }
}
```

### Obtaining Credentials

From the App's Keys and tokens page:
- **Client ID**: OAuth 2.0 Client ID
- **Client Secret**: OAuth 2.0 Client Secret

::: tip Twitter API Specifics
Twitter v2 API returns data nested within a `data` object, so `field_mapping` uses dot notation paths (e.g., `data.id`) to extract fields.
:::

## Field Mapping Reference

`field_mapping` maps platform-specific user info fields to unified internal fields:

| Internal Field | Description |
|----------------|-------------|
| `account_id` | User's unique identifier on the platform (required) |
| `username` | Username |
| `email` | Email address |
| `nickname` | Display name |
| `avatar` | Avatar URL |
| `bio` | User bio |

### Nested Fields

For nested JSON responses, use dot-separated paths:

```json
{
  "field_mapping": {
    "account_id": "data.user.id",
    "nickname": "data.user.display_name"
  }
}
```

## Security Mechanisms

### PKCE

PKCE (Proof Key for Code Exchange) provides an additional security layer to prevent authorization code interception:

| Platform | PKCE Support |
|----------|--------------|
| Google | ✅ Supported (recommended) |
| Apple | ✅ Supported (recommended) |
| GitHub | ❌ Not supported |
| Discord | ✅ Supported (recommended) |
| Twitter | ✅ Supported (recommended) |

### State Validation

All platforms use the `state` parameter to prevent CSRF attacks. Bingo automatically generates and validates it (stored in Redis with 5-minute TTL).

## API Operations

### Get Enabled Providers

```bash
GET /v1/auth/providers
```

Response:

```json
{
  "code": 0,
  "data": {
    "providers": [
      {
        "name": "google",
        "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&state=..."
      },
      {
        "name": "github",
        "auth_url": "https://github.com/login/oauth/authorize?client_id=...&state=..."
      }
    ]
  }
}
```

### Login via OAuth

```bash
POST /v1/auth/login/{provider}
Content-Type: application/json

{
  "code": "authorization_code",
  "code_verifier": "PKCE verifier (if PKCE enabled)"
}
```

## Troubleshooting

### redirect_uri Mismatch

Ensure the configured `redirect_url` exactly matches the callback URL registered with the platform, including:
- Protocol (http/https)
- Domain
- Port
- Path

### Unable to Get Email

Some platforms require user consent to share email:
- **Apple**: Users can choose to hide their email
- **GitHub**: Requires `user:email` scope, and email must be set to public
- **Discord**: Requires `email` scope

### PKCE Validation Failed

Ensure the client correctly implements the PKCE flow:
1. Client generates `code_verifier` (43-128 character random string)
2. Compute `code_challenge` (S256 hash)
3. Send `code_challenge` during authorization
4. Send original `code_verifier` when exchanging for token

## Related Documentation

- [Unified Authentication](unified-auth.md) - Authentication architecture design
- [Unified Error Handling](unified-error-handling.md) - Error code specifications
