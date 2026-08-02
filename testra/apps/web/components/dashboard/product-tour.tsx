"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { usePathname } from "next/navigation";

interface TourStep {
  selector: string;
  title: string;
  body: string;
}

const STEPS: TourStep[] = [
  {
    selector: '[data-tour="nav"]',
    title: "Everything, one keystroke away",
    body: "The rail groups work by what you're doing — workspace, quality, system — and the accent marker follows you as you move.",
  },
  {
    selector: '[data-tour="hero"]',
    title: "The numbers that matter first",
    body: "Release readiness and what's running right now, before anything else on the page.",
  },
  {
    selector: '[data-tour="search"]',
    title: "Command palette",
    body: "Press Ctrl+K anywhere to jump, create a case, start a run, or flip the theme. No hunting through menus.",
  },
];

interface Rect {
  top: number;
  left: number;
  width: number;
  height: number;
}

function measure(selector: string): Rect | null {
  const el = document.querySelector<HTMLElement>(selector);
  if (!el) return null;
  const r = el.getBoundingClientRect();
  return { top: r.top, left: r.left, width: r.width, height: r.height };
}

export function ProductTour() {
  const pathname = usePathname();
  const [stepIndex, setStepIndex] = useState<number | null>(null);
  const [rect, setRect] = useState<Rect | null>(null);

  const cancelMeasureRef = useRef<(() => void) | null>(null);

  const measureCurrent = useCallback((index: number) => {
    // Only the latest retry chain may ever call setRect — cancel whatever
    // was still in flight for a previous step before starting a new one.
    cancelMeasureRef.current?.();

    let cancelled = false;
    let timeoutId: ReturnType<typeof setTimeout> | undefined;

    // Retry briefly in case the target hasn't painted yet (e.g. just navigated).
    function tryMeasure(attempts: number) {
      if (cancelled) return;
      const r = measure(STEPS[index].selector);
      if (r) {
        setRect(r);
      } else if (attempts < 10) {
        timeoutId = setTimeout(() => tryMeasure(attempts + 1), 100);
      } else {
        setRect(null);
      }
    }
    tryMeasure(0);

    cancelMeasureRef.current = () => {
      cancelled = true;
      if (timeoutId) clearTimeout(timeoutId);
    };
  }, []);

  useEffect(() => {
    function onStart() {
      setStepIndex(0);
    }
    document.addEventListener("start-product-tour", onStart);
    return () => document.removeEventListener("start-product-tour", onStart);
  }, []);

  useEffect(() => {
    if (stepIndex === null) {
      cancelMeasureRef.current?.();
      cancelMeasureRef.current = null;
      return;
    }
    measureCurrent(stepIndex);
    return () => {
      cancelMeasureRef.current?.();
      cancelMeasureRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stepIndex, pathname]);

  useEffect(() => {
    if (stepIndex === null) return;
    const currentStep = stepIndex;
    function onResize() {
      measureCurrent(currentStep);
    }
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, [stepIndex, measureCurrent]);

  if (stepIndex === null) return null;

  function next() {
    if (stepIndex === null) return;
    if (stepIndex >= STEPS.length - 1) {
      setStepIndex(null);
      return;
    }
    setStepIndex(stepIndex + 1);
  }

  function end() {
    setStepIndex(null);
  }

  const step = STEPS[stepIndex];
  const pad = 8;
  const tipWidth = 290;
  const viewportW = typeof window !== "undefined" ? window.innerWidth : 1280;
  const viewportH = typeof window !== "undefined" ? window.innerHeight : 800;
  let tipX = rect ? rect.left + rect.width + 20 : viewportW / 2 - tipWidth / 2;
  let tipY = rect ? rect.top : viewportH / 2 - 80;
  if (rect && tipX + tipWidth > viewportW - 20) {
    tipX = Math.max(20, rect.left);
    tipY = rect.top + rect.height + 16;
  }
  tipY = Math.min(Math.max(20, tipY), viewportH - 220);

  return (
    <div className="fixed inset-0 z-[82]">
      <div className="absolute inset-0" onClick={next} />
      {rect && (
        <div
          className="pointer-events-none absolute rounded-2xl transition-all duration-300"
          style={{
            left: rect.left - pad,
            top: rect.top - pad,
            width: rect.width + pad * 2,
            height: rect.height + pad * 2,
            boxShadow: "0 0 0 2px var(--acc), 0 0 0 9999px rgba(4,5,12,.66)",
          }}
        />
      )}
      <div
        className="absolute w-[290px] animate-pop rounded-[18px] border border-hair-hi bg-glass p-[18px] shadow-glass backdrop-blur-[24px] backdrop-saturate-150 transition-all duration-300"
        style={{ left: tipX, top: tipY }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="font-mono text-[10px] tracking-[0.14em] text-acc">
          {stepIndex + 1} / {STEPS.length}
        </div>
        <h3 className="m-0 mb-1.5 mt-2 text-[15px] font-semibold tracking-tight text-fg">{step.title}</h3>
        <p className="m-0 text-[12.5px] leading-relaxed text-fg2">{step.body}</p>
        <div className="mt-4 flex items-center justify-between">
          <button onClick={end} className="border-0 bg-transparent text-xs text-fg3">
            Skip
          </button>
          <button
            onClick={next}
            className="h-8 rounded-[11px] bg-gradient-to-br from-acc to-acc2 px-[15px] text-[12.5px] font-bold text-white"
          >
            {stepIndex === STEPS.length - 1 ? "Done" : "Next"}
          </button>
        </div>
      </div>
    </div>
  );
}
