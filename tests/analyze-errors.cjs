const fs = require('fs');
const d = JSON.parse(fs.readFileSync('reports/json/test-results.json', 'utf8'));

const failed = [];
function walk(suite) {
    if (suite.specs) {
        for (const spec of suite.specs) {
            for (const test of spec.tests) {
                if (test.status === 'unexpected') {
                    const failResult = test.results.find(r => r.status === 'failed') || test.results[test.results.length - 1];
                    const errors = (failResult.errors || []).map(e => e.message || e.snippet || '').join('\n');
                    failed.push({
                        title: spec.title,
                        project: test.projectName,
                        file: suite.title,
                        error: errors.substring(0, 500),
                    });
                }
            }
        }
    }
    if (suite.suites) for (const sub of suite.suites) walk(sub);
}
for (const s of d.suites) walk(s);

// Group by error pattern
const byError = {};
for (const f of failed) {
    // Extract key error message
    let key = f.error.split('\n')[0].substring(0, 200);
    if (!byError[key]) byError[key] = [];
    byError[key].push(f);
}

console.log('=== ERROR PATTERNS (sorted by frequency) ===');
const sorted = Object.entries(byError).sort((a, b) => b[1].length - a[1].length);
for (const [error, tests] of sorted) {
    const browsers = [...new Set(tests.map(t => t.project))];
    console.log(`\n[${tests.length} occurrences] [${browsers.join(',')}]`);
    console.log(`  Error: ${error}`);
    console.log(`  Tests: ${tests.map(t => t.title).slice(0, 3).join(', ')}${tests.length > 3 ? '...' : ''}`);
}
