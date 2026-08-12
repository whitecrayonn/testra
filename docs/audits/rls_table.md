| Table | Enabled in migration | Policies | Pattern summary |
|---|---|---|---|
| analytics_daily_metrics | 000021_add_analytics.up.sql | analytics_daily_metrics_tenant | current_setting(app.tenant_id, true) |
| analytics_dashboards | 000021_add_analytics.up.sql | analytics_dashboards_tenant | current_setting(app.tenant_id, true) |
| api_collections | 000032_add_api_testing.up.sql | api_collections_tenant | current_setting(app.tenant_id, true) |
| api_environments | 000032_add_api_testing.up.sql | api_environments_tenant | current_setting(app.tenant_id, true) |
| api_folders | 000032_add_api_testing.up.sql | api_folders_tenant | current_setting(app.tenant_id, true) |
| api_keys | 000009_add_rls_policies.up.sql | api_keys_tenant, api_keys_lookup_by_hash, api_keys_lookup_user | current_setting(app.tenant_id, true); COALESCE; lookup/non-tenant session variable |
| api_request_history | 000032_add_api_testing.up.sql | api_request_history_tenant | current_setting(app.tenant_id, true) |
| api_requests | 000032_add_api_testing.up.sql | api_requests_tenant | current_setting(app.tenant_id, true) |
| automation_artifacts | 000033_add_automation_hub.up.sql | automation_artifacts_tenant, automation_artifacts_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| automation_executions | 000033_add_automation_hub.up.sql | automation_executions_tenant, automation_executions_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| automation_logs | 000033_add_automation_hub.up.sql | automation_logs_tenant, automation_logs_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| automation_projects | 000033_add_automation_hub.up.sql | automation_projects_tenant, automation_projects_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| defects | 000020_add_defects.up.sql | defects_tenant, defects_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| failure_clusters | 000022_add_intelligence.up.sql | failure_clusters_tenant | current_setting(app.tenant_id, true) |
| flaky_predictions | 000022_add_intelligence.up.sql | flaky_predictions_tenant | current_setting(app.tenant_id, true) |
| idempotency_records | 000017_add_idempotency_records.up.sql | idempotency_records_tenant, idempotency_records_tenant | current_setting(app.tenant_id, true); current_setting(app.tenant_id, true) |
| integration_events | 000023_add_integrationhub.up.sql | integration_events_tenant | current_setting(app.tenant_id, true) |
| integrations | 000023_add_integrationhub.up.sql | integrations_tenant | current_setting(app.tenant_id, true) |
| invoices | 000024_add_billing.up.sql | invoices_tenant | current_setting(app.tenant_id, true) |
| notification_channels | 000018_add_notifications.up.sql | notification_channels_tenant, notification_channels_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| notification_history | 000036_notification_templates_history.up.sql | notification_history_tenant, notification_history_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| notification_preferences | 000018_add_notifications.up.sql | notification_preferences_tenant, notification_preferences_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| notification_templates | 000036_notification_templates_history.up.sql | notification_templates_tenant, notification_templates_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| notifications | 000018_add_notifications.up.sql | notifications_tenant, notifications_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| organization_members | 000009_add_rls_policies.up.sql | org_members_tenant, organization_members_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| organizations | 000009_add_rls_policies.up.sql | org_tenant_isolation, organizations_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| projects | 000009_add_rls_policies.up.sql | projects_tenant, projects_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| queue_jobs | 000026_worker_queue.up.sql | tenant_isolation_queue_jobs | OR |
| role_assignments | 000009_add_rls_policies.up.sql | role_assignments_tenant, role_assignments_tenant | current_setting(app.tenant_id, true); current_setting(app.tenant_id, true) OR |
| subscriptions | 000024_add_billing.up.sql | subscriptions_tenant | current_setting(app.tenant_id, true) |
| test_case_versions | 000014_add_test_management_rls.up.sql | test_case_versions_tenant, test_case_versions_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| test_cases | 000014_add_test_management_rls.up.sql | test_cases_tenant, test_cases_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| test_folders | 000014_add_test_management_rls.up.sql | test_folders_tenant, test_folders_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| test_plan_items | 000034_manual_execution.up.sql | test_plan_items_tenant | current_setting(app.tenant_id, true) |
| test_plans | 000034_manual_execution.up.sql | test_plans_tenant | current_setting(app.tenant_id, true) |
| test_run_item_defects | 000034_manual_execution.up.sql | test_run_item_defects_tenant | current_setting(app.tenant_id, true) |
| test_run_item_evidence | 000034_manual_execution.up.sql | test_run_item_evidence_tenant | current_setting(app.tenant_id, true) |
| test_run_item_history | 000034_manual_execution.up.sql | test_run_item_history_tenant | current_setting(app.tenant_id, true) |
| test_run_items | 000015_add_test_runs.up.sql | test_run_items_tenant, test_run_items_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| test_runs | 000015_add_test_runs.up.sql | test_runs_tenant, test_runs_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| test_suites | 000014_add_test_management_rls.up.sql | test_suites_tenant, test_suites_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
| workspace_members | 000009_add_rls_policies.up.sql | workspace_members_tenant | current_setting(app.tenant_id, true) |
| workspaces | 000009_add_rls_policies.up.sql | workspaces_tenant, workspaces_lookup_user | current_setting(app.tenant_id, true); lookup/non-tenant session variable |
