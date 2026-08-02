interface SegmentedBarProps {
  passed: number;
  failed: number;
  skipped: number;
  blocked?: number;
  total: number;
  className?: string;
}

export function SegmentedBar({ passed, failed, skipped, blocked = 0, total, className }: SegmentedBarProps) {
  const safeTotal = total > 0 ? total : 1;
  const pct = (n: number) => `${(n / safeTotal) * 100}%`;

  return (
    <span className={`flex h-1.5 overflow-hidden rounded-full bg-hair ${className ?? ""}`}>
      <span style={{ width: pct(passed) }} className="bg-pass" />
      <span style={{ width: pct(failed) }} className="bg-fail" />
      <span style={{ width: pct(skipped) }} className="bg-warn" />
      {blocked > 0 && <span style={{ width: pct(blocked) }} className="bg-info" />}
    </span>
  );
}
