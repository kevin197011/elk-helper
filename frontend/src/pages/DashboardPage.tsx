// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { Card, Row, Col, Statistic, Spin, Typography } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { statusApi, alertsApi } from '../services/api';
import {
  DashboardOutlined,
  CheckCircleOutlined,
  AlertOutlined,
  DatabaseOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import { Area, AreaChart, CartesianGrid, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import { useMemo, useState, useEffect } from 'react';

const { Title, Text } = Typography;

// 自动刷新指示器组件
function AutoRefreshIndicator({ isFetching }: { isFetching: boolean }) {
  const [timeUntilRefresh, setTimeUntilRefresh] = useState(300); // 5分钟 = 300秒

  useEffect(() => {
    if (isFetching) {
      setTimeUntilRefresh(300); // 刷新时重置倒计时
    }

    const interval = setInterval(() => {
      setTimeUntilRefresh(prev => {
        if (prev <= 1) {
          return 300; // 重置为5分钟
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [isFetching]);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div style={{ 
      display: 'flex', 
      alignItems: 'center', 
      gap: 8,
      padding: '4px 12px',
      background: 'rgba(22, 119, 255, 0.08)',
      borderRadius: 6,
      fontSize: 12
    }}>
      <SyncOutlined style={{ color: '#1677ff', fontSize: 12 }} />
      <Text type="secondary" style={{ fontSize: 12 }}>
        自动刷新: {formatTime(timeUntilRefresh)}
      </Text>
    </div>
  );
}

// InteractiveAreaChart component
function InteractiveAreaChart({ data }: { data: any[] }) {
  const chartData = useMemo(() => {
    if (!data || data.length === 0) return [];

    const timePoints = data[0]?.data_points || [];

    return timePoints.map((point: any, index: number) => {
      const dataPoint: any = { time: point.time };
      data.forEach((rule: any) => {
        const value = rule.data_points[index]?.value || 0;
        dataPoint[rule.rule_name] = value;
      });
      return dataPoint;
    });
  }, [data]);

  const colors = [
    '#1677ff', '#722ed1', '#52c41a', '#fa8c16',
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
        <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" vertical={false} />
        <XAxis dataKey="time" tickLine={false} axisLine={false} tickMargin={8} tick={{ fill: '#8c8c8c' }} />
        <YAxis tickLine={false} axisLine={false} tickMargin={8} tick={{ fill: '#8c8c8c' }} allowDecimals={false} />
        <Tooltip
          content={({ active, payload, label }) => {
            if (!active || !payload) return null;
            // 过滤掉值为 0 的规则，只显示有告警的规则
            const activeRules = payload.filter((entry: any) => entry.value && entry.value > 0);
            const total = payload.reduce((sum: number, entry: any) => sum + (entry.value || 0), 0);
            
            return (
              <div style={{
                background: '#fff',
                padding: '12px 16px',
                borderRadius: 8,
                boxShadow: '0 2px 8px rgba(0,0,0,0.15)',
                border: '1px solid #f0f0f0',
                maxWidth: 300
              }}>
                <div style={{ fontWeight: 500, marginBottom: 12, color: '#262626' }}>时间: {label}</div>
                {activeRules.length > 0 ? (
                  <>
                    <div style={{ marginBottom: 8, maxHeight: 200, overflowY: 'auto' }}>
                      {activeRules.map((entry: any, index: number) => (
                        <div key={index} style={{ 
                          display: 'flex', 
                          alignItems: 'center', 
                          justifyContent: 'space-between',
                          gap: 12,
                          marginBottom: 6,
                          padding: '4px 0'
                        }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 6, flex: 1, minWidth: 0 }}>
                            <div style={{ 
                              width: 10, 
                              height: 10, 
                              borderRadius: 2, 
                              backgroundColor: entry.color,
                              flexShrink: 0
                            }} />
                            <span style={{ 
                              fontSize: 13,
                              color: '#595959',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap'
                            }}>{entry.name}</span>
                          </div>
                          <span style={{ 
                            fontWeight: 600, 
                            fontFamily: 'monospace',
                            fontSize: 13,
                            color: '#262626',
                            flexShrink: 0
                          }}>{entry.value}</span>
                        </div>
                      ))}
                    </div>
                    <div style={{ 
                      borderTop: '1px solid #f0f0f0', 
                      marginTop: 8, 
                      paddingTop: 8, 
                      display: 'flex', 
                      justifyContent: 'space-between', 
                      alignItems: 'center',
                      fontWeight: 600 
                    }}>
                      <span style={{ color: '#262626' }}>总计:</span>
                      <span style={{ 
                        fontFamily: 'monospace', 
                        fontSize: 14,
                        color: '#1677ff'
                      }}>{total}</span>
                    </div>
                  </>
                ) : (
                  <div style={{ color: '#8c8c8c', fontSize: 13 }}>该时间点无告警</div>
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

  const { data: statusData, isLoading, isFetching: isStatusFetching } = useQuery({
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
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <Title level={4} style={{ marginBottom: 4 }}>
              系统概览
              {(isStatusFetching || isChartFetching) && (
                <SyncOutlined spin style={{ marginLeft: 12, fontSize: 14, color: '#1677ff' }} />
              )}
            </Title>
            <Text type="secondary">实时监控系统运行状态和关键指标</Text>
          </div>
          <AutoRefreshIndicator isFetching={isStatusFetching || isChartFetching} />
        </div>
      </div>

      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="总规则数"
              value={statusData?.rules.total || 0}
              prefix={<DashboardOutlined style={{ color: '#1677ff' }} />}
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
                       esStatus.status === 'warning' ? '#faad14' : '#1677ff'
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
            规则告警趋势 (24小时)
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
          <div style={{ textAlign: 'center', padding: 48, color: '#8c8c8c' }}>
            暂无启用的规则
          </div>
        ) : (
          <InteractiveAreaChart data={ruleTimeSeriesData} />
        )}
      </Card>
    </div>
  );
}
