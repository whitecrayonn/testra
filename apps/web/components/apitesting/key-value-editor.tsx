"use client";

import { Plus, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import type { KeyValuePair } from "@/types/apitesting";

export function emptyPair(): KeyValuePair {
  return { key: "", value: "", enabled: true };
}

/**
 * Reusable enabled/key/value row editor shared by the API Testing "Studio"
 * page and the test-cases "Quick Generate" form — extracted from
 * api-tests/page.tsx so both places build query params, headers, and
 * variables the same way instead of duplicating the table markup.
 */
export function KeyValueEditor({
  pairs,
  onChange,
  onAdd,
  onRemove,
}: {
  pairs: KeyValuePair[];
  onChange: (index: number, key: "key" | "value" | "enabled", value: string | boolean) => void;
  onAdd: () => void;
  onRemove: (index: number) => void;
}) {
  return (
    <div className="space-y-2">
      <table className="w-full text-sm">
        <thead className="text-left text-slate-500">
          <tr>
            <th className="pb-2 font-medium w-16">Enabled</th>
            <th className="pb-2 font-medium">Key</th>
            <th className="pb-2 font-medium">Value</th>
            <th className="pb-2 font-medium"></th>
          </tr>
        </thead>
        <tbody>
          {pairs.map((pair, i) => (
            <tr key={i} className="border-b border-slate-100">
              <td className="py-2 pr-2">
                <Switch checked={pair.enabled} onCheckedChange={(checked) => onChange(i, "enabled", checked)} />
              </td>
              <td className="py-2 pr-2">
                <Input value={pair.key} onChange={(e) => onChange(i, "key", e.target.value)} className="h-8" />
              </td>
              <td className="py-2 pr-2">
                <Input value={pair.value} onChange={(e) => onChange(i, "value", e.target.value)} className="h-8" />
              </td>
              <td className="py-2">
                <Button variant="ghost" size="sm" onClick={() => onRemove(i)}>
                  <X className="h-4 w-4 text-red-500" />
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      <Button variant="secondary" size="sm" onClick={onAdd}>
        <Plus className="mr-1 h-4 w-4" /> Add
      </Button>
    </div>
  );
}
