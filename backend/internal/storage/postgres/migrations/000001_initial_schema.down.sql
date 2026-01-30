-- Rollback initial schema for Nitro Drag Royale MVP
-- Drops all tables and functions in reverse dependency order

-- Drop triggers first
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_system_wallets_updated_at ON system_wallets;
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS match_settlements;
DROP TABLE IF EXISTS match_participants;
DROP TABLE IF EXISTS ghost_replays;
DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS system_wallets;
DROP TABLE IF EXISTS users;

-- Drop extension (only if no other schemas use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";