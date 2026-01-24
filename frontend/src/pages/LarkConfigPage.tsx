// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState } from 'react';
import { Table, Tag, Button, Space, Typography, Tooltip, App } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  StarOutlined,
  StarFilled,
  LoadingOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { larkConfigApi, LarkConfig } from '../services/api';
import LarkConfigEditDialog from '../components/LarkConfigEditDialog';
import PageHeader from '../components/PageHeader';

const { Text } = Typography;

export default function LarkConfigPage() {
  const queryClient = useQueryClient();
  const { message, modal } = App.useApp();
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [selectedConfig, setSelectedConfig] = useState<LarkConfig | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['lark-configs'],
    queryFn: () => larkConfigApi.getAll().then(res => res.data.data),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => larkConfigApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
      message.success('删除成功');
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '删除失败');
    },
  });

  const testMutation = useMutation({
    mutationFn: (id: number) => larkConfigApi.test(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
      if (data.data.success) {
        message.success('连接测试成功');
      } else {
        message.error(data.data.error || '连接测试失败');
      }
    },
    onError: () => {
      message.error('测试失败');
    },
  });

  const setDefaultMutation = useMutation({
    mutationFn: (id: number) => larkConfigApi.setDefault(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
      message.success('已设置为默认配置');
    },
    onError: () => {
      message.error('设置失败');
    },
  });

  const handleDelete = (config: LarkConfig) => {
    modal.confirm({
      title: '确认删除',
      icon: <ExclamationCircleOutlined />,
      content: `确定要删除告警配置 "${config.name}" 吗？如果有规则使用此配置，将无法删除。`,
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => deleteMutation.mutateAsync(config.id),
    });
  };

  const columns: ColumnsType<LarkConfig> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 70,
    },
    {
      title: '配置名称',
      dataIndex: 'name',
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: 'Webhook URL',
      dataIndex: 'webhook_url',
      ellipsis: true,
      render: (url: string) => (
        <Tooltip title={url}>
          <Text code style={{ maxWidth: 300, display: 'inline-block' }}>{url}</Text>
        </Tooltip>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 90,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>
          {enabled ? '启用' : '禁用'}
        </Tag>
      ),
    },
    {
      title: '测试状态',
      dataIndex: 'test_status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'success' ? 'success' : status === 'failed' ? 'error' : 'default'}>
          {status === 'success' ? '成功' : status === 'failed' ? '失败' : '未测试'}
        </Tag>
      ),
    },
    {
      title: '默认',
      dataIndex: 'is_default',
      width: 70,
      render: (isDefault: boolean) =>
        isDefault ? (
          <StarFilled style={{ color: '#faad14' }} />
        ) : (
          <StarOutlined style={{ color: '#d9d9d9' }} />
        ),
    },
    {
      title: '操作',
      width: 180,
      render: (_, record) => (
        <Space>
          <Tooltip title="测试连接">
            <Button
              type="text"
              icon={testMutation.isPending ? <LoadingOutlined /> : <PlayCircleOutlined />}
              onClick={() => testMutation.mutate(record.id)}
              loading={testMutation.isPending}
            />
          </Tooltip>
          {!record.is_default && (
            <Tooltip title="设为默认">
              <Button
                type="text"
                icon={<StarOutlined />}
                onClick={() => setDefaultMutation.mutate(record.id)}
              />
            </Tooltip>
          )}
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedConfig(record);
                setEditModalOpen(true);
              }}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDelete(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title="告警配置"
        description="管理告警通知渠道（Webhook）与默认配置。"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setSelectedConfig(null);
              setEditModalOpen(true);
            }}
          >
            新建告警配置
          </Button>
        }
      />

      <Table
        columns={columns}
        dataSource={data || []}
        rowKey="id"
        loading={isLoading}
        pagination={false}
        locale={{
          emptyText: '暂无配置，点击"新建告警配置"创建第一个告警配置',
        }}
      />

      <LarkConfigEditDialog
        open={editModalOpen}
        onOpenChange={setEditModalOpen}
        config={selectedConfig}
        onSuccess={() => {
          setEditModalOpen(false);
          queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
        }}
      />
    </div>
  );
}
