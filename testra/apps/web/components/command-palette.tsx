"use client";

import { useEffect, useState, useMemo, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { Search, X, Home, FolderKanban, FlaskConical, Bug, Zap, Globe, FileText, PlayCircle, Bell, Settings, Plus } from "lucide-react";
import { cn } from "@/lib/utils";

interface Command {
  id: string;
  title: string;
  shortcut?: string;
  icon: React.ReactNode;
  href?: string;
  onClick?: () => void;
  section: string;
}

export function CommandPalette() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [selected, setSelected] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const commands: Command[] = useMemo(
    () => [
      { id: "home", title: "Go to Dashboard", icon: <Home className="h-4 w-4" />, href: "/dashboard", section: "Navigation" },
      { id: "projects", title: "Go to Projects", icon: <FolderKanban className="h-4 w-4" />, href: "/dashboard/projects", section: "Navigation" },
      { id: "test-cases", title: "Go to Test Cases", icon: <FlaskConical className="h-4 w-4" />, href: "/dashboard/test-cases", section: "Navigation" },
      { id: "defects", title: "Go to Defects", icon: <Bug className="h-4 w-4" />, href: "/dashboard/defects", section: "Navigation" },
      { id: "automation", title: "Go to Automation", icon: <Zap className="h-4 w-4" />, href: "/dashboard/automation", section: "Navigation" },
      { id: "api-tests", title: "Go to API Tests", icon: <Globe className="h-4 w-4" />, href: "/dashboard/api-tests", section: "Navigation" },
      { id: "test-plans", title: "Go to Test Plans", icon: <FileText className="h-4 w-4" />, href: "/dashboard/test-plans", section: "Navigation" },
      { id: "test-runs", title: "Go to Test Runs", icon: <PlayCircle className="h-4 w-4" />, href: "/dashboard/test-runs", section: "Navigation" },
      { id: "notifications", title: "Go to Notifications", icon: <Bell className="h-4 w-4" />, href: "/dashboard/notifications", section: "Navigation" },
      { id: "settings", title: "Go to Settings", icon: <Settings className="h-4 w-4" />, href: "/dashboard/settings", section: "Navigation" },
      { id: "new-project", title: "Create new project", icon: <Plus className="h-4 w-4" />, href: "/dashboard/projects", section: "Actions" },
      { id: "new-test-case", title: "Create new test case", icon: <Plus className="h-4 w-4" />, href: "/dashboard/test-cases", section: "Actions" },
      { id: "new-defect", title: "Create new defect", icon: <Plus className="h-4 w-4" />, href: "/dashboard/defects", section: "Actions" },
      { id: "search", title: "Open global search", icon: <Search className="h-4 w-4" />, onClick: () => {
        setOpen(false);
        document.dispatchEvent(new CustomEvent("open-global-search"));
      }, section: "Actions" },
    ],
    [],
  );

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return commands;
    return commands.filter((c) => c.title.toLowerCase().includes(q) || c.section.toLowerCase().includes(q));
  }, [commands, query]);

  useEffect(() => {
    setSelected(0);
  }, [query]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setOpen((o) => !o);
      }
      if (e.key === "Escape") setOpen(false);
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, []);

  useEffect(() => {
    if (open) {
      inputRef.current?.focus();
    }
  }, [open]);

  const run = useCallback((cmd: Command) => {
    setOpen(false);
    setQuery("");
    if (cmd.onClick) {
      cmd.onClick();
    } else if (cmd.href) {
      router.push(cmd.href);
    }
  }, [router]);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (!open) return;
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelected((s) => (s + 1) % filtered.length);
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelected((s) => (s - 1 + filtered.length) % filtered.length);
      } else if (e.key === "Enter") {
        e.preventDefault();
        const cmd = filtered[selected];
        if (cmd) run(cmd);
      }
    };
    document.addEventListener("keydown", handler);
    return () => document.removeEventListener("keydown", handler);
  }, [open, filtered, selected, run]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-slate-900/50 p-4 pt-24" role="dialog" aria-modal="true" aria-label="Command palette">
      <div className="w-full max-w-2xl overflow-hidden rounded-xl bg-white shadow-2xl dark:bg-slate-900">
        <div className="flex items-center gap-3 border-b border-slate-100 px-4 py-3 dark:border-slate-800">
          <Search className="h-5 w-5 text-slate-400" aria-hidden="true" />
          <input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Type a command or search..."
            className="flex-1 bg-transparent text-slate-900 outline-none placeholder:text-slate-400 dark:text-white"
            aria-label="Command palette search"
          />
          <button onClick={() => setOpen(false)} className="text-slate-400 hover:text-slate-600" aria-label="Close">
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="max-h-96 overflow-y-auto p-2">
          {filtered.length === 0 ? (
            <p className="p-4 text-center text-sm text-slate-500">No commands found.</p>
          ) : (
            <ul role="listbox">
              {filtered.map((cmd, idx) => (
                <li key={cmd.id}>
                  <button
                    role="option"
                    aria-selected={idx === selected}
                    onClick={() => run(cmd)}
                    onMouseEnter={() => setSelected(idx)}
                    className={cn(
                      "flex w-full items-center gap-3 rounded-lg px-3 py-2 text-left text-sm",
                      idx === selected ? "bg-brand-50 text-brand-900 dark:bg-slate-800" : "text-slate-700 hover:bg-slate-50 dark:text-slate-300",
                    )}
                  >
                    <span className="text-slate-500">{cmd.icon}</span>
                    <span className="flex-1">{cmd.title}</span>
                    <span className="text-xs text-slate-400">{cmd.section}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="flex items-center justify-between border-t border-slate-100 px-4 py-2 text-xs text-slate-500 dark:border-slate-800">
          <span>↑↓ to navigate</span>
          <span>↵ to select</span>
          <span>Ctrl+K to toggle</span>
        </div>
      </div>
    </div>
  );
}
