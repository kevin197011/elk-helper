// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useQuery } from '@tanstack/react-query';
import { statusApi, alertsApi } from '../services/api';
import { Activity, CheckCircle, AlertCircle, Database, TrendingUp } from 'lucide-react';

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
                  <span>TOP {ruleTimeSeriesData.length} 规则</span>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {ruleTimeSeriesLoading ? (
              <div className="text-center py-8 text-muted-foreground">加载中...</div>
            ) : !ruleTimeSeriesData || ruleTimeSeriesData.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                最近24小时暂无告警记录
              </div>
            ) : (
              <div className="space-y-6">
                {ruleTimeSeriesData.map((ruleStat, ruleIndex) => {
                  const maxValue = Math.max(...ruleStat.data_points.map(p => p.value), 1);
                  const colors = [
                    'from-blue-500 to-blue-600',
                    'from-purple-500 to-purple-600',
                    'from-green-500 to-green-600',
                    'from-orange-500 to-orange-600',
                    'from-pink-500 to-pink-600',
                  ];
                  const colorClass = colors[ruleIndex % colors.length];

                  return (
                    <div key={ruleStat.rule_id} className="space-y-2">
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <div className={`w-3 h-3 rounded-full bg-gradient-to-r ${colorClass}`} />
                          <span className="font-medium">{ruleStat.rule_name}</span>
                        </div>
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <AlertCircle className="h-3.5 w-3.5" />
                          <span>共 {ruleStat.total} 次告警</span>
                        </div>
                      </div>

                      {/* Time series chart */}
                      <div className="relative h-24 flex items-end gap-1 bg-muted/30 rounded-lg p-2">
                        {ruleStat.data_points.map((point, index) => {
                          const height = point.value > 0 ? (point.value / maxValue * 100) : 0;
                          const isLast = index === ruleStat.data_points.length - 1;
                          const showLabel = index % Math.ceil(ruleStat.data_points.length / 6) === 0 || isLast;

                          return (
                            <div key={index} className="flex-1 flex flex-col items-center justify-end group relative">
                              <div
                                className={`w-full bg-gradient-to-t ${colorClass} rounded-t transition-all duration-300 hover:opacity-80`}
                                style={{ height: `${height}%`, minHeight: point.value > 0 ? '4px' : '0' }}
                                title={`${point.time}: ${point.value} 次告警`}
                              />
                              {point.value > 0 && (
                                <div className="absolute -top-5 opacity-0 group-hover:opacity-100 transition-opacity bg-popover text-popover-foreground text-xs px-1.5 py-0.5 rounded shadow-md whitespace-nowrap z-10">
                                  {point.value}
                                </div>
                              )}
                              {showLabel && (
                                <div className="absolute -bottom-5 text-[10px] text-muted-foreground">
                                  {point.time}
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                      <div className="h-4" /> {/* Spacer for time labels */}
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
