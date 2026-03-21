CREATE INDEX IF NOT EXISTS idx_block_rules_exact_lookup
    ON block_rules (lower(pattern))
    WHERE enabled = 1 AND rule_type = 'exact';

CREATE INDEX IF NOT EXISTS idx_block_rules_enabled_rule_type
    ON block_rules (enabled, rule_type, pattern);
