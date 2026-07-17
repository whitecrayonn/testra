DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE name IN (
        'analytics:read','analytics:create','analytics:update','analytics:delete',
        'intelligence:read','intelligence:create','intelligence:update',
        'integrations:read','integrations:create','integrations:update','integrations:delete',
        'billing:read','billing:update',
        'members:read','members:create','members:update','members:delete',
        'roles:read','roles:create','roles:update','roles:delete'
    )
);

DELETE FROM permissions
WHERE name IN (
    'analytics:read','analytics:create','analytics:update','analytics:delete',
    'intelligence:read','intelligence:create','intelligence:update',
    'integrations:read','integrations:create','integrations:update','integrations:delete',
    'billing:read','billing:update',
    'members:read','members:create','members:update','members:delete',
    'roles:read','roles:create','roles:update','roles:delete'
);
