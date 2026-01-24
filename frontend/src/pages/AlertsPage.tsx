// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useMemo } from 'react';
import { Table, Tag, Button, Input, Select, Modal, Space, Typography, App } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { EyeOutlined, SearchOutlined, SyncOutlined, CopyOutlined } from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { alertsApi, Alert } from '../services/api';
import PageHeader from '../components/PageHeader';

const { Text, Paragraph } = Typography;

export default function AlertsPage() {
  const { message } = App.useApp();
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null);
  const [detailModalOpen, setDetailModalOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'sent' | 'failed'>('all');

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['alerts', page],
    queryFn: () => alertsApi.getAll(page, pageSize).then(res => res.data),
    staleTime: 30000,
    gcTime: 5 * 60 * 1000,
    refetchOnWindowFocus: false,
  });

  const filteredAlerts = useMemo(() => {
    if (!data?.data) return [];
    let filtered = data.data;

    if (statusFilter !== 'all') {
      filtered = filtered.filter(alert => alert.status === statusFilter);
    }

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(alert =>
        alert.rule?.name?.toLowerCase().includes(query) ||
        alert.index_name.toLowerCase().includes(query) ||
        alert.time_range.toLowerCase().includes(query)
      );
    }

    return filtered;
  }, [data?.data, searchQuery, statusFilter]);

  const handleView = async (alert: Alert) => {
    try {
      const response = await alertsApi.getById(alert.id);
      setSelectedAlert(response.data.data);
      setDetailModalOpen(true);
    } catch (error: any) {
      message.error(error?.response?.data?.error || '获取告警详情失败');
      setSelectedAlert(alert);
      setDetailModalOpen(true);
    }
  };

  const handleCopyLogs = () => {
    if (selectedAlert?.logs) {
      navigator.clipboard.writeText(JSON.stringify(selectedAlert.logs, null, 2));
      message.success('已复制到剪贴板');
    }
  };

  const columns: ColumnsType<Alert> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 70,
    },
    {
      title: '规则名称',
      dataIndex: ['rule', 'name'],
      render: (name: string) => name || '-',
    },
    {
      title: '索引名称',
      dataIndex: 'index_name',
      render: (name: string) => <Text code>{name}</Text>,
    },
    {
      title: '日志数量',
      dataIndex: 'log_count',
      width: 90,
    },
    {
      title: '时间范围',
      dataIndex: 'time_range',
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (status: string) => (
        <Tag color={status === 'sent' ? 'success' : 'error'}>
          {status === 'sent' ? '已发送' : '失败'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      width: 80,
      render: (_, record) => (
        <Button
          type="text"
          icon={<EyeOutlined />}
          onClick={() => handleView(record)}
        />
      ),
    },
  ];

  const displayAlerts = filteredAlerts.length > 0 ? filteredAlerts : (data?.data || []);

  return (
    <div>
      <PageHeader
        title="告警历史"
        description="查看告警记录与发送状态。"
        extra={
          isFetching && !isLoading ? (
            <Space>
              <SyncOutlined spin />
              <Text type="secondary">刷新中...</Text>
            </Space>
          ) : null
        }
      />

      <Space className="app-page-section" style={{ marginTop: 0, marginBottom: 16 }}>
        <Input
          placeholder="搜索规则名称、索引或时间范围..."
          prefix={<SearchOutlined />}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          allowClear
          style={{ width: 300 }}
        />
        <Select
          value={statusFilter}
          onChange={setStatusFilter}
          style={{ width: 120 }}
          options={[
            { value: 'all', label: '全部状态' },
            { value: 'sent', label: '已发送' },
            { value: 'failed', label: '失败' },
          ]}
        />
      </Space>

      <Table
        columns={columns}
        dataSource={displayAlerts}
        rowKey="id"
        loading={isLoading}
        pagination={
          !searchQuery && statusFilter === 'all'
            ? {
                current: page,
                pageSize: pageSize,
                total: data?.pagination.total || 0,
                showSizeChanger: false,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 条`,
                onChange: (p) => setPage(p),
              }
            : false
        }
        locale={{
          emptyText: searchQuery || statusFilter !== 'all'
            ? '没有找到匹配的告警'
            : '暂无告警记录',
        }}
      />

      {/* Alert Detail Modal */}
      <Modal
        title={`告警详情 #${selectedAlert?.id}`}
        open={detailModalOpen}
        onCancel={() => setDetailModalOpen(false)}
        footer={<Button onClick={() => setDetailModalOpen(false)}>关闭</Button>}
        width={800}
      >
        {selectedAlert && (
          <div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 24 }}>
              <div>
                <Text type="secondary">规则名称</Text>
                <Paragraph style={{ marginBottom: 0 }}>{selectedAlert.rule?.name || '-'}</Paragraph>
              </div>
              <div>
                <Text type="secondary">索引名称</Text>
                <Paragraph code style={{ marginBottom: 0 }}>{selectedAlert.index_name}</Paragraph>
              </div>
              <div>
                <Text type="secondary">日志数量</Text>
                <Paragraph style={{ marginBottom: 0 }}>{selectedAlert.log_count}</Paragraph>
              </div>
              <div>
                <Text type="secondary">时间范围</Text>
                <Paragraph style={{ marginBottom: 0 }}>{selectedAlert.time_range}</Paragraph>
              </div>
              <div>
                <Text type="secondary">状态</Text>
                <div>
                  <Tag color={selectedAlert.status === 'sent' ? 'success' : 'error'}>
                    {selectedAlert.status === 'sent' ? '已发送' : '失败'}
                  </Tag>
                </div>
              </div>
              <div>
                <Text type="secondary">创建时间</Text>
                <Paragraph style={{ marginBottom: 0 }}>
                  {new Date(selectedAlert.created_at).toLocaleString()}
                </Paragraph>
              </div>
              {selectedAlert.error_msg && (
                <div style={{ gridColumn: '1 / -1' }}>
                  <Text type="secondary">错误信息</Text>
                  <Paragraph type="danger" style={{ marginBottom: 0 }}>
                    {selectedAlert.error_msg}
                  </Paragraph>
                </div>
              )}
            </div>

            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                <Text type="secondary">
                  日志详情
                  {selectedAlert.log_count > 0 && ` (共 ${selectedAlert.log_count} 条，显示前 10 条)`}
                </Text>
                {selectedAlert.logs?.length > 0 && (
                  <Button size="small" icon={<CopyOutlined />} onClick={handleCopyLogs}>
                    复制 JSON
                  </Button>
                )}
              </div>
              {selectedAlert.logs?.length > 0 ? (
                <>
                  <pre className="app-code-block" style={{ maxHeight: 400 }}>
                    {JSON.stringify(selectedAlert.logs, null, 2)}
                  </pre>
                  {selectedAlert.log_count > 10 && (
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      为了性能考虑，详情仅显示前 10 条日志。完整日志已发送到 Lark/飞书通知。
                    </Text>
                  )}
                </>
              ) : (
                <div className="app-surface-muted app-muted" style={{ padding: 16 }}>
                  {selectedAlert.log_count > 0
                    ? '日志数据为空，可能已被清理或数据格式异常'
                    : '暂无日志数据'}
                </div>
              )}
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
