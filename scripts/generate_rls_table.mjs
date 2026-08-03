import { readFileSync, writeFileSync } from "fs";

const d = JSON.parse(readFileSync("rls_policies.json", "utf8"));
const map = {};
for (const t of d.enabledTables) {
  map[t.table] = { file: t.file, policies: [] };
}
for (const p of d.policies) {
  map[p.table].policies.push(p);
}

let out = "| Table | Enabled in migration | Policies | Pattern summary |\n";
out += "|---|---|---|---|\n";
const sorted = Object.keys(map).sort();
for (const tbl of sorted) {
  const m = map[tbl];
  const names = m.policies.map((x) => x.name).join(", ");
  const patterns = m.policies
    .map((x) => {
      let s = "";
      if (x.hasTrue) s += "current_setting(app.tenant_id, true) ";
      if (x.hasOr) s += "OR ";
      if (x.hasCoalesce) s += "COALESCE ";
      if (!s) s = "lookup/non-tenant session variable";
      return s.trim();
    })
    .join("; ");
  out += `| ${tbl} | ${m.file} | ${names} | ${patterns} |\n`;
}
writeFileSync("rls_table.md", out);
console.log("wrote rls_table.md");
