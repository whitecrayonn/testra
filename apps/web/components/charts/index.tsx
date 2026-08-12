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
  tick: { fill: "var(--fg3)", fontSize: 12 },
  axisLine: { stroke: "var(--hair-hi)" },
  tickLine: { stroke: "var(--hair-hi)" },
};

const gridProps = { stroke: "var(--hair)" };

const tooltipStyle = {
  backgroundColor: "var(--panel-hi)",
  color: "var(--fg)",
  border: "1px solid var(--hair-hi)",
  borderRadius: 12,
  fontSize: 12.5,
};

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
          <Tooltip contentStyle={tooltipStyle} />
          <Legend />
          {children ?? (yKey ? <Line type="monotone" dataKey={yKey} stroke="var(--acc)" strokeWidth={2} dot={false} /> : null)}
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
          <Tooltip contentStyle={tooltipStyle} />
          <Legend />
          {children ?? (yKey ? <Bar dataKey={yKey} fill="var(--acc)" radius={[4, 4, 0, 0]} /> : null)}
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
          <Tooltip contentStyle={tooltipStyle} />
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
          <Tooltip contentStyle={tooltipStyle} />
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
          <Tooltip contentStyle={tooltipStyle} />
          <Area type="monotone" dataKey={yKey ?? "value"} stroke="var(--acc)" fill="var(--acc)" fillOpacity={0.2} />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}

export const chartColors = [
  "var(--acc)",
  "var(--acc2)",
  "var(--pass)",
  "var(--fail)",
  "var(--warn)",
  "var(--info)",
  "var(--fg2)",
  "var(--fg3)",
];
