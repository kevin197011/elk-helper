// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState } from 'react';
import { Table, Tag, Button, Space, Typography, App } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { usersApi, User } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import UserEditDialog from '../components/UserEditDialog';
import PageHeader from '../components/PageHeader';

const { Text } = Typography;

function formatTime(value?: string) {
  if (!value) return '-';
  return new Date(value).toLocaleString();
}

export default function UsersPage() {
  const queryClient = useQueryClient();
  const { message, modal } = App.useApp();
  const { user: currentUser } = useAuth();
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);

  const {
    data,
    isLoading,
    isError,
    error,
    refetch,
  } = useQuery({
    queryKey: ['users'],
    queryFn: () => usersApi.getAll().then(res => res.data.data),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => usersApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      message.success('用户已删除');
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '删除失败');
    },
  });

  const handleDelete = (record: User) => {
    modal.confirm({
      title: '确认删除',
      icon: <ExclamationCircleOutlined />,
      content: `确定要删除用户 "${record.username}" 吗？此操作不可恢复。`,
      okText: '确认删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: () => deleteMutation.mutateAsync(record.id),
    });
  };

  const columns: ColumnsType<User> = [
    { title: 'ID', dataIndex: 'id', width: 70 },
    {
      title: '用户名',
      dataIndex: 'username',
      render: (name: string) => <Text strong>{name}</Text>,
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      render: (email: string) => email || '-',
    },
    {
      title: '角色',
      dataIndex: 'role',
      width: 100,
      render: (role: string) => (
        <Tag color={role === 'admin' ? 'blue' : 'default'}>
          {role === 'admin' ? '管理员' : '普通用户'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 90,
      render: (enabled: boolean) => (
        <Tag color={enabled ? 'success' : 'default'}>{enabled ? '启用' : '禁用'}</Tag>
      ),
    },
    {
      title: '最后登录',
      dataIndex: 'last_login_at',
      width: 180,
      render: formatTime,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 180,
      render: formatTime,
    },
    {
      title: '操作',
      width: 140,
      render: (_, record) => {
        const isSelf = record.id === currentUser?.id;
        return (
          <Space>
            <Button
              type="text"
              icon={<EditOutlined />}
              onClick={() => {
                setSelectedUser(record);
                setEditModalOpen(true);
              }}
            />
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              disabled={isSelf}
              title={isSelf ? '不能删除当前登录账号' : undefined}
              onClick={() => handleDelete(record)}
            />
          </Space>
        );
      },
    },
  ];

  return (
    <div>
      <PageHeader
        title="用户管理"
        extra={
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setSelectedUser(null);
              setEditModalOpen(true);
            }}
          >
            新建用户
          </Button>
        }
      />
      {isError && (
        <div style={{ marginBottom: 16 }}>
          <Text type="danger">
            {(error as any)?.response?.data?.error || '加载用户列表失败'}
          </Text>
          <Button type="link" onClick={() => refetch()}>
            重试
          </Button>
        </div>
      )}
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data ?? []}
        loading={isLoading}
        pagination={false}
      />
      <UserEditDialog
        open={editModalOpen}
        onOpenChange={setEditModalOpen}
        user={selectedUser}
        onSuccess={() => queryClient.invalidateQueries({ queryKey: ['users'] })}
      />
    </div>
  );
}
