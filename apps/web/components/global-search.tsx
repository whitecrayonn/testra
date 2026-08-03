"use client";

import { useEffect, useRef, useState, useMemo, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Search, X, Loader2 } from "lucide-react";
import { useWorkspace } from "@/lib/hooks/use-workspace";
import { globalSearch, type SearchItem, type SearchResult } from "@/features/search/api";
import { cn } from "@/lib/utils";

const sectionLabels: Record<string, string> = {
  projects: "Projects",
  test_cases: "Test Cases",
  defects: "Defects",
  automation: "Automation",
  api_collections: "API Collections",
  test_plans: "Test Plans",
  users: "Users",
};

const sectionIcons: Record<string, string> = {
  projects: "📁",
  test_cases: "🧪",
  defects: "🐛",
  automation: "⚡",
  api_collections: "🌐",
  test_plans: "📋",
  users: "👤",
};

export function GlobalSearch() {
  const router = useRouter();
  const { workspace } = useWorkspace();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<SearchResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "/" && !open && !["INPUT", "TEXTAREA"].includes((e.target as HTMLElement)?.tagName)) {
        e.preventDefault();
        setOpen(true);
      }
      if (e.key === "Escape") setOpen(false);
    };
    const onOpen = () => setOpen(true);
    document.addEventListener("keydown", onKey);
    document.addEventListener("open-global-search", onOpen);
    return () => {
      document.removeEventListener("keydown", onKey);
      document.removeEventListener("open-global-search", onOpen);
    };
  }, [open]);

  useEffect(() => {
    if (open) inputRef.current?.focus();
  }, [open]);

  useEffect(() => {
    setResult(null);
    setError(null);
    setSelected(0);
    if (!query.trim() || !workspace) return;
    const timer = setTimeout(async () => {
      setLoading(true);
      try {
        const res = await globalSearch(workspace.id, query.trim());
        setResult(res);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Search failed");
      } finally {
        setLoading(false);
      }
    }, 200);
    return () => clearTimeout(timer);
  }, [query, workspace]);

  const flatItems = useMemo(() => {
    if (!result) return [];
    const items: { section: string; item: SearchItem }[] = [];
    (Object.keys(sectionLabels) as (keyof SearchResult)[]).forEach((key) => {
      const arr = result[key] as SearchItem[] | undefined;
      if (arr) {
        arr.forEach((item) => items.push({ section: key, item }));
      }
    });
    return items;
  }, [result]);

  useEffect(() => {
    setSelected(0);
  }, [flatItems.length]);

  const navigate = useCallback((item: SearchItem) => {
    setOpen(false);
    setQuery("");
    setResult(null);
    if (item.url) router.push(item.url);
  }, [router]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (!open) return;
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelected((s) => (s + 1) % flatItems.length);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelected((s) => (s - 1 + flatItems.length) % flatItems.length);
      } else if (e.key === "Enter") {
        e.preventDefault();
        const selectedItem = flatItems[selected];
        if (selectedItem) navigate(selectedItem.item);
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open, flatItems, selected, navigate]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[70] flex animate-fade-in items-start justify-center bg-[rgba(5,6,14,0.55)] p-4 pt-[12vh] backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-label="Global search"
    >
      <div className="w-full max-w-3xl animate-pop overflow-hidden rounded-[22px] border border-hair-hi bg-glass shadow-glass backdrop-blur-[30px] backdrop-saturate-150">
        <div className="flex items-center gap-3 border-b border-hair px-[18px] py-[15px]">
          <Search className="h-[17px] w-[17px] text-acc" aria-hidden="true" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search projects, test cases, defects, collections, plans, users..."
            className="flex-1 bg-transparent text-[15px] text-fg outline-none placeholder:text-fg3"
            aria-label="Global search input"
          />
          {loading && <Loader2 className="h-4 w-4 animate-spin text-fg3" />}
          <button onClick={() => setOpen(false)} className="text-fg3 transition-colors hover:text-fg" aria-label="Close">
            <X className="h-4 w-4" />
          </button>
        </div>
        <div className="max-h-[60vh] overflow-y-auto p-2">
          {error && <p className="p-4 text-[13px] text-fail">{error}</p>}
          {!query.trim() && <p className="p-[30px] text-center text-[13px] text-fg3">Type to search across Testra.</p>}
          {flatItems.length === 0 && query.trim() && !loading && !error && (
            <p className="p-[30px] text-center text-[13px] text-fg3">No results found.</p>
          )}
          {flatItems.length > 0 && (
            <ul role="listbox">
              {flatItems.map(({ section, item }, idx) => (
                <li key={`${section}-${item.id}`}>
                  <button
                    role="option"
                    aria-selected={idx === selected}
                    onClick={() => navigate(item)}
                    onMouseEnter={() => setSelected(idx)}
                    className={cn(
                      "flex w-full items-center gap-3 rounded-[13px] px-3 py-2.5 text-left transition-colors",
                      idx === selected ? "bg-acc-soft" : "hover:bg-panel",
                    )}
                  >
                    <span className="flex h-7 w-7 flex-none items-center justify-center rounded-lg bg-panel text-[13px]" aria-hidden="true">
                      {sectionIcons[section]}
                    </span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-[13.5px] font-medium text-fg">{item.title}</p>
                      <p className="truncate text-[11.5px] text-fg3">{sectionLabels[section]} · {item.subtitle}</p>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="flex items-center gap-4 border-t border-hair px-[18px] py-2.5 font-mono text-[10px] text-fg3">
          <span>↑↓ navigate</span>
          <span>↵ select</span>
          <span>&quot;/&quot; to search</span>
        </div>
      </div>
    </div>
  );
}
