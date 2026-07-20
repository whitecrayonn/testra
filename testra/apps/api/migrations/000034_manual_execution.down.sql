-- Roll back manual execution schema additions.

DROP TABLE IF EXISTS test_plan_items;
DROP TABLE IF EXISTS test_plans;
DROP TABLE IF EXISTS test_run_item_defects;
DROP TABLE IF EXISTS test_run_item_evidence;
DROP TABLE IF EXISTS test_run_item_history;

ALTER TABLE test_run_items
    DROP COLUMN IF EXISTS step_results,
    DROP COLUMN IF EXISTS executed_by,
    DROP COLUMN IF EXISTS executed_at,
    DROP COLUMN IF EXISTS comment;
