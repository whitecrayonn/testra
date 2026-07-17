DROP POLICY IF EXISTS subscriptions_tenant ON subscriptions;
DROP POLICY IF EXISTS invoices_tenant ON invoices;
ALTER TABLE subscriptions DISABLE ROW LEVEL SECURITY;
ALTER TABLE invoices DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
