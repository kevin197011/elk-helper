// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Form,
  Input,
  Modal,
  Popconfirm,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  App,
} from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { ssoApi, SSOAdminProvider, SSOProviderConfig } from '../services/api';

const { Text } = Typography;

const DEFAULT_OIDC_CONFIG: SSOProviderConfig = {
  issuer: '',
  client_id: '',
  client_secret: '',
  redirect_url: '',
  scopes: ['openid', 'profile', 'email'],
  username_claim: 'preferred_username',
  skip_tls_verify: false,
};

type FormValues = {
  name: string;
  enabled: boolean;
  issuer: string;
  client_id: string;
  client_secret: string;
  redirect_url: string;
  scopes_str: string;
  username_claim: string;
  skip_tls_verify: boolean;
};

function flatten(cfg: SSOProviderConfig): Partial<FormValues> {
  return {
    issuer: cfg.issuer,
    client_id: cfg.client_id,
    client_secret: cfg.client_secret,
    redirect_url: cfg.redirect_url,
    scopes_str: (cfg.scopes || ['openid', 'profile', 'email']).join(','),
    username_claim: cfg.username_claim || 'preferred_username',
    skip_tls_verify: !!cfg.skip_tls_verify,
  };
}

function unflatten(values: FormValues): SSOProviderConfig {
  const scopes = (values.scopes_str || 'openid,profile,email')
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
  return {
    issuer: values.issuer || '',
    client_id: values.client_id || '',
    client_secret: values.client_secret || '',
    redirect_url: values.redirect_url || '',
    scopes,
    username_claim: values.username_claim || 'preferred_username',
    skip_tls_verify: !!values.skip_tls_verify,
  };
}

export default function SsoConfigPage() {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [list, setList] = useState<SSOAdminProvider[]>([]);
  const [open, setOpen] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form] = Form.useForm<FormValues>();

  const fetchList = async () => {
    setLoading(true);
    try {
      const res = await ssoApi.listAdmin();
      setList(res.data.data || []);
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to load SSO providers');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void fetchList();
  }, []);

  const onAdd = () => {
    setEditingId(null);
    form.resetFields();
    form.setFieldsValue({
      name: 'OIDC',
      enabled: true,
      ...flatten(DEFAULT_OIDC_CONFIG),
    } as FormValues);
    setOpen(true);
  };

  const onEdit = (record: SSOAdminProvider) => {
    setEditingId(record.id);
    form.resetFields();
    form.setFieldsValue({
      name: record.name,
      enabled: record.enabled,
      ...flatten({ ...DEFAULT_OIDC_CONFIG, ...(record.config || {}) }),
    } as FormValues);
    setOpen(true);
  };

  const onToggle = async (record: SSOAdminProvider) => {
    try {
      await ssoApi.toggle(record.id);
      message.success('Status updated');
      void fetchList();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to update status');
    }
  };

  const onDelete = async (record: SSOAdminProvider) => {
    try {
      await ssoApi.delete(record.id);
      message.success('Deleted');
      void fetchList();
    } catch (err: any) {
      message.error(err.response?.data?.error || 'Failed to delete');
    }
  };

  const onSave = async () => {
    try {
      const values = await form.validateFields();
      const payload = {
        type: 'oidc' as const,
        name: values.name,
        enabled: !!values.enabled,
        config: unflatten(values),
      };
      if (editingId) {
        await ssoApi.update(editingId, payload);
        message.success('Updated');
      } else {
        await ssoApi.create(payload);
        message.success('Created');
      }
      setOpen(false);
      void fetchList();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err.response?.data?.error || 'Failed to save');
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    {
      title: 'Type',
      dataIndex: 'type',
      width: 90,
      render: () => <Tag color="blue">OIDC</Tag>,
    },
    { title: 'Name', dataIndex: 'name' },
    {
      title: 'Status',
      dataIndex: 'enabled',
      width: 100,
      render: (v: boolean, record: SSOAdminProvider) => (
        <Switch checked={v} onChange={() => void onToggle(record)} checkedChildren="On" unCheckedChildren="Off" />
      ),
    },
    {
      title: 'Callback path',
      key: 'callback',
      render: (_: unknown, record: SSOAdminProvider) => (
        <Text copyable code style={{ fontSize: 12 }}>
          {`/api/v1/auth/sso/oidc/${record.id}/callback`}
        </Text>
      ),
    },
    {
      title: 'Actions',
      key: 'action',
      width: 200,
      render: (_: unknown, record: SSOAdminProvider) => (
        <Space>
          <Button size="small" icon={<EditOutlined />} onClick={() => onEdit(record)}>
            Edit
          </Button>
          <Popconfirm title="Delete this provider?" onConfirm={() => void onDelete(record)}>
            <Button size="small" danger icon={<DeleteOutlined />}>
              Delete
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <Card
      title="SSO (OIDC)"
      extra={
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => void fetchList()}>
            Refresh
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={onAdd}>
            Add OIDC
          </Button>
        </Space>
      }
    >
      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="Notes"
        description={
          <>
            <div>Only OpenID Connect (OIDC) is supported.</div>
            <div>
              Register callback URL with your IdP, typically{' '}
              <code>https://your-host/api/v1/auth/sso/oidc/&lt;id&gt;/callback</code>.
            </div>
            <div>New SSO users are created with the default user role; promote to admin in User Management if needed.</div>
          </>
        }
      />

      <Table rowKey="id" loading={loading} columns={columns} dataSource={list} pagination={false} />

      <Modal
        title={editingId ? 'Edit OIDC provider' : 'Add OIDC provider'}
        open={open}
        onOk={() => void onSave()}
        onCancel={() => setOpen(false)}
        okText="Save"
        width={700}
        destroyOnClose
      >
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="Display name" rules={[{ required: true }]}>
            <Input placeholder="e.g. Company Keycloak" />
          </Form.Item>
          <Form.Item name="enabled" label="Enabled" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="issuer" label="Issuer URL" rules={[{ required: true }]}>
            <Input placeholder="https://accounts.google.com" />
          </Form.Item>
          <Form.Item name="client_id" label="Client ID" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="client_secret" label="Client Secret" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="redirect_url"
            label="Redirect URL"
            rules={[{ required: true }]}
            tooltip="Must match the URL registered at the IdP"
          >
            <Input placeholder="https://your.host/api/v1/auth/sso/oidc/1/callback" />
          </Form.Item>
          <Form.Item name="scopes_str" label="Scopes" tooltip="Comma-separated; default openid,profile,email">
            <Input placeholder="openid,profile,email" />
          </Form.Item>
          <Form.Item name="username_claim" label="Username claim">
            <Input placeholder="preferred_username" />
          </Form.Item>
          <Form.Item name="skip_tls_verify" label="Skip TLS verify" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  );
}
