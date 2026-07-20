"use client";

import type { ReactNode } from "react";
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Legend,
  Line,
  LineChart,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

const axisProps = {
  tick: { fill: "currentColor", fontSize: 12 },
  axisLine: { stroke: "currentColor" },
  tickLine: { stroke: "currentColor" },
};

const gridProps = { stroke: "rgba(148, 163, 184, 0.2)" };

interface ChartProps {
  data: unknown[];
  xKey: string;
  yKey?: string;
  children?: ReactNode;
  className?: string;
}

export function LineChartComponent({ data, xKey, yKey, children, className }: ChartProps) {
  return (
    <div className={className}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <CartesianGrid {...gridProps} />
          <XAxis dataKey={xKey} {...axisProps} />
          <YAxis {...axisProps} />
          <Tooltip contentStyle={{ backgroundColor: "var(--background)", color: "var(--foreground)", border: "1px solid #e2e8f0" }} />
          <Legend />
          {children ?? (yKey ? <Line type="monotone" dataKey={yKey} stroke="#4f46e5" strokeWidth={2} dot={false} /> : null)}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

export function BarChartComponent({ data, xKey, yKey, children, className }: ChartProps) {
  return (
    <div className={className}>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data}>
          <CartesianGrid {...gridProps} />
          <XAxis dataKey={xKey} {...axisProps} />
          <YAxis {...axisProps} />
          <Tooltip contentStyle={{ backgroundColor: "var(--background)", color: "var(--foreground)", border: "1px solid #e2e8f0" }} />
          <Legend />
          {children ?? (yKey ? <Bar dataKey={yKey} fill="#4f46e5" radius={[4, 4, 0, 0]} /> : null)}
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

export function StackedBarChart({ data, xKey, keys, colors, className }: ChartProps & { keys: string[]; colors: string[] }) {
  return (
    <div className={className}>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data}>
          <CartesianGrid {...gridProps} />
          <XAxis dataKey={xKey} {...axisProps} />
          <YAxis {...axisProps} />
          <Tooltip contentStyle={{ backgroundColor: "var(--background)", color: "var(--foreground)", border: "1px solid #e2e8f0" }} />
          <Legend />
          {keys.map((k, i) => (
            <Bar key={k} dataKey={k} stackId="a" fill={colors[i % colors.length]} radius={[0, 0, 0, 0]} />
          ))}
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

export function PieChartComponent({ data, nameKey, dataKey, colors, className, innerRadius }: ChartProps & { nameKey: string; dataKey: string; colors: string[]; innerRadius?: number }) {
  return (
    <div className={className}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie data={data} dataKey={dataKey ?? "value"} nameKey={nameKey} cx="50%" cy="50%" outerRadius={80} innerRadius={innerRadius} label>
            {data.map((_, i) => (
              <Cell key={`cell-${i}`} fill={colors[i % colors.length]} />
            ))}
          </Pie>
          <Tooltip contentStyle={{ backgroundColor: "var(--background)", color: "var(--foreground)", border: "1px solid #e2e8f0" }} />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}

export function AreaChartComponent({ data, xKey, yKey, className }: ChartProps) {
  return (
    <div className={className}>
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data}>
          <CartesianGrid {...gridProps} />
          <XAxis dataKey={xKey} {...axisProps} />
          <YAxis {...axisProps} />
          <Tooltip contentStyle={{ backgroundColor: "var(--background)", color: "var(--foreground)", border: "1px solid #e2e8f0" }} />
          <Area type="monotone" dataKey={yKey ?? "value"} stroke="#4f46e5" fill="#4f46e5" fillOpacity={0.2} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}

export const chartColors = ["#4f46e5", "#22c55e", "#ef4444", "#eab308", "#f97316", "#a855f7", "#06b6d4", "#64748b"];
