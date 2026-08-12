DROP POLICY IF EXISTS idempotency_records_tenant ON idempotency_records;

ALTER TABLE idempotency_records DISABLE ROW LEVEL SECURITY;

DROP INDEX IF EXISTS idx_idempotency_records_expires;
DROP INDEX IF EXISTS idx_idempotency_records_lookup;

DROP TABLE IF EXISTS idempotency_records;
