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
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-slate-900/50 p-4 pt-24" role="dialog" aria-modal="true" aria-label="Global search">
      <div className="w-full max-w-3xl overflow-hidden rounded-xl bg-white shadow-2xl dark:bg-slate-900">
        <div className="flex items-center gap-3 border-b border-slate-100 px-4 py-3 dark:border-slate-800">
          <Search className="h-5 w-5 text-slate-400" aria-hidden="true" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search projects, test cases, defects, collections, plans, users..."
            className="flex-1 bg-transparent text-slate-900 outline-none placeholder:text-slate-400 dark:text-white"
            aria-label="Global search input"
          />
          {loading && <Loader2 className="h-5 w-5 animate-spin text-slate-400" />}
          <button onClick={() => setOpen(false)} className="text-slate-400 hover:text-slate-600" aria-label="Close">
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="max-h-[60vh] overflow-y-auto p-2">
          {error && <p className="p-4 text-sm text-red-600">{error}</p>}
          {!query.trim() && <p className="p-4 text-center text-sm text-slate-500">Type to search across Testra.</p>}
          {flatItems.length === 0 && query.trim() && !loading && !error && (
            <p className="p-4 text-center text-sm text-slate-500">No results found.</p>
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
                      "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left",
                      idx === selected ? "bg-brand-50 text-brand-900 dark:bg-slate-800" : "hover:bg-slate-50 dark:hover:bg-slate-800",
                    )}
                  >
                    <span aria-hidden="true">{sectionIcons[section]}</span>
                    <div className="flex-1">
                      <p className="text-sm font-medium text-slate-900 dark:text-slate-100">{item.title}</p>
                      <p className="text-xs text-slate-500">{sectionLabels[section]} · {item.subtitle}</p>
                    </div>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="flex items-center justify-between border-t border-slate-100 px-4 py-2 text-xs text-slate-500 dark:border-slate-800">
          <span>↑↓ navigate</span>
          <span>↵ select</span>
          <span>&quot;/&quot; to search</span>
        </div>
      </div>
    </div>
  );
}
