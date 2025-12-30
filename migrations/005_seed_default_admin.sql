-- Migration: Seed default admin account
-- Created: 2025-12-31
-- Default credentials:
--   Username: admin
--   Password: Admin123!
--   Email: admin@docode.local

-- Insert default admin user
-- Password hash for "Admin123!" generated with bcrypt cost 10
INSERT INTO admins (
    id,
    username,
    email,
    password_hash,
    full_name,
    role,
    is_active,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'admin',
    'admin@docode.local',
    '$2a$10$ivX.3Bp.cZaROB6ZTe901eC3FTb/3EpCSrD1XqYnbmydlhoW3DSO.',
    'System Administrator',
    'superadmin',
    true,
    NOW(),
    NOW()
)
ON CONFLICT (username) DO NOTHING;

-- Note: ON CONFLICT ensures this is idempotent and safe to run multiple times

