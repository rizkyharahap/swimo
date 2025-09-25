-- ACCOUNTS: auth-only
CREATE TABLE IF NOT EXISTS accounts (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email         citext UNIQUE NOT NULL,
  password_hash text   NOT NULL,                -- bcrypt/argon2
  is_locked     boolean NOT NULL DEFAULT false, -- for 403
  created_at    timestamptz NOT NULL DEFAULT now(),
  updated_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_accounts_email ON accounts(email);

-- USERS: profile (1:1 with accounts)
CREATE TABLE IF NOT EXISTS users (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id uuid NOT NULL UNIQUE
             REFERENCES accounts(id) ON DELETE CASCADE,
  name       text   NOT NULL,
  weight_kg  numeric(6,2),                     -- kg
  height_cm  numeric(6,2),                     -- cm
  age_years  smallint,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_weight CHECK (weight_kg IS NULL OR (weight_kg >= 0 AND weight_kg <= 500)),
  CONSTRAINT chk_height CHECK (height_cm IS NULL OR (height_cm >= 0 AND height_cm <= 300)),
  CONSTRAINT chk_age    CHECK (age_years  IS NULL OR (age_years  >= 0 AND age_years  <= 120))
);

-- SESSIONS: guest/user sessions + refresh (opaque, hashed)
CREATE TABLE IF NOT EXISTS sessions (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id          uuid REFERENCES accounts(id) ON DELETE CASCADE, -- NULL for guest
  kind                text NOT NULL CHECK (kind IN ('guest','user')),
  user_agent          text,
  created_at          timestamptz NOT NULL DEFAULT now(),
  last_seen_at        timestamptz NOT NULL DEFAULT now(),
  expires_at          timestamptz NOT NULL,        -- housekeeping window
  refresh_token_hash  text,                        -- hash of opaque refresh token
  refresh_expires_at  timestamptz,
  revoked_at          timestamptz,
  CONSTRAINT guest_no_account CHECK (
    (kind='guest' AND account_id IS NULL) OR
    (kind='user'  AND account_id IS NOT NULL)
  )
);
CREATE INDEX IF NOT EXISTS idx_sessions_account           ON sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires           ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_expires   ON sessions(refresh_expires_at);