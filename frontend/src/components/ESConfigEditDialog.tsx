// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect } from 'react';
import { Modal, Form, Input, Switch, App } from 'antd';
import { useMutation } from '@tanstack/react-query';
import { esConfigApi, ESConfig } from '../services/api';

const { TextArea } = Input;

interface ESConfigEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config?: ESConfig | null;
  onSuccess?: () => void;
}

export default function ESConfigEditDialog({
  open,
  onOpenChange,
  config,
  onSuccess,
}: ESConfigEditDialogProps) {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const isEdit = !!config;

  useEffect(() => {
    if (open) {
      if (config) {
        form.setFieldsValue({
          ...config,
          password: '',
          ca_certificate: '',
        });
      } else {
        form.resetFields();
      }
    }
  }, [config, form, open]);

  const createMutation = useMutation({
    mutationFn: (data: Partial<ESConfig>) => esConfigApi.create(data),
    onSuccess: () => {
      message.success('创建成功');
      onSuccess?.();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '创建失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<ESConfig> }) =>
      esConfigApi.update(id, data),
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
      
      // Ensure password field is always included, even if empty
      // This is important for creation - empty password should be saved as empty string
      if (!isEdit && values.password === undefined) {
        values.password = '';
      }
      
      if (isEdit && config) {
        updateMutation.mutate({ id: config.id, data: values });
      } else {
        createMutation.mutate(values);
      }
    } catch (error) {
      // Form validation failed
    }
  };

  const handleUrlChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value.trim().toLowerCase();
    if (value.includes('https://')) {
      form.setFieldsValue({ use_ssl: true });
    } else if (value.startsWith('http://') && !value.includes('https://')) {
      form.setFieldsValue({ use_ssl: false });
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑数据源配置' : '新建数据源配置'}
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
          use_ssl: false,
          skip_verify: false,
        }}
        style={{ marginTop: 16 }}
      >
        <Form.Item
          name="name"
          label="配置名称"
          rules={[{ required: true, message: '请输入配置名称' }]}
        >
          <Input placeholder="例如：生产环境 ES" />
        </Form.Item>

        <Form.Item
          name="url"
          label="ES 地址"
          rules={[
            { required: true, message: '请输入 ES 地址' },
            {
              validator: (_, value) => {
                if (!value) return Promise.resolve();
                const urls = value.split(';').map((u: string) => u.trim()).filter((u: string) => u);
                for (const url of urls) {
                  if (!url.startsWith('http://') && !url.startsWith('https://')) {
                    return Promise.reject(`地址必须以 http:// 或 https:// 开头: ${url}`);
                  }
                }
                return Promise.resolve();
              },
            },
          ]}
          extra={
            <div style={{ fontSize: 12 }}>
              <div>• 单节点: 直接输入地址，如 https://10.170.1.54:9200</div>
              <div>• 多节点: 用分号分隔多个地址，系统会轮询查询</div>
            </div>
          }
        >
          <TextArea
            rows={2}
            placeholder="https://10.170.1.54:9200 或 https://es1:9200;https://es2:9200"
            onChange={handleUrlChange}
          />
        </Form.Item>

        <Form.Item name="username" label="用户名（可选）">
          <Input placeholder="留空则使用无认证连接" />
        </Form.Item>

        <Form.Item
          name="password"
          label={isEdit ? '密码（留空则不修改）' : '密码'}
          extra={isEdit ? '当前已配置密码。留空则保持原密码不变' : undefined}
        >
          <Input.Password placeholder={isEdit ? '••••••••' : '密码'} />
        </Form.Item>

        <Form.Item
          name="use_ssl"
          label="启用 SSL/TLS"
          valuePropName="checked"
          extra="启用后使用 HTTPS 加密连接"
        >
          <Switch />
        </Form.Item>

        <Form.Item noStyle shouldUpdate={(prev, cur) => prev.use_ssl !== cur.use_ssl}>
          {({ getFieldValue }) =>
            getFieldValue('use_ssl') && (
              <>
                <Form.Item
                  name="skip_verify"
                  label="跳过证书验证"
                  valuePropName="checked"
                  extra="仅用于开发/测试环境，生产环境不建议使用"
                >
                  <Switch />
                </Form.Item>

                <Form.Item noStyle shouldUpdate={(prev, cur) => prev.skip_verify !== cur.skip_verify}>
                  {({ getFieldValue: getValue }) =>
                    !getValue('skip_verify') && (
                      <Form.Item
                        name="ca_certificate"
                        label={`CA 证书${isEdit ? '（留空则不修改）' : ''}（可选）`}
                        extra="用于验证服务器证书的 CA 证书（PEM 格式）"
                      >
                        <TextArea rows={4} placeholder="PEM 格式的 CA 证书内容" />
                      </Form.Item>
                    )
                  }
                </Form.Item>
              </>
            )
          }
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
