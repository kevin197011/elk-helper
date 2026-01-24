// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { systemConfigApi, CleanupConfig } from '../services/api';
import { Card, Form, InputNumber, Switch, Button, Space, Typography, Spin, App, theme } from 'antd';
import {
  SaveOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import PageHeader from '../components/PageHeader';

const { Text, Paragraph } = Typography;

export default function CleanupConfigPage() {
  const { message, modal } = App.useApp();
  const queryClient = useQueryClient();
  const [form] = Form.useForm();
  const [isSaving, setIsSaving] = useState(false);
  const [isCleaning, setIsCleaning] = useState(false);
  const { token } = theme.useToken();

  const { data, isLoading } = useQuery({
    queryKey: ['cleanup-config'],
    queryFn: () => systemConfigApi.getCleanupConfig().then(res => res.data.data),
  });

  useEffect(() => {
    if (data) {
      form.setFieldsValue({
        enabled: data.enabled,
        hour: data.hour,
        minute: data.minute,
        retention_days: data.retention_days,
      });
    }
  }, [data, form]);

  const updateMutation = useMutation({
    mutationFn: (config: CleanupConfig) => systemConfigApi.updateCleanupConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cleanup-config'] });
      message.success('清理任务配置已更新');
      setIsSaving(false);
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || '更新配置时发生错误');
      setIsSaving(false);
    },
  });

  const manualCleanupMutation = useMutation({
    mutationFn: () => systemConfigApi.manualCleanup(),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      queryClient.invalidateQueries({ queryKey: ['cleanup-config'] });
      message.success(`已删除 ${response.data.deleted_count} 条超过 ${response.data.retention_days} 天的历史告警数据`);
      setIsCleaning(false);
    },
    onError: (error: any) => {
      message.error(error.response?.data?.error || '执行清理时发生错误');
      setIsCleaning(false);
    },
  });

  const handleSubmit = (values: any) => {
    setIsSaving(true);
    updateMutation.mutate(values);
  };

  const handleManualCleanup = () => {
    const retentionDays = form.getFieldValue('retention_days');
    modal.confirm({
      title: '确认立即清理',
      icon: <ExclamationCircleOutlined />,
      content: (
        <div>
          <p>此操作将立即删除超过 <strong>{retentionDays} 天</strong> 的历史告警数据。</p>
        </div>
      ),
      okText: '确认清理',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => {
        setIsCleaning(true);
        manualCleanupMutation.mutate();
      },
    });
  };

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="清理任务配置"
        description="配置定时清理历史告警数据。"
        extra={
          <Button danger icon={<DeleteOutlined />} onClick={handleManualCleanup} loading={isCleaning}>
            立即清理
          </Button>
        }
      />

      <Card title="定时清理设置" style={{ maxWidth: 600 }}>
        <Paragraph type="secondary" style={{ marginBottom: 24 }}>
          系统将自动删除超过保留期限的历史告警数据，以节省存储空间
        </Paragraph>

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            enabled: true,
            hour: 3,
            minute: 0,
            retention_days: 90,
          }}
        >
          <Form.Item
            name="enabled"
            label="启用清理任务"
            valuePropName="checked"
            extra="开启后，系统将按设定的时间自动清理历史数据"
          >
            <Switch />
          </Form.Item>

          <Space size="large">
            <Form.Item
              name="hour"
              label="执行时间 - 小时"
              rules={[
                { required: true, message: '请输入小时' },
                { type: 'number', min: 0, max: 23, message: '小时必须在 0-23 之间' },
              ]}
              extra="0-23"
            >
              <InputNumber min={0} max={23} style={{ width: 120 }} />
            </Form.Item>

            <Form.Item
              name="minute"
              label="执行时间 - 分钟"
              rules={[
                { required: true, message: '请输入分钟' },
                { type: 'number', min: 0, max: 59, message: '分钟必须在 0-59 之间' },
              ]}
              extra="0-59"
            >
              <InputNumber min={0} max={59} style={{ width: 120 }} />
            </Form.Item>
          </Space>

          <Form.Item
            name="retention_days"
            label="数据保留天数"
            rules={[
              { required: true, message: '请输入保留天数' },
              { type: 'number', min: 1, message: '保留天数必须至少为 1 天' },
            ]}
            extra="超过此天数的告警数据将被自动删除。例如：90 天表示删除 3 个月前的数据"
          >
            <InputNumber min={1} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" icon={<SaveOutlined />} loading={isSaving}>
              保存配置
            </Button>
          </Form.Item>
        </Form>

        <Form.Item noStyle shouldUpdate>
          {() => {
            const enabled = form.getFieldValue('enabled');
            const hour = form.getFieldValue('hour');
            const minute = form.getFieldValue('minute');
            const retentionDays = form.getFieldValue('retention_days');

            return enabled ? (
              <div className="app-surface-muted" style={{ padding: 16, marginBottom: 24 }}>
                <Text strong>当前配置预览</Text>
                <Paragraph style={{ marginBottom: 0, marginTop: 8 }}>
                  系统将在每天 <Text strong>{String(hour || 0).padStart(2, '0')}:{String(minute || 0).padStart(2, '0')}</Text> 执行清理任务，
                  删除 <Text strong>{retentionDays || 90} 天</Text>前的告警数据
                </Paragraph>
              </div>
            ) : null;
          }}
        </Form.Item>

        {/* Last Execution Status */}
        {data && (
          <div style={{ border: `1px solid ${token.colorBorderSecondary}`, padding: 16, borderRadius: 10 }}>
            <Text strong style={{ marginBottom: 12, display: 'block' }}>上次执行状态</Text>
            {!data.last_execution_status || data.last_execution_status === 'never' ? (
              <Space>
                <ClockCircleOutlined style={{ color: token.colorTextSecondary }} />
                <Text type="secondary">尚未执行</Text>
              </Space>
            ) : data.last_execution_status === 'success' ? (
              <div>
                <Space>
                  <CheckCircleOutlined style={{ color: '#52c41a' }} />
                  <Text style={{ color: '#52c41a' }}>执行成功</Text>
                </Space>
                {data.last_execution_time && (
                  <Paragraph type="secondary" style={{ marginLeft: 22, marginBottom: 4 }}>
                    执行时间: {new Date(data.last_execution_time).toLocaleString('zh-CN')}
                  </Paragraph>
                )}
                {data.last_execution_result && (
                  <Paragraph type="secondary" style={{ marginLeft: 22, marginBottom: 0 }}>
                    {data.last_execution_result}
                  </Paragraph>
                )}
              </div>
            ) : data.last_execution_status === 'failed' ? (
              <div>
                <Space>
                  <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
                  <Text type="danger">执行失败</Text>
                </Space>
                {data.last_execution_time && (
                  <Paragraph type="secondary" style={{ marginLeft: 22, marginBottom: 4 }}>
                    执行时间: {new Date(data.last_execution_time).toLocaleString('zh-CN')}
                  </Paragraph>
                )}
                {data.last_execution_result && (
                  <Paragraph type="danger" style={{ marginLeft: 22, marginBottom: 0 }}>
                    {data.last_execution_result}
                  </Paragraph>
                )}
              </div>
            ) : null}
          </div>
        )}
      </Card>
    </div>
  );
}
