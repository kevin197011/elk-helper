// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect, useState } from 'react';
import { Modal, Form, Input, Select, Switch, App } from 'antd';
import { useMutation } from '@tanstack/react-query';
import { usersApi, User } from '../services/api';

interface UserEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user?: User | null;
  onSuccess?: () => void;
}

export default function UserEditDialog({
  open,
  onOpenChange,
  user,
  onSuccess,
}: UserEditDialogProps) {
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const [resetPasswordOpen, setResetPasswordOpen] = useState(false);
  const [resetForm] = Form.useForm();
  const isEdit = !!user;

  useEffect(() => {
    if (open) {
      if (user) {
        form.setFieldsValue({
          username: user.username,
          email: user.email || '',
          role: user.role,
          enabled: user.enabled,
        });
      } else {
        form.resetFields();
      }
    }
  }, [user, form, open]);

  const createMutation = useMutation({
    mutationFn: (values: {
      username: string;
      password: string;
      email?: string;
      role: 'admin' | 'user';
      enabled: boolean;
    }) => usersApi.create(values),
    onSuccess: () => {
      message.success('用户创建成功');
      onOpenChange(false);
      onSuccess?.();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '创建失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number;
      data: { email?: string; role: 'admin' | 'user'; enabled: boolean };
    }) => usersApi.update(id, data),
    onSuccess: () => {
      message.success('用户更新成功');
      onOpenChange(false);
      onSuccess?.();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '更新失败');
    },
  });

  const resetPasswordMutation = useMutation({
    mutationFn: ({ id, newPassword }: { id: number; newPassword: string }) =>
      usersApi.resetPassword(id, newPassword),
    onSuccess: () => {
      message.success('密码已重置');
      setResetPasswordOpen(false);
      resetForm.resetFields();
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '重置密码失败');
    },
  });

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (isEdit && user) {
        updateMutation.mutate({
          id: user.id,
          data: {
            email: values.email || '',
            role: values.role,
            enabled: values.enabled,
          },
        });
      } else {
        createMutation.mutate({
          username: values.username,
          password: values.password,
          email: values.email || '',
          role: values.role,
          enabled: values.enabled ?? true,
        });
      }
    } catch {
      // validation failed
    }
  };

  const handleResetPassword = async () => {
    if (!user) return;
    try {
      const values = await resetForm.validateFields();
      resetPasswordMutation.mutate({ id: user.id, newPassword: values.new_password });
    } catch {
      // validation failed
    }
  };

  return (
    <>
      <Modal
        title={isEdit ? '编辑用户' : '新建用户'}
        open={open}
        onCancel={() => onOpenChange(false)}
        onOk={handleSubmit}
        confirmLoading={createMutation.isPending || updateMutation.isPending}
        okText="保存"
        cancelText="取消"
        width={480}
        destroyOnClose
        footer={(_, { OkBtn, CancelBtn }) => (
          <>
            {isEdit && (
              <a
                style={{ float: 'left', lineHeight: '32px' }}
                onClick={() => setResetPasswordOpen(true)}
              >
                重置密码
              </a>
            )}
            <CancelBtn />
            <OkBtn />
          </>
        )}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{ role: 'user', enabled: true }}
          style={{ marginTop: 16 }}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input disabled={isEdit} placeholder="登录用户名" />
          </Form.Item>
          {!isEdit && (
            <Form.Item
              name="password"
              label="密码"
              rules={[
                { required: true, message: '请输入密码' },
                { min: 6, message: '密码至少 6 位' },
              ]}
            >
              <Input.Password placeholder="至少 6 位" />
            </Form.Item>
          )}
          <Form.Item name="email" label="邮箱">
            <Input placeholder="可选" />
          </Form.Item>
          <Form.Item
            name="role"
            label="角色"
            rules={[{ required: true, message: '请选择角色' }]}
          >
            <Select
              options={[
                { value: 'user', label: '普通用户' },
                { value: 'admin', label: '管理员' },
              ]}
            />
          </Form.Item>
          <Form.Item name="enabled" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`重置密码：${user?.username ?? ''}`}
        open={resetPasswordOpen}
        onCancel={() => setResetPasswordOpen(false)}
        onOk={handleResetPassword}
        confirmLoading={resetPasswordMutation.isPending}
        okText="确认重置"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={resetForm} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            name="new_password"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少 6 位' },
            ]}
          >
            <Input.Password placeholder="至少 6 位" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
