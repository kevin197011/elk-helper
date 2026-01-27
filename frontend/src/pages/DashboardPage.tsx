// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { Card, Row, Col, Statistic, Spin, Typography, theme } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { statusApi, alertsApi } from '../services/api';
import {
  DashboardOutlined,
  CheckCircleOutlined,
  AlertOutlined,
  DatabaseOutlined,
  SyncOutlined,
  ClockCircleOutlined,
  LineChartOutlined,
} from '@ant-design/icons';
import { Area, AreaChart, CartesianGrid, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { useMemo } from 'react';
import PageHeader from '../components/PageHeader';

const { Text } = Typography;

function formatTooltipLabel(label: any) {
  // label 可能是 ISO 字符串或已格式化字符串；尽量输出“YYYY-MM-DD HH:mm”
  const d = new Date(label);
  if (!Number.isNaN(d.getTime())) {
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, '0');
    const dd = String(d.getDate()).padStart(2, '0');
    const hh = String(d.getHours()).padStart(2, '0');
    const mi = String(d.getMinutes()).padStart(2, '0');
    return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
  }
  return String(label ?? '');
}

// InteractiveAreaChart component
function InteractiveAreaChart({ data }: { data: any[] }) {
  const { token } = theme.useToken();
  const chartData = useMemo(() => {
    if (!data || data.length === 0) return [];

    // Collect all unique time points from all rules
    // Since different rules may have different bucket intervals, we need to merge all time points
    const timePointSet = new Set<string>();
    data.forEach((rule: any) => {
      rule.data_points?.forEach((point: any) => {
        if (point.time) {
          timePointSet.add(point.time);
        }
      });
    });

    // Sort time points by actual time (HH:mm format), not by string order
    // This ensures correct chronological order for rolling 24-hour window
    const sortedTimePoints = Array.from(timePointSet).sort((a: string, b: string) => {
      // Parse HH:mm format to compare as time
      const parseTime = (timeStr: string): number => {
        const [hours, minutes] = timeStr.split(':').map(Number);
        return hours * 60 + (minutes || 0); // Convert to minutes for comparison
      };
      return parseTime(a) - parseTime(b);
    });

    // Create a map for each rule: time -> value
    const ruleValueMap = new Map<string, Map<string, number>>();
    data.forEach((rule: any) => {
      const valueMap = new Map<string, number>();
      rule.data_points?.forEach((point: any) => {
        if (point.time) {
          valueMap.set(point.time, point.value || 0);
        }
      });
      ruleValueMap.set(rule.rule_name, valueMap);
    });

    // Build chart data points
    return sortedTimePoints.map((time: string) => {
      const dataPoint: any = { time };
      data.forEach((rule: any) => {
        const valueMap = ruleValueMap.get(rule.rule_name);
        dataPoint[rule.rule_name] = valueMap?.get(time) || 0;
      });
      return dataPoint;
    });
  }, [data]);

  const colors = [
    token.colorPrimary, '#7c3aed', '#16a34a', '#ea580c',
    '#eb2f96', '#13c2c2', '#faad14', '#f5222d',
  ];

  return (
    <ResponsiveContainer width="100%" height={400}>
      <AreaChart data={chartData} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
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
        <CartesianGrid strokeDasharray="3 3" stroke={token.colorBorderSecondary} vertical={false} />
        <XAxis
          dataKey="time"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          tick={{ fill: token.colorTextSecondary }}
          tickFormatter={(v: string) => {
            // v 是后端返回的 HH:mm 格式字符串（例如 "14:30"）
            // 直接返回，因为已经是正确的格式
            if (v && typeof v === 'string' && /^\d{2}:\d{2}$/.test(v)) {
              return v;
            }
            // 如果不是 HH:mm 格式，尝试解析
            const d = new Date(v);
            if (!Number.isNaN(d.getTime())) {
              return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
            }
            return String(v);
          }}
        />
        <YAxis tickLine={false} axisLine={false} tickMargin={8} tick={{ fill: token.colorTextSecondary }} allowDecimals={false} />
        <Tooltip
          content={({ active, payload, label }) => {
            if (!active || !payload) return null;
            // 显示所有规则，包括值为0的规则，按值降序排序
            const allRules = [...payload].sort((a: any, b: any) => (b.value || 0) - (a.value || 0));
            const total = payload.reduce((sum: number, entry: any) => sum + (entry.value || 0), 0);
            const activeRulesCount = payload.filter((entry: any) => entry.value && entry.value > 0).length;
            
            return (
              <div style={{
                background: token.colorBgContainer,
                padding: '12px 14px',
                borderRadius: 10,
                border: `1px solid ${token.colorBorderSecondary}`,
                minWidth: 400,
                maxWidth: 600,
                boxShadow: '0 4px 12px rgba(0, 0, 0, 0.15)'
              }}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12, marginBottom: 12 }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <ClockCircleOutlined style={{ color: token.colorTextSecondary }} />
                    <span style={{ fontWeight: 600, color: token.colorText }}>
                      {formatTooltipLabel(label)}
                    </span>
                  </div>
                  <span style={{ fontWeight: 700, color: token.colorPrimary, fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace', fontSize: 16 }}>
                    {total}
                  </span>
                </div>
                {allRules.length > 0 ? (
                  <>
                    <div style={{ marginBottom: 8 }}>
                      {allRules.map((entry: any, index: number) => (
                        <div key={index} style={{ 
                          display: 'flex', 
                          alignItems: 'center', 
                          justifyContent: 'space-between',
                          gap: 12,
                          marginBottom: 8,
                          padding: '6px 0',
                          borderBottom: index < allRules.length - 1 ? `1px solid ${token.colorBorderSecondary}` : 'none'
                        }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1, minWidth: 0 }}>
                            <div style={{ 
                              width: 12, 
                              height: 12, 
                              borderRadius: 2, 
                              backgroundColor: entry.color,
                              flexShrink: 0
                            }} />
                            <span style={{ 
                              fontSize: 13,
                              color: entry.value && entry.value > 0 ? token.colorText : token.colorTextTertiary,
                              whiteSpace: 'normal',
                              wordBreak: 'break-word'
                            }}>{entry.name}</span>
                          </div>
                          <span style={{ 
                            fontWeight: 600, 
                            fontFamily: 'monospace',
                            fontSize: 13,
                            color: entry.value && entry.value > 0 ? token.colorText : token.colorTextTertiary,
                            flexShrink: 0,
                            minWidth: 50,
                            textAlign: 'right'
                          }}>{entry.value || 0}</span>
                        </div>
                      ))}
                    </div>
                    <div style={{ 
                      borderTop: `1px solid ${token.colorBorderSecondary}`, 
                      marginTop: 8, 
                      paddingTop: 8, 
                      display: 'flex', 
                      justifyContent: 'space-between', 
                      alignItems: 'center',
                      fontWeight: 600 
                    }}>
                      <span style={{ color: token.colorTextSecondary, fontSize: 12 }}>
                        有告警规则：{activeRulesCount} / {payload.length}
                      </span>
                      <span style={{ fontFamily: 'monospace', fontSize: 12, color: token.colorTextSecondary }}>
                        单位：次
                      </span>
                    </div>
                  </>
                ) : (
                  <div style={{ color: token.colorTextSecondary, fontSize: 13 }}>该时间点无告警</div>
                )}
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
  // 每5分钟（300000毫秒）自动刷新数据
  const REFETCH_INTERVAL = 5 * 60 * 1000; // 5 minutes
  const { token } = theme.useToken();

  const { data: statusData, isLoading } = useQuery({
    queryKey: ['status'],
    queryFn: () => statusApi.getStatus().then(res => res.data.data),
    refetchInterval: REFETCH_INTERVAL,
    refetchIntervalInBackground: true, // 即使页面在后台也继续刷新
  });

  const { data: ruleTimeSeriesData, isLoading: ruleTimeSeriesLoading, isFetching: isChartFetching } = useQuery({
    queryKey: ['rule-timeseries-stats'],
    queryFn: () => alertsApi.getRuleTimeSeries('24h', 60).then(res => res.data.data),
    refetchInterval: REFETCH_INTERVAL,
    refetchIntervalInBackground: true, // 即使页面在后台也继续刷新
  });

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  const getESStatus = () => {
    const es = statusData?.elasticsearch;
    if (!es || es.total === 0) return { value: '未配置', status: 'warning' };
    if (es.success_count === es.total) return { value: `${es.success_count}/${es.total} 正常`, status: 'success' };
    if (es.success_count > 0) return { value: `${es.success_count}/${es.total} 正常`, status: 'warning' };
    if (es.failed_count > 0) return { value: `${es.failed_count}/${es.total} 异常`, status: 'error' };
    return { value: `${es.total} 个配置`, status: 'default' };
  };

  const esStatus = getESStatus();

  return (
    <div>
      <PageHeader
        title="系统概览"
        description="实时监控系统运行状态与关键指标。"
        extra={null}
      />

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总规则数"
              value={statusData?.rules.total || 0}
              prefix={<DashboardOutlined style={{ color: token.colorPrimary }} />}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>已配置的规则总数</Text>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="启用规则"
              value={statusData?.rules.enabled || 0}
              prefix={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>当前启用的规则</Text>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="24小时告警"
              value={statusData?.alerts_24h?.total || 0}
              prefix={<AlertOutlined style={{ color: '#faad14' }} />}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>最近24小时告警数</Text>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="ES 数据源"
              value={esStatus.value}
              valueStyle={{
                color: esStatus.status === 'success' ? '#52c41a' :
                       esStatus.status === 'error' ? '#ff4d4f' :
                       esStatus.status === 'warning' ? '#faad14' : undefined
              }}
              prefix={<DatabaseOutlined style={{
                color: esStatus.status === 'success' ? '#52c41a' :
                       esStatus.status === 'error' ? '#ff4d4f' :
                       esStatus.status === 'warning' ? '#faad14' : token.colorPrimary
              }} />}
            />
            <Text type="secondary" style={{ fontSize: 12 }}>
              {statusData?.elasticsearch?.total
                ? `正常: ${statusData.elasticsearch.success_count || 0}, 异常: ${statusData.elasticsearch.failed_count || 0}`
                : '未配置数据源'
              }
            </Text>
          </Card>
        </Col>
      </Row>

      <Card
        title={
          <span>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
              <span
                style={{
                  width: 28,
                  height: 28,
                  borderRadius: 8,
                  display: 'inline-flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  background: token.colorBgLayout,
                  border: `1px solid ${token.colorBorderSecondary}`,
                }}
              >
                <LineChartOutlined style={{ color: token.colorPrimary }} />
              </span>
              <span>规则告警趋势</span>
              <Text type="secondary" style={{ fontSize: 12 }}>(24小时)</Text>
            </span>
            {isChartFetching && <SyncOutlined spin style={{ marginLeft: 8 }} />}
          </span>
        }
        extra={
          ruleTimeSeriesData && ruleTimeSeriesData.length > 0 && (
            <Text type="secondary">{ruleTimeSeriesData.length} 条规则</Text>
          )
        }
      >
        {ruleTimeSeriesLoading ? (
          <div style={{ textAlign: 'center', padding: 48 }}>
            <Spin tip="加载中..." />
          </div>
        ) : !ruleTimeSeriesData || ruleTimeSeriesData.length === 0 ? (
          <div className="app-muted" style={{ textAlign: 'center', padding: 48 }}>
            暂无启用的规则
          </div>
        ) : (
          <InteractiveAreaChart data={ruleTimeSeriesData} />
        )}
      </Card>
    </div>
  );
}
