CREATE TABLE IF NOT EXISTS teller_institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    teller_id VARCHAR(255) NOT NULL,
    user_id integer NOT NULL,
    access_token TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_teller_institutions_user_id ON teller_institutions(user_id);
CREATE INDEX IF NOT EXISTS idx_teller_institutions_teller_id ON teller_institutions(teller_id);

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_teller_institutions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_teller_institutions_updated_at 
    BEFORE UPDATE ON teller_institutions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_teller_institutions_updated_at();
