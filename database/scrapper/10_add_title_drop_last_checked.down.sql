ALTER TABLE links
DROP COLUMN IF EXISTS title,
ADD COLUMN last_checked timestamp;