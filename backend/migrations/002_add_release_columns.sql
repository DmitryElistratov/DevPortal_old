-- Migration: Add missing columns to releases table
-- This migration adds columns that were added after the initial table creation

-- Add git_commit column if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='releases' AND column_name='git_commit'
    ) THEN
        ALTER TABLE releases ADD COLUMN git_commit VARCHAR(100);
    END IF;
END $$;

-- Add is_primary column if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='releases' AND column_name='is_primary'
    ) THEN
        ALTER TABLE releases ADD COLUMN is_primary BOOLEAN DEFAULT true;
    END IF;
END $$;

-- Add traffic_percentage column if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name='releases' AND column_name='traffic_percentage'
    ) THEN
        ALTER TABLE releases ADD COLUMN traffic_percentage INTEGER DEFAULT 100;
    END IF;
END $$;
