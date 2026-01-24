// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect } from 'react';
import { Modal, Form, Input, Switch, App } from 'antd';
import { useMutation } from '@tanstack/react-query';
import { larkConfigApi, LarkConfig } from '../services/api';

const { TextArea } = Input;

interface LarkConfigEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config?: LarkConfig | null;
  onSuccess?: () => void;
}

export default function LarkConfigEditDialog({
  open,
  onOpenChange,
  config,
  onSuccess,
}: LarkConfigEditDialogProps) {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const isEdit = !!config;

  useEffect(() => {
    if (open) {
      if (config) {
        form.setFieldsValue(config);
      } else {
        form.resetFields();
      }
    }
  }, [config, form, open]);

  const createMutation = useMutation({
    mutationFn: (data: Partial<LarkConfig>) => larkConfigApi.create(data),
    onSuccess: () => {
      message.success('创建成功');
      onSuccess?.();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '创建失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<LarkConfig> }) =>
      larkConfigApi.update(id, data),
    onSuccess: () => {
      message.success('更新成功');
      onSuccess?.();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '更新失败');
    },
  });

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (isEdit && config) {
        updateMutation.mutate({ id: config.id, data: values });
      } else {
        createMutation.mutate(values);
      }
    } catch (error) {
      // Form validation failed
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑告警配置' : '新建告警配置'}
      open={open}
      onCancel={() => onOpenChange(false)}
      onOk={handleSubmit}
      confirmLoading={createMutation.isPending || updateMutation.isPending}
      okText="保存"
      cancelText="取消"
      width={600}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          enabled: true,
          is_default: false,
        }}
        style={{ marginTop: 16 }}
      >
        <Form.Item
          name="name"
          label="配置名称"
          rules={[{ required: true, message: '请输入配置名称' }]}
        >
          <Input placeholder="例如：生产环境通知" />
        </Form.Item>

        <Form.Item
          name="webhook_url"
          label="Webhook URL"
          rules={[{ required: true, message: '请输入 Webhook URL' }]}
          extra="Webhook 地址"
        >
          <Input placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..." />
        </Form.Item>

        <Form.Item name="description" label="描述">
          <TextArea rows={2} placeholder="配置描述（可选）" />
        </Form.Item>

        <Form.Item
          name="enabled"
          label="启用配置"
          valuePropName="checked"
          extra="禁用后此配置将不会被使用"
        >
          <Switch />
        </Form.Item>

        <Form.Item
          name="is_default"
          label="设为默认"
          valuePropName="checked"
          extra="设为默认后，其他配置的默认状态将被取消"
        >
          <Switch />
        </Form.Item>
      </Form>
    </Modal>
  );
}
