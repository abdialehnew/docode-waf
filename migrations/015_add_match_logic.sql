-- Migration: Add match_logic to rules table
-- Created: 2026-02-05

ALTER TABLE rules ADD COLUMN IF NOT EXISTS match_logic VARCHAR(10) DEFAULT 'AND';
