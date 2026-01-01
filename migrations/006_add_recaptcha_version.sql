-- Migration: Add recaptcha_version column to vhosts table
-- This allows admin to choose between reCAPTCHA v2 (checkbox) or v3 (score-based)

-- Add recaptcha_version column (default to v2 for backward compatibility)
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS recaptcha_version VARCHAR(10) DEFAULT 'v2';

-- Add check constraint to ensure only v2 or v3 values
ALTER TABLE vhosts 
ADD CONSTRAINT check_recaptcha_version 
CHECK (recaptcha_version IN ('v2', 'v3'));

-- Create index for faster filtering
CREATE INDEX IF NOT EXISTS idx_vhosts_recaptcha_version 
ON vhosts(recaptcha_version);

-- Update existing rows to have v2 as default
UPDATE vhosts 
SET recaptcha_version = 'v2' 
WHERE recaptcha_version IS NULL;

-- Add comment
COMMENT ON COLUMN vhosts.recaptcha_version IS 'reCAPTCHA version: v2 (checkbox) or v3 (score-based, invisible)';
