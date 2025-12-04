// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useQuery } from '@tanstack/react-query';
import { statusApi, alertsApi } from '../services/api';
import { Activity, CheckCircle, AlertCircle, Database, TrendingUp } from 'lucide-react';
import { Area, AreaChart, CartesianGrid, XAxis, YAxis, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { useMemo } from 'react';

// InteractiveAreaChart component
function InteractiveAreaChart({ data }: { data: any[] }) {
  // Transform data for recharts
  const chartData = useMemo(() => {
    if (!data || data.length === 0) return [];

    // Get all time points from first rule
    const timePoints = data[0]?.data_points || [];

    // Transform to recharts format
    return timePoints.map((point: any, index: number) => {
      const dataPoint: any = {
        time: point.time,
      };

      // Add each rule's value at this time point
      data.forEach((rule: any) => {
        const value = rule.data_points[index]?.value || 0;
        dataPoint[rule.rule_name] = value;
      });

      return dataPoint;
    });
  }, [data]);

  // Colors for different rules
  const colors = [
    '#3b82f6',  // Blue
    '#a855f7',  // Purple
    '#10b981',  // Green
    '#f97316',  // Orange
    '#ec4899',  // Pink
    '#06b6d4',  // Cyan
    '#eab308',  // Yellow
    '#ef4444',  // Red
  ];

  return (
    <ResponsiveContainer width="100%" height={400}>
      <AreaChart
        data={chartData}
        margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
      >
        <defs>
          {data.map((rule: any, index: number) => {
            const color = colors[index % colors.length];
            return (
              <linearGradient key={rule.rule_id} id={`gradient-${rule.rule_id}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={color} stopOpacity={0.8} />
                <stop offset="100%" stopColor={color} stopOpacity={0.1} />
              </linearGradient>
            );
          })}
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" vertical={false} />
        <XAxis
          dataKey="time"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          tick={{ fill: 'hsl(var(--muted-foreground))' }}
        />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          tick={{ fill: 'hsl(var(--muted-foreground))' }}
          allowDecimals={false}
        />
        <Tooltip
          content={({ active, payload, label }) => {
            if (!active || !payload) return null;

            return (
              <div className="rounded-lg border bg-background p-2 shadow-md">
                <div className="text-sm font-medium mb-2">时间: {label}</div>
                <div className="space-y-1">
                  {payload.map((entry: any, index: number) => (
                    <div key={index} className="flex items-center gap-2 text-xs">
                      <div
                        className="w-3 h-3 rounded-sm"
                        style={{ backgroundColor: entry.color }}
                      />
                      <span className="flex-1">{entry.name}:</span>
                      <span className="font-mono font-semibold">{entry.value}</span>
                    </div>
                  ))}
                  <div className="border-t pt-1 mt-1 flex justify-between text-xs font-semibold">
                    <span>总计:</span>
                    <span className="font-mono">
                      {payload.reduce((sum: number, entry: any) => sum + (entry.value || 0), 0)}
                    </span>
                  </div>
                </div>
              </div>
            );
          }}
        />
        <Legend
          verticalAlign="top"
          height={36}
          content={({ payload }) => {
            if (!payload) return null;
            return (
              <div className="flex flex-wrap justify-center gap-4 pb-4">
                {payload.map((entry: any, index: number) => (
                  <div key={index} className="flex items-center gap-2 text-sm">
                    <div
                      className="w-3 h-3 rounded-sm"
                      style={{ backgroundColor: entry.color }}
                    />
                    <span>{entry.value}</span>
                  </div>
                ))}
              </div>
            );
          }}
        />
        {data.map((rule: any, index: number) => (
          <Area
            key={rule.rule_id}
            dataKey={rule.rule_name}
            type="monotone"
            fill={`url(#gradient-${rule.rule_id})`}
            stroke={colors[index % colors.length]}
            strokeWidth={2}
            stackId="1"
          />
        ))}
      </AreaChart>
    </ResponsiveContainer>
  );
}

export default function DashboardPage() {
  const { data: statusData, isLoading } = useQuery({
    queryKey: ['status'],
    queryFn: () => statusApi.getStatus().then(res => res.data.data),
    refetchInterval: 30000,
  });

  const { data: ruleTimeSeriesData, isLoading: ruleTimeSeriesLoading } = useQuery({
    queryKey: ['rule-timeseries-stats'],
    queryFn: () => alertsApi.getRuleTimeSeries('24h', 60).then(res => res.data.data),
    refetchInterval: 30000,
  });

  const stats = [
    {
      title: '总规则数',
      value: statusData?.rules.total || 0,
      icon: Activity,
      description: '已配置的规则总数',
    },
    {
      title: '启用规则',
      value: statusData?.rules.enabled || 0,
      icon: CheckCircle,
      description: '当前启用的规则',
    },
    {
      title: '24小时告警',
      value: statusData?.alerts_24h?.total || 0,
      icon: AlertCircle,
      description: '最近24小时告警数',
    },
    {
      title: 'ES 数据源',
      value: (() => {
        const es = statusData?.elasticsearch;
        if (!es || es.total === 0) return '未配置';
        if (es.success_count === es.total) {
          return `${es.success_count}/${es.total} 正常`;
        }
        if (es.success_count > 0) {
          return `${es.success_count}/${es.total} 正常`;
        }
        if (es.failed_count > 0) {
          return `${es.failed_count}/${es.total} 异常`;
        }
        return `${es.total} 个配置`;
      })(),
      icon: Database,
      description: (() => {
        const es = statusData?.elasticsearch;
        if (!es || es.total === 0) return '未配置数据源';
        return `正常: ${es.success_count || 0}, 异常: ${es.failed_count || 0}, 未测试: ${es.unknown_count || 0}`;
      })(),
      status: (() => {
        const es = statusData?.elasticsearch;
        if (!es || es.total === 0) return 'warning';
        const status = es.status;
        if (status === 'connected') return 'success';
        if (status === 'disconnected') return 'error';
        return 'warning';
      })(),
    },
  ];

  if (isLoading) {
    return <div className="text-center py-8">加载中...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
            系统概览
          </h2>
          <p className="text-muted-foreground mt-1">实时监控系统运行状态和关键指标</p>
        </div>
      </div>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <Card key={stat.title} className="border-border/50 shadow-md hover:shadow-lg transition-shadow duration-200 bg-card/80 backdrop-blur-sm">
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-3">
                <CardTitle className="text-sm font-semibold text-muted-foreground">
                  {stat.title}
                </CardTitle>
                <div className={`p-2 rounded-lg ${
                  stat.status === 'success' ? 'bg-green-100 text-green-600' :
                  stat.status === 'error' ? 'bg-red-100 text-red-600' :
                  stat.status === 'warning' ? 'bg-yellow-100 text-yellow-600' :
                  'bg-muted text-muted-foreground'
                }`}>
                  <Icon className="h-4 w-4" />
                </div>
              </CardHeader>
              <CardContent>
                <div className="text-3xl font-bold mb-2">
                  {stat.status === 'success' ? (
                    <span className="text-green-600">{stat.value}</span>
                  ) : stat.status === 'error' ? (
                    <span className="text-red-600">{stat.value}</span>
                  ) : stat.status === 'warning' ? (
                    <span className="text-yellow-600">{stat.value}</span>
                  ) : (
                    <span className="text-foreground">{stat.value}</span>
                  )}
                </div>
                <p className="text-xs text-muted-foreground leading-relaxed">
                  {stat.description}
                </p>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Rule Alert Time Series */}
      <div className="mt-6">
        <Card className="border-border/50 shadow-md bg-card/80 backdrop-blur-sm">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-muted-foreground" />
                <CardTitle>规则告警趋势 (24小时)</CardTitle>
              </div>
              <div className="text-sm text-muted-foreground">
                {ruleTimeSeriesData && ruleTimeSeriesData.length > 0 && (
                  <span>{ruleTimeSeriesData.length} 条规则</span>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {ruleTimeSeriesLoading ? (
              <div className="text-center py-8 text-muted-foreground">加载中...</div>
            ) : !ruleTimeSeriesData || ruleTimeSeriesData.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                暂无启用的规则
              </div>
            ) : (
              <div className="space-y-4">
                <InteractiveAreaChart data={ruleTimeSeriesData} />
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
