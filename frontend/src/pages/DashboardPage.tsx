// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useQuery } from '@tanstack/react-query';
import { statusApi, alertsApi } from '../services/api';
import { Activity, CheckCircle, AlertCircle, Database, TrendingUp } from 'lucide-react';
import { ChartContainer, ChartTooltip, ChartTooltipContent, ChartLegend, ChartLegendContent } from '@/components/ui/chart';
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from 'recharts';
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

  // Generate chart config dynamically
  const chartConfig = useMemo(() => {
    if (!data || data.length === 0) return {};

    const colors = [
      'hsl(221, 83%, 53%)',  // Blue
      'hsl(262, 83%, 58%)',  // Purple
      'hsl(142, 71%, 45%)',  // Green
      'hsl(25, 95%, 53%)',   // Orange
      'hsl(330, 81%, 60%)',  // Pink
      'hsl(199, 89%, 48%)',  // Cyan
      'hsl(48, 96%, 53%)',   // Yellow
      'hsl(348, 83%, 47%)',  // Red
    ];

    const config: any = {};
    data.forEach((rule: any, index: number) => {
      config[rule.rule_name] = {
        label: rule.rule_name,
        color: colors[index % colors.length],
      };
    });

    return config;
  }, [data]);

  return (
    <ChartContainer config={chartConfig} className="h-[400px] w-full">
      <AreaChart
        data={chartData}
        margin={{ top: 10, right: 10, left: 0, bottom: 0 }}
      >
        <defs>
          {data.map((rule: any, index: number) => {
            const colors = [
              ['#3b82f6', '#1d4ed8'],  // Blue
              ['#a855f7', '#7e22ce'],  // Purple
              ['#10b981', '#059669'],  // Green
              ['#f97316', '#ea580c'],  // Orange
              ['#ec4899', '#db2777'],  // Pink
              ['#06b6d4', '#0891b2'],  // Cyan
              ['#eab308', '#ca8a04'],  // Yellow
              ['#ef4444', '#dc2626'],  // Red
            ];
            const [startColor, endColor] = colors[index % colors.length];

            return (
              <linearGradient key={rule.rule_id} id={`fill-${rule.rule_id}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={startColor} stopOpacity={0.8} />
                <stop offset="100%" stopColor={endColor} stopOpacity={0.1} />
              </linearGradient>
            );
          })}
        </defs>
        <CartesianGrid strokeDasharray="3 3" vertical={false} />
        <XAxis
          dataKey="time"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          tickFormatter={(value) => value}
        />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          tickFormatter={(value) => `${value}`}
        />
        <ChartTooltip
          content={
            <ChartTooltipContent
              labelFormatter={(value) => `时间: ${value}`}
              indicator="line"
            />
          }
        />
        <ChartLegend content={<ChartLegendContent />} />
        {data.map((rule: any) => (
          <Area
            key={rule.rule_id}
            dataKey={rule.rule_name}
            type="monotone"
            fill={`url(#fill-${rule.rule_id})`}
            fillOpacity={0.4}
            stroke={`var(--color-${rule.rule_name})`}
            strokeWidth={2}
            stackId="1"
          />
        ))}
      </AreaChart>
    </ChartContainer>
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
