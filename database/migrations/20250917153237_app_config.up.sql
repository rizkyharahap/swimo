CREATE TABLE IF NOT EXISTS app_config (
  id                    boolean PRIMARY KEY DEFAULT true,  -- singleton row
  guest_sign_in_enabled boolean NOT NULL DEFAULT true,
  guest_active_limit    integer NOT NULL DEFAULT 1000,
  created_at            timestamptz NOT NULL DEFAULT now(),
  updated_at            timestamptz NOT NULL DEFAULT now()
);

INSERT INTO app_config (id, guest_sign_in_enabled, guest_active_limit)
VALUES (true, true, 4)
ON CONFLICT (id) DO NOTHING;
