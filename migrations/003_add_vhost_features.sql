-- Add bot detection and rate limiter settings to vhosts table
ALTER TABLE vhosts 
ADD COLUMN bot_detection_enabled BOOLEAN DEFAULT false,
ADD COLUMN bot_detection_type VARCHAR(50) DEFAULT 'turnstile', -- turnstile, captcha, slide_puzzle
ADD COLUMN rate_limit_enabled BOOLEAN DEFAULT false,
ADD COLUMN rate_limit_requests INT DEFAULT 100,
ADD COLUMN rate_limit_window INT DEFAULT 60; -- in seconds

-- Add indexes for performance
CREATE INDEX idx_vhosts_bot_detection ON vhosts(bot_detection_enabled);
CREATE INDEX idx_vhosts_rate_limit ON vhosts(rate_limit_enabled);

COMMENT ON COLUMN vhosts.bot_detection_enabled IS 'Enable bot detection for this vhost';
COMMENT ON COLUMN vhosts.bot_detection_type IS 'Type of bot detection: turnstile, captcha, slide_puzzle';
COMMENT ON COLUMN vhosts.rate_limit_enabled IS 'Enable rate limiting for this vhost';
COMMENT ON COLUMN vhosts.rate_limit_requests IS 'Maximum requests allowed per window';
COMMENT ON COLUMN vhosts.rate_limit_window IS 'Time window in seconds';
