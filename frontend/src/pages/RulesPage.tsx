// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useMemo } from 'react';
import { Table, Button, Input, Select, Tag, Switch, Space, Modal, Upload, Typography, App, Tooltip } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SearchOutlined,
  DownloadOutlined,
  UploadOutlined,
  CopyOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { rulesApi, Rule } from '../services/api';
import { theme } from 'antd';
import PageHeader from '../components/PageHeader';
import RuleEditDialog from '../components/RuleEditDialog';

const { Text } = Typography;

export default function RulesPage() {
  const queryClient = useQueryClient();
  const { message, modal } = App.useApp();
  const { token } = theme.useToken();
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all');
  const [importModalOpen, setImportModalOpen] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [cloneModalOpen, setCloneModalOpen] = useState(false);
  const [ruleToClone, setRuleToClone] = useState<{ id: number; name: string } | null>(null);
  const [clonedRuleName, setClonedRuleName] = useState('');
  const [editOpen, setEditOpen] = useState(false);
  const [editingRuleId, setEditingRuleId] = useState<number | undefined>(undefined);

  const { data, isLoading } = useQuery({
    queryKey: ['rules'],
    queryFn: () => rulesApi.getAll().then(res => res.data.data),
    refetchInterval: false,
    staleTime: 30000,
  });

  const filteredRules = useMemo(() => {
    if (!data) return [];
    let filtered = [...data];

    if (statusFilter !== 'all') {
      filtered = filtered.filter(rule =>
        statusFilter === 'enabled' ? rule.enabled : !rule.enabled
      );
    }

    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(rule =>
        rule.name.toLowerCase().includes(query) ||
        rule.index_pattern.toLowerCase().includes(query) ||
        (rule.description && rule.description.toLowerCase().includes(query))
      );
    }

    filtered.sort((a, b) => b.id - a.id);
    return filtered;
  }, [data, searchQuery, statusFilter]);

  const deleteMutation = useMutation({
    mutationFn: (id: number) => rulesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      message.success('删除成功');
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '删除失败');
    },
  });

  const toggleMutation = useMutation({
    mutationFn: (id: number) => rulesApi.toggleEnabled(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      message.success(data.data.data.enabled ? '规则已启用' : '规则已禁用');
    },
    onError: () => {
      message.error('操作失败');
    },
  });

  const exportMutation = useMutation({
    mutationFn: async () => {
      const response = await rulesApi.export();
      return response.data;
    },
    onSuccess: (data) => {
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `rules_export_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.json`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      message.success('规则导出成功');
    },
    onError: () => {
      message.error('规则导出失败');
    },
  });

  const importMutation = useMutation({
    mutationFn: (rules: any[]) => rulesApi.import(rules),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      const { created_count, updated_count, skipped_count, error_count, errors } = response.data;

      const parts: string[] = [];
      if (created_count > 0) parts.push(`新建 ${created_count} 条`);
      if (updated_count > 0) parts.push(`更新 ${updated_count} 条`);
      if (skipped_count > 0) parts.push(`跳过 ${skipped_count} 条`);

      if (error_count === 0) {
        message.success(`导入完成：${parts.join('，') || '无变更'}`);
      } else {
        message.error(`导入完成：${parts.join('，')}，失败 ${error_count} 条。${errors.slice(0, 3).join('; ')}`);
      }
      setImportModalOpen(false);
      setImportFile(null);
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '规则导入失败');
    },
  });

  const cloneMutation = useMutation({
    mutationFn: ({ id, name }: { id: number; name: string }) => rulesApi.clone(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      message.success('规则克隆成功');
      setCloneModalOpen(false);
      setRuleToClone(null);
      setClonedRuleName('');
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '克隆失败');
    },
  });

  const handleDelete = (id: number, name: string) => {
    modal.confirm({
      title: '确认删除',
      icon: <ExclamationCircleOutlined />,
      content: `确定要删除规则 "${name}" 吗？`,
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => deleteMutation.mutateAsync(id),
    });
  };

  const handleImport = () => {
    if (!importFile) {
      message.error('请选择要导入的文件');
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const text = e.target?.result as string;
        const data = JSON.parse(text);

        let rules: any[] = [];
        if (Array.isArray(data)) {
          rules = data;
        } else if (data.rules && Array.isArray(data.rules)) {
          rules = data.rules;
        } else {
          message.error('文件格式错误：无法识别规则数据');
          return;
        }

        if (rules.length === 0) {
          message.error('文件中没有规则数据');
          return;
        }

        importMutation.mutate(rules);
      } catch (error) {
        message.error('文件解析失败：' + (error instanceof Error ? error.message : '未知错误'));
      }
    };
    reader.readAsText(importFile);
  };

  const handleClone = (id: number, name: string) => {
    setRuleToClone({ id, name });
    setClonedRuleName(`${name} - 副本`);
    setCloneModalOpen(true);
  };

  const openCreate = () => {
    setEditingRuleId(undefined);
    setEditOpen(true);
  };

  const openEdit = (id: number) => {
    setEditingRuleId(id);
    setEditOpen(true);
  };

  const columns: ColumnsType<Rule> = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 70,
    },
    {
      title: '规则名称',
      dataIndex: 'name',
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: '索引模式',
      dataIndex: 'index_pattern',
      render: (pattern: string) => <Text code>{pattern}</Text>,
    },
    {
      title: '数据源',
      dataIndex: 'es_config',
      render: (config: any) => config ? (
        <Space>
          <Tag>{config.name}</Tag>
          {config.is_default && <Text type="secondary" style={{ fontSize: 12 }}>(默认)</Text>}
        </Space>
      ) : <Text type="secondary">-</Text>,
    },
    {
      title: 'Lark配置',
      render: (_, record) => record.lark_config ? (
        <Space>
          <Tag>{record.lark_config.name}</Tag>
          {record.lark_config.is_default && <Text type="secondary" style={{ fontSize: 12 }}>(默认)</Text>}
        </Space>
      ) : record.lark_webhook ? (
        <Tag>自定义Webhook</Tag>
      ) : <Text type="secondary">-</Text>,
    },
    {
      title: '查询间隔',
      dataIndex: 'interval',
      width: 90,
      render: (interval: number) => `${interval}秒`,
    },
    {
      title: '执行次数',
      dataIndex: 'run_count',
      width: 90,
    },
    {
      title: '告警次数',
      dataIndex: 'alert_count',
      width: 90,
    },
    {
      title: '上次执行',
      dataIndex: 'last_run_time',
      width: 130,
      render: (time: string) => time ? (
        <div>
          <div>{new Date(time).toLocaleDateString()}</div>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {new Date(time).toLocaleTimeString()}
          </Text>
        </div>
      ) : <Text type="secondary">未执行</Text>,
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 120,
      render: (enabled: boolean, record) => (
        <Space>
          <Switch
            checked={enabled}
            onChange={() => toggleMutation.mutate(record.id)}
            loading={toggleMutation.isPending}
            size="small"
          />
          <Tag color={enabled ? 'success' : 'default'}>
            {enabled ? '启用' : '禁用'}
          </Tag>
        </Space>
      ),
    },
    {
      title: '操作',
      width: 130,
      render: (_, record) => (
        <Space>
          <Tooltip title="编辑">
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => openEdit(record.id)}
            />
          </Tooltip>
          <Tooltip title="克隆">
            <Button
              type="text"
              icon={<CopyOutlined style={{ color: token.colorPrimary }} />}
              onClick={() => handleClone(record.id, record.name)}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDelete(record.id, record.name)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        title="规则管理"
        description="创建与维护告警规则，支持测试、导入导出与启停。"
        extra={
          <Space>
            <Button icon={<DownloadOutlined />} onClick={() => exportMutation.mutate()} loading={exportMutation.isPending}>
              导出
            </Button>
            <Button icon={<UploadOutlined />} onClick={() => setImportModalOpen(true)}>
              导入
            </Button>
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              新建规则
            </Button>
          </Space>
        }
      />

      <Space className="app-page-section" style={{ marginTop: 0, marginBottom: 16 }}>
        <Input
          placeholder="搜索规则名称、索引模式或描述..."
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
            { value: 'enabled', label: '已启用' },
            { value: 'disabled', label: '已禁用' },
          ]}
        />
      </Space>

      <Table
        columns={columns}
        dataSource={filteredRules}
        rowKey="id"
        loading={isLoading}
        pagination={{
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total) => `共 ${total} 条`,
        }}
        locale={{
          emptyText: searchQuery || statusFilter !== 'all'
            ? '没有找到匹配的规则'
            : '暂无规则，点击"新建规则"创建第一个规则',
        }}
      />

      {/* 导入弹窗 */}
      <Modal
        title="导入规则"
        open={importModalOpen}
        onCancel={() => {
          setImportModalOpen(false);
          setImportFile(null);
        }}
        onOk={handleImport}
        confirmLoading={importMutation.isPending}
        okText="确认导入"
        cancelText="取消"
      >
        <p style={{ marginBottom: 16 }} className="app-muted">
          请选择要导入的规则 JSON 文件。如果规则名称已存在，将会更新现有规则。
        </p>
        <Upload.Dragger
          accept=".json"
          beforeUpload={(file) => {
            setImportFile(file);
            return false;
          }}
          fileList={importFile ? [importFile as any] : []}
          onRemove={() => setImportFile(null)}
          maxCount={1}
        >
          <p className="ant-upload-drag-icon">
            <UploadOutlined style={{ fontSize: 32, color: token.colorPrimary }} />
          </p>
          <p className="ant-upload-text">点击或拖拽文件到此区域上传</p>
          <p className="ant-upload-hint">支持 JSON 格式文件</p>
        </Upload.Dragger>
      </Modal>

      {/* 克隆弹窗 */}
      <Modal
        title="克隆规则"
        open={cloneModalOpen}
        onCancel={() => {
          setCloneModalOpen(false);
          setRuleToClone(null);
          setClonedRuleName('');
        }}
        onOk={() => {
          if (!clonedRuleName.trim()) {
            message.error('请输入新规则名称');
            return;
          }
          cloneMutation.mutate({ id: ruleToClone!.id, name: clonedRuleName.trim() });
        }}
        confirmLoading={cloneMutation.isPending}
        okText="确认克隆"
        cancelText="取消"
      >
        <p style={{ marginBottom: 16 }}>
          克隆规则 <strong>"{ruleToClone?.name}"</strong> 并创建一个新规则。
        </p>
        <Input
          placeholder="请输入新规则名称"
          value={clonedRuleName}
          onChange={(e) => setClonedRuleName(e.target.value)}
          onPressEnter={() => {
            if (clonedRuleName.trim() && !cloneMutation.isPending) {
              cloneMutation.mutate({ id: ruleToClone!.id, name: clonedRuleName.trim() });
            }
          }}
        />
        <Text type="secondary" style={{ fontSize: 12, display: 'block', marginTop: 8 }}>
          新规则将复制原规则的所有配置，但统计数据会重置。
        </Text>
      </Modal>

      <RuleEditDialog
        open={editOpen}
        ruleId={editingRuleId}
        onOpenChange={(next) => {
          setEditOpen(next);
          if (!next) setEditingRuleId(undefined);
        }}
      />
    </div>
  );
}
