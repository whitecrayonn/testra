-- Rollback API Testing Engine schema

DELETE FROM role_permissions
WHERE permission_id IN (
    '00000000-0000-0000-0000-000000002001',
    '00000000-0000-0000-0000-000000002002',
    '00000000-0000-0000-0000-000000002003',
    '00000000-0000-0000-0000-000000002004',
    '00000000-0000-0000-0000-000000002005'
);

DELETE FROM permissions
WHERE name IN ('api_tests:read', 'api_tests:create', 'api_tests:update', 'api_tests:delete', 'api_tests:execute');

DROP TABLE IF EXISTS api_request_history;
DROP TABLE IF EXISTS api_requests;
DROP TABLE IF EXISTS api_environments;
DROP TABLE IF EXISTS api_folders;
DROP TABLE IF EXISTS api_collections;
