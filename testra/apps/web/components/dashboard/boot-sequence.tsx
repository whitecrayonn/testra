"use client";

import { useEffect, useState } from "react";
import { Check } from "lucide-react";
import { useWorkspace } from "@/lib/hooks/use-workspace";

const STEP_DEFS = [
  { label: "Establishing connection", hint: "Secure" },
  { label: "Verifying session", hint: "Authenticated" },
  { label: "Loading workspace", hint: null as string | null },
  { label: "Syncing test repository", hint: "Repository" },
  { label: "Ready", hint: "All set" },
];

const STEP_DURATION_MS = 380;
const TOTAL_MS = STEP_DURATION_MS * STEP_DEFS.length;
const SESSION_KEY = "testra_boot_shown";

export function BootSequence() {
  const { workspace } = useWorkspace();
  const [visible, setVisible] = useState(false);
  const [step, setStep] = useState(0);
  const [exiting, setExiting] = useState(false);

  function play() {
    setStep(0);
    setExiting(false);
    setVisible(true);
  }

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (!sessionStorage.getItem(SESSION_KEY)) {
      play();
    }
    const onReplay = () => play();
    document.addEventListener("replay-boot-sequence", onReplay);
    return () => document.removeEventListener("replay-boot-sequence", onReplay);
  }, []);

  useEffect(() => {
    if (!visible) return;
    const timers = STEP_DEFS.map((_, i) => setTimeout(() => setStep(i + 1), STEP_DURATION_MS * (i + 1)));
    const exitTimer = setTimeout(() => {
      setExiting(true);
      sessionStorage.setItem(SESSION_KEY, "1");
      setTimeout(() => setVisible(false), 450);
    }, TOTAL_MS + 300);
    return () => {
      timers.forEach(clearTimeout);
      clearTimeout(exitTimer);
    };
  }, [visible]);

  if (!visible) return null;

  const pct = Math.min(100, (step / STEP_DEFS.length) * 100);

  return (
    <div
      className="fixed inset-0 z-[90] flex flex-col items-center justify-center gap-6 transition-all duration-500"
      style={{
        background:
          "radial-gradient(900px 620px at 50% 26%, var(--blob1), transparent 62%), var(--bg)",
        opacity: exiting ? 0 : 1,
        transform: exiting ? "scale(1.02)" : "scale(1)",
      }}
    >
      <div className="flex h-[62px] w-[62px] items-end justify-center gap-1.5 rounded-[20px] bg-gradient-to-br from-acc to-acc2 pb-4 shadow-[0_20px_50px_-14px_var(--ring)]">
        <span className="h-[14px] w-[5px] animate-boot-bar rounded-sm bg-white" style={{ animationDelay: "0ms" }} />
        <span className="h-[26px] w-[5px] animate-boot-bar rounded-sm bg-white" style={{ animationDelay: "160ms" }} />
        <span className="h-[20px] w-[5px] animate-boot-bar rounded-sm bg-white" style={{ animationDelay: "320ms" }} />
      </div>
      <div className="text-center">
        <div className="text-[19px] font-bold tracking-[0.34em] text-fg">TESTRA</div>
        <div className="mt-1.5 font-mono text-[10px] tracking-[0.16em] text-fg3">ONE PLATFORM · EVERY TEST</div>
      </div>
      <div className="w-80">
        <div className="h-[3px] overflow-hidden rounded-full bg-hair">
          <div
            className="h-full rounded-full bg-gradient-to-r from-acc to-acc2 transition-[width] duration-500 ease-[cubic-bezier(.2,.8,.2,1)]"
            style={{ width: `${pct}%` }}
          />
        </div>
        <div className="mt-5 flex flex-col gap-2.5">
          {STEP_DEFS.map((s, i) => {
            const done = i < step;
            const active = i === step;
            return (
              <div
                key={s.label}
                className="flex items-center gap-2.5 text-[12.5px] transition-opacity duration-300"
                style={{ color: "var(--fg)", opacity: done || active ? 1 : 0.4 }}
              >
                <span
                  className="flex h-4 w-4 flex-none items-center justify-center rounded-full border text-[#0a0b12]"
                  style={{
                    borderColor: done ? "var(--acc)" : "var(--hair-hi)",
                    background: done ? "var(--acc)" : "transparent",
                  }}
                >
                  {done && <Check className="h-2.5 w-2.5" strokeWidth={3} />}
                </span>
                <span className="flex-1">{s.label}</span>
                <span className="font-mono text-[10px] text-fg3">
                  {i === 2 ? (workspace?.name ?? "Workspace") : s.hint}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
