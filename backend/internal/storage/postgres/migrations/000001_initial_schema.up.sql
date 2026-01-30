-- Initial schema for Nitro Drag Royale MVP
-- Creates all 9 tables with constraints, indexes, and seed data

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id BIGINT UNIQUE NOT NULL CHECK (telegram_id > 0),
    telegram_username VARCHAR(255),
    telegram_first_name VARCHAR(255) NOT NULL,
    telegram_last_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for Telegram auth lookup
CREATE INDEX idx_users_telegram_id ON users(telegram_id);

-- 2. System Wallets table (must be created before wallets for FK reference)
CREATE TABLE system_wallets (
    wallet_name VARCHAR(50) PRIMARY KEY,
    fuel_balance DECIMAL(16,2) NOT NULL DEFAULT 0.00,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed system wallets
INSERT INTO system_wallets (wallet_name, fuel_balance) VALUES
('HOUSE_FUEL', 10000.00),  -- Initial house balance for Ghost operations
('RAKE_FUEL', 0.00);       -- Rake collection wallet starts empty

-- 3. Wallets table
CREATE TABLE wallets (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    ton_balance DECIMAL(16,2) NOT NULL DEFAULT 0.00 CHECK (ton_balance >= 0),
    fuel_balance DECIMAL(16,2) NOT NULL DEFAULT 0.00 CHECK (fuel_balance >= 0),
    burn_balance DECIMAL(16,2) NOT NULL DEFAULT 0.00 CHECK (burn_balance >= 0),
    rookie_races_completed INT NOT NULL DEFAULT 0 CHECK (rookie_races_completed >= 0 AND rookie_races_completed <= 3),
    ton_wallet_address VARCHAR(66),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 4. Ledger Entries table
CREATE TABLE ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    system_wallet VARCHAR(50) REFERENCES system_wallets(wallet_name),
    currency VARCHAR(10) NOT NULL CHECK (currency IN ('TON', 'FUEL', 'BURN')),
    amount DECIMAL(16,2) NOT NULL,
    operation_type VARCHAR(50) NOT NULL CHECK (operation_type IN (
        'DEPOSIT', 'WITHDRAWAL', 'MATCH_BUYIN', 'MATCH_PRIZE', 
        'MATCH_RAKE', 'MATCH_BURN_REWARD', 'INITIAL_BALANCE'
    )),
    reference_id UUID,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ensure either user_id or system_wallet is set, but not both
    CONSTRAINT ledger_entries_wallet_check CHECK (
        (user_id IS NOT NULL AND system_wallet IS NULL) OR
        (user_id IS NULL AND system_wallet IS NOT NULL)
    )
);

-- Indexes for ledger queries
CREATE INDEX idx_ledger_user_id ON ledger_entries(user_id);
CREATE INDEX idx_ledger_reference_id ON ledger_entries(reference_id);
CREATE INDEX idx_ledger_created_at ON ledger_entries(created_at);
CREATE INDEX idx_ledger_system_wallet ON ledger_entries(system_wallet);

-- 5. Matches table
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    league VARCHAR(20) NOT NULL CHECK (league IN ('ROOKIE', 'STREET', 'PRO', 'TOP_FUEL')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('FORMING', 'IN_PROGRESS', 'COMPLETED', 'ABORTED')),
    live_player_count INT NOT NULL CHECK (live_player_count >= 1 AND live_player_count <= 10),
    ghost_player_count INT NOT NULL CHECK (ghost_player_count >= 0 AND ghost_player_count <= 9),
    prize_pool DECIMAL(16,2) NOT NULL CHECK (prize_pool >= 0),
    rake_amount DECIMAL(16,2) NOT NULL CHECK (rake_amount >= 0),
    crash_seed VARCHAR(128) NOT NULL,
    crash_seed_hash VARCHAR(64) NOT NULL,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Always exactly 10 players total
    CONSTRAINT matches_player_count_check CHECK (live_player_count + ghost_player_count = 10)
);

-- Indexes for match queries
CREATE INDEX idx_matches_status ON matches(status);
CREATE INDEX idx_matches_created_at ON matches(created_at);
CREATE INDEX idx_matches_league ON matches(league);

-- 6. Ghost Replays table (must be created before match_participants for FK reference)
CREATE TABLE ghost_replays (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_match_id UUID REFERENCES matches(id) ON DELETE CASCADE,
    source_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    league VARCHAR(20) NOT NULL CHECK (league IN ('ROOKIE', 'STREET', 'PRO', 'TOP_FUEL')),
    display_name VARCHAR(255) NOT NULL,
    heat1_score DECIMAL(8,2) NOT NULL CHECK (heat1_score >= 0),
    heat2_score DECIMAL(8,2) NOT NULL CHECK (heat2_score >= 0),
    heat3_score DECIMAL(8,2) NOT NULL CHECK (heat3_score >= 0),
    total_score DECIMAL(8,2) NOT NULL CHECK (total_score >= 0),
    behavioral_data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ensure total_score equals sum of heats
    CONSTRAINT ghost_replays_total_score_check CHECK (total_score = heat1_score + heat2_score + heat3_score)
);

-- Indexes for Ghost selection
CREATE INDEX idx_ghost_replays_league ON ghost_replays(league);
CREATE INDEX idx_ghost_replays_total_score ON ghost_replays(total_score);

-- 7. Match Participants table
CREATE TABLE match_participants (
    id BIGSERIAL PRIMARY KEY,  -- Surrogate key for Ghosts (user_id can be NULL)
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    is_ghost BOOLEAN NOT NULL DEFAULT FALSE,
    ghost_replay_id UUID REFERENCES ghost_replays(id),
    player_display_name VARCHAR(255) NOT NULL,
    buyin_amount DECIMAL(16,2) NOT NULL CHECK (buyin_amount >= 0),
    heat1_score DECIMAL(8,2) CHECK (heat1_score >= 0),
    heat2_score DECIMAL(8,2) CHECK (heat2_score >= 0),
    heat3_score DECIMAL(8,2) CHECK (heat3_score >= 0),
    total_score DECIMAL(8,2) CHECK (total_score >= 0),
    final_position INT CHECK (final_position >= 1 AND final_position <= 10),
    prize_amount DECIMAL(16,2) NOT NULL DEFAULT 0.00 CHECK (prize_amount >= 0),
    burn_reward DECIMAL(16,2) NOT NULL DEFAULT 0.00 CHECK (burn_reward >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Ghost XOR live player constraint
    CONSTRAINT match_participants_ghost_check CHECK (
        (is_ghost = TRUE AND user_id IS NULL AND ghost_replay_id IS NOT NULL) OR
        (is_ghost = FALSE AND user_id IS NOT NULL AND ghost_replay_id IS NULL)
    ),
    
    -- Total score equals sum of heats (when all heats are completed)
    CONSTRAINT match_participants_total_score_check CHECK (
        total_score IS NULL OR 
        total_score = COALESCE(heat1_score, 0) + COALESCE(heat2_score, 0) + COALESCE(heat3_score, 0)
    )
);

-- Indexes for participant queries
CREATE INDEX idx_match_participants_user_id ON match_participants(user_id);
CREATE INDEX idx_match_participants_match_id ON match_participants(match_id);
CREATE INDEX idx_match_participants_is_ghost ON match_participants(is_ghost);

-- Unique constraint for live players (one participation per match per user)
CREATE UNIQUE INDEX idx_match_participants_user_match ON match_participants(match_id, user_id) 
WHERE user_id IS NOT NULL;

-- 8. Match Settlements table
CREATE TABLE match_settlements (
    match_id UUID PRIMARY KEY REFERENCES matches(id) ON DELETE CASCADE,
    settled_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 9. Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payment_type VARCHAR(20) NOT NULL CHECK (payment_type IN ('DEPOSIT', 'WITHDRAWAL')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('PENDING', 'CONFIRMED', 'FAILED')),
    ton_amount DECIMAL(16,2) NOT NULL CHECK (ton_amount > 0),
    fuel_amount DECIMAL(16,2) NOT NULL CHECK (fuel_amount > 0),
    ton_tx_hash VARCHAR(128) UNIQUE,
    client_request_id UUID NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for payment queries
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_ton_tx_hash ON payments(ton_tx_hash);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_client_request_id ON payments(client_request_id);

-- Update triggers for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply update triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_wallets_updated_at BEFORE UPDATE ON system_wallets 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at BEFORE UPDATE ON payments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();