const fs = require('fs');
const d = JSON.parse(fs.readFileSync('reports/json/test-results.json', 'utf8'));

console.log('=== OVERALL STATS ===');
console.log(JSON.stringify(d.stats, null, 2));

const failed = [];
const passed = [];
const flaky = [];

function walkSuite(suite) {
    if (suite.specs) {
        for (const spec of suite.specs) {
            for (const test of spec.tests) {
                const statuses = test.results.map(r => r.status);
                const hasFail = statuses.includes('failed');
                const hasPass = statuses.includes('passed');
                const info = {
                    title: spec.title,
                    project: test.projectName,
                    file: suite.title,
                    status: test.status,
                };
                if (test.status === 'unexpected' || (hasFail && !hasPass)) {
                    failed.push(info);
                } else if (test.status === 'flaky') {
                    flaky.push(info);
                } else {
                    passed.push(info);
                }
            }
        }
    }
    if (suite.suites) {
        for (const sub of suite.suites) walkSuite(sub);
    }
}

for (const suite of d.suites) walkSuite(suite);

console.log('\n=== FAILURES BY BROWSER ===');
const byBrowser = {};
for (const f of failed) byBrowser[f.project] = (byBrowser[f.project] || 0) + 1;
for (const [k, v] of Object.entries(byBrowser).sort((a,b) => b[1]-a[1])) {
    console.log(`  ${k}: ${v}`);
}

console.log('\n=== FAILURES BY TEST FILE ===');
const byFile = {};
for (const f of failed) byFile[f.file] = (byFile[f.file] || 0) + 1;
for (const [k, v] of Object.entries(byFile).sort((a,b) => b[1]-a[1])) {
    console.log(`  ${k}: ${v}`);
}

console.log('\n=== UNIQUE FAILING TEST TITLES ===');
const unique = [...new Set(failed.map(f => f.title))].sort();
for (const t of unique) {
    const browsers = [...new Set(failed.filter(f => f.title === t).map(f => f.project))];
    console.log(`  [${browsers.join(',')}] ${t}`);
}

console.log('\n=== FLAKY TESTS ===');
for (const f of flaky) {
    console.log(`  [${f.project}] ${f.title}`);
}
