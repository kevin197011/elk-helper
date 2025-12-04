// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { useQuery } from '@tanstack/react-query';
import { statusApi, alertsApi } from '../services/api';
import { Activity, CheckCircle, AlertCircle, Database, TrendingUp, Clock } from 'lucide-react';

export default function DashboardPage() {
  const { data: statusData, isLoading } = useQuery({
    queryKey: ['status'],
    queryFn: () => statusApi.getStatus().then(res => res.data.data),
    refetchInterval: 30000,
  });

  const { data: ruleStatsData, isLoading: ruleStatsLoading } = useQuery({
    queryKey: ['rule-alert-stats'],
    queryFn: () => alertsApi.getRuleStats('24h').then(res => res.data.data),
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

      {/* Rule Alert Statistics */}
      <div className="mt-6">
        <Card className="border-border/50 shadow-md bg-card/80 backdrop-blur-sm">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <TrendingUp className="h-5 w-5 text-muted-foreground" />
                <CardTitle>规则告警统计 (24小时)</CardTitle>
              </div>
              <div className="text-sm text-muted-foreground">
                {ruleStatsData && ruleStatsData.length > 0 && (
                  <span>共 {ruleStatsData.length} 条规则产生告警</span>
                )}
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {ruleStatsLoading ? (
              <div className="text-center py-8 text-muted-foreground">加载中...</div>
            ) : !ruleStatsData || ruleStatsData.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                最近24小时暂无告警记录
              </div>
            ) : (
              <div className="space-y-3">
                {ruleStatsData.slice(0, 10).map((stat, index) => {
                  const successRate = stat.total > 0 ? (stat.sent / stat.total * 100).toFixed(1) : '0';
                  const maxTotal = Math.max(...ruleStatsData.map(s => s.total));
                  const barWidth = stat.total > 0 ? (stat.total / maxTotal * 100) : 0;

                  return (
                    <div key={stat.rule_id} className="group hover:bg-muted/50 p-3 rounded-lg transition-colors">
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center gap-3 flex-1">
                          <span className="text-xs font-mono text-muted-foreground w-6">#{index + 1}</span>
                          <span className="font-medium truncate max-w-xs">{stat.rule_name}</span>
                        </div>
                        <div className="flex items-center gap-4 text-sm">
                          <div className="flex items-center gap-1">
                            <AlertCircle className="h-3.5 w-3.5 text-orange-500" />
                            <span className="font-semibold">{stat.total}</span>
                            <span className="text-muted-foreground">次</span>
                          </div>
                          {stat.last_alert && (
                            <div className="flex items-center gap-1 text-muted-foreground">
                              <Clock className="h-3.5 w-3.5" />
                              <span className="text-xs">
                                {new Date(stat.last_alert).toLocaleString('zh-CN', {
                                  month: '2-digit',
                                  day: '2-digit',
                                  hour: '2-digit',
                                  minute: '2-digit'
                                })}
                              </span>
                            </div>
                          )}
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        <div className="flex-1 h-2 bg-muted rounded-full overflow-hidden">
                          <div
                            className="h-full bg-gradient-to-r from-orange-500 to-red-500 transition-all duration-500"
                            style={{ width: `${barWidth}%` }}
                          />
                        </div>
                        <div className="flex gap-4 text-xs">
                          <span className="text-green-600">
                            ✓ {stat.sent}
                          </span>
                          {stat.failed > 0 && (
                            <span className="text-red-600">
                              ✗ {stat.failed}
                            </span>
                          )}
                          <span className="text-muted-foreground">
                            成功率 {successRate}%
                          </span>
                        </div>
                      </div>
                    </div>
                  );
                })}
                {ruleStatsData.length > 10 && (
                  <div className="text-center py-2 text-sm text-muted-foreground">
                    还有 {ruleStatsData.length - 10} 条规则未显示
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
