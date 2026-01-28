-- Add type column to vhosts table
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'proxy';

-- Update existing records if needed (optional, default handles it)
-- UPDATE vhosts SET type = 'proxy' WHERE type IS NULL;
