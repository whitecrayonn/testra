const r = require('./reports/json/test-results.json');
const stats = { passed: 0, failed: 0, flaky: 0, skipped: 0 };
const fails = [];

function walkSuite(suite, filePrefix) {
  const file = suite.file || filePrefix;
  (suite.suites || []).forEach(s => walkSuite(s, file));
  (suite.specs || []).forEach(spec => {
    spec.tests.forEach(t => {
      const lastStatus = t.results[t.results.length - 1].status;
      const hadFailure = t.results.some(rr => rr.status === 'failed' || rr.status === 'timedOut');
      if (lastStatus === 'passed' && hadFailure) {
        stats.flaky++;
        fails.push({ type: 'FLAKY', file, title: spec.title, project: t.projectName });
      } else if (lastStatus === 'passed') {
        stats.passed++;
      } else if (lastStatus === 'skipped') {
        stats.skipped++;
      } else {
        stats.failed++;
        const err = (t.results[t.results.length - 1].error || {}).message || '';
        fails.push({ type: 'FAIL', file, title: spec.title, project: t.projectName, error: err.split('\n').slice(0,4).join(' | ') });
      }
    });
  });
}

r.suites.forEach(s => walkSuite(s));

console.log('STATS', JSON.stringify(stats));
console.log('TOTAL_ISSUES', fails.length);
require('fs').writeFileSync('failure-report.json', JSON.stringify(fails, null, 2));

// group by file
const byFile = {};
fails.forEach(f => {
  byFile[f.file] = byFile[f.file] || [];
  byFile[f.file].push(f.title);
});
Object.keys(byFile).sort().forEach(f => {
  console.log(f, byFile[f].length);
});
