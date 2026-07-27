-- 005_rls_payment_methods.down.sql — revierte RLS fail-closed en payment_methods (PLAT-E26)

DROP POLICY IF EXISTS tenant_isolation ON payment_methods;
ALTER TABLE payment_methods NO FORCE ROW LEVEL SECURITY;
ALTER TABLE payment_methods DISABLE ROW LEVEL SECURITY;
