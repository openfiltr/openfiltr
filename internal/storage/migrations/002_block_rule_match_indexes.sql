CREATE INDEX IF NOT EXISTS idx_block_rules_pattern_lower_exact
    ON block_rules (LOWER(pattern))
    WHERE enabled = 1 AND rule_type = 'exact';

CREATE INDEX IF NOT EXISTS idx_block_rules_enabled_type
    ON block_rules (enabled, rule_type);

CREATE INDEX IF NOT EXISTS idx_block_rules_regex_lookup
    ON block_rules (LOWER(pattern))
    WHERE enabled = 1 AND rule_type = 'regex';
