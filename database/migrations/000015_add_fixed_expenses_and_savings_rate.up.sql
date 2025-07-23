ALTER TABLE monthly_summary
    ADD COLUMN fixed_expenses float DEFAULT 0,
    ADD COLUMN saving_target_percentage float DEFAULT 0;
