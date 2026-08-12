import { readFileSync, readdirSync, writeFileSync } from "fs";
import { join, resolve } from "path";

const root = resolve(process.cwd(), "apps/api/migrations");
const files = readdirSync(root)
  .filter((f) => f.endsWith(".up.sql"))
  .sort();

const policies = [];
const enabledTables = [];

for (const file of files) {
  const sql = readFileSync(join(root, file), "utf8");

  // Find ENABLE ROW LEVEL SECURITY statements
  const enableRe = /ALTER TABLE\s+(\w+)\s+ENABLE ROW LEVEL SECURITY/gi;
  let em;
  while ((em = enableRe.exec(sql)) !== null) {
    enabledTables.push({ file, table: em[1] });
  }

  // Find CREATE POLICY / ALTER POLICY blocks (non-greedy up to the first semicolon after USING/CHECK)
  const policyRe = /(?:CREATE|ALTER)\s+POLICY\s+(\w+)\s+ON\s+(\w+)\s*([\s\S]*?);/gi;
  let m;
  while ((m = policyRe.exec(sql)) !== null) {
    const name = m[1];
    const table = m[2];
    const body = m[3].trim();

    const hasTrue = body.includes("current_setting('app.tenant_id', true)");
    const hasPlain = /current_setting\('app\.tenant_id'\s*\)/.test(body);
    const hasOr = /\bOR\b/i.test(body);
    const hasCoalesce = /\bCOALESCE\b/i.test(body);
    const hasDefault = /default|DEFAULT/.test(body);
    const usingMatch = body.match(/USING\s*\(([\s\S]*)\)/i);
    const usingClause = usingMatch ? usingMatch[1].trim().replace(/\s+/g, " ") : "";

    policies.push({
      file,
      table,
      name,
      hasTrue,
      hasPlain,
      hasOr,
      hasCoalesce,
      hasDefault,
      usingClause,
      bodyPreview: body.slice(0, 160).replace(/\s+/g, " "),
    });
  }
}

const out = { enabledTables, policies, summary: {
  totalEnabled: enabledTables.length,
  totalPolicies: policies.length,
  withTrue: policies.filter((p) => p.hasTrue).length,
  withPlain: policies.filter((p) => p.hasPlain).length,
  withOr: policies.filter((p) => p.hasOr).length,
  withCoalesce: policies.filter((p) => p.hasCoalesce).length,
}};

writeFileSync("rls_policies.json", JSON.stringify(out, null, 2));
console.log(`Enabled ${enabledTables.length} tables, found ${policies.length} policies.`);
