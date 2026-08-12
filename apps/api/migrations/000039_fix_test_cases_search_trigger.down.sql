DROP TRIGGER IF EXISTS test_cases_search_tsv_insert ON test_cases;
DROP TRIGGER IF EXISTS test_cases_search_tsv_update ON test_cases;

CREATE TRIGGER test_cases_search_tsv_insert
    AFTER INSERT ON test_cases
    FOR EACH ROW EXECUTE FUNCTION
    tsvector_update_trigger(search_tsv, 'pg_catalog.english', title, description);

CREATE TRIGGER test_cases_search_tsv_update
    AFTER UPDATE OF title, description ON test_cases
    FOR EACH ROW EXECUTE FUNCTION
    tsvector_update_trigger(search_tsv, 'pg_catalog.english', title, description);
