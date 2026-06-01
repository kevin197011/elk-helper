// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect, useMemo, useState } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import { Form, Input, Button, Typography, App, theme, Card, Space, Divider } from 'antd';
import { UserOutlined, LockOutlined, CloudOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import { ssoApi, SSOProviderPublic } from '../services/api';

const { Title, Text } = Typography;

export default function LoginPage() {
  const [loading, setLoading] = useState(false);
  const [ssoProviders, setSsoProviders] = useState<SSOProviderPublic[]>([]);
  const { login, isAuthenticated, authStatus } = useAuth();
  const { message } = App.useApp();
  const navigate = useNavigate();
  const location = useLocation();
  const { token } = theme.useToken();
  const redirectTo = useMemo(() => {
    const from = (location.state as any)?.from?.pathname;
    return typeof from === 'string' && from.startsWith('/') ? from : '/';
  }, [location.state]);

  useEffect(() => {
    if (authStatus === 'authenticated' && isAuthenticated) {
      navigate(redirectTo, { replace: true });
    }
  }, [authStatus, isAuthenticated, navigate, redirectTo]);

  useEffect(() => {
    void (async () => {
      try {
        const res = await ssoApi.listPublic();
        setSsoProviders(res.data.data || []);
      } catch {
        // SSO list is optional on the login page
      }
    })();
  }, []);

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功，欢迎回来！');
      navigate(redirectTo, { replace: true });
    } catch (error: any) {
      message.error(error.response?.data?.error || '用户名或密码错误');
    } finally {
      setLoading(false);
    }
  };

  if (authStatus === 'authenticated' && isAuthenticated) {
    return <Navigate to={redirectTo} replace />;
  }

  return (
    <div
      className="app-login-bg"
      style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        padding: 24,
      }}
    >
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', width: '100%' }}>
        <div style={{ width: '100%', maxWidth: 420 }}>
          <div style={{ textAlign: 'center', marginBottom: 18 }}>
            <Space direction="vertical" size={6}>
              <div
                style={{
                  width: 56,
                  height: 56,
                  borderRadius: 14,
                  background: token.colorBgContainer,
                  border: `1px solid ${token.colorBorderSecondary}`,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  margin: '0 auto',
                }}
              >
                <img src="/favicon.svg" alt="ELK Helper" style={{ width: 32, height: 32 }} />
              </div>
              <Title level={3} style={{ margin: 0 }}>
                ELK Helper
              </Title>
              <Text type="secondary">智能告警系统</Text>
            </Space>
          </div>

          <Card className="app-page-card" styles={{ body: { padding: 24 } }}>
            <Form name="login" onFinish={handleSubmit} size="large" autoComplete="off" layout="vertical">
              <Form.Item name="username" label="用户名" rules={[{ required: true, message: '请输入用户名' }]}>
                <Input prefix={<UserOutlined style={{ color: token.colorTextTertiary }} />} placeholder="请输入用户名" disabled={loading} />
              </Form.Item>

              <Form.Item name="password" label="密码" rules={[{ required: true, message: '请输入密码' }]}>
                <Input.Password prefix={<LockOutlined style={{ color: token.colorTextTertiary }} />} placeholder="请输入密码" disabled={loading} />
              </Form.Item>

              <Form.Item style={{ marginBottom: 0, marginTop: 8 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={loading}
                  block
                  size="large"
                  style={{ height: 44, fontWeight: 600 }}
                >
                  登录
                </Button>
              </Form.Item>
            </Form>

            {ssoProviders.length > 0 && (
              <>
                <Divider plain style={{ marginTop: 8 }}>
                  <span style={{ fontSize: 12, color: token.colorTextTertiary }}>或使用单点登录</span>
                </Divider>
                <Space direction="vertical" style={{ width: '100%' }}>
                  {ssoProviders
                    .filter((p) => p.type === 'oidc')
                    .map((p) => (
                      <Button
                        key={`oidc-${p.id}`}
                        block
                        size="large"
                        icon={<CloudOutlined />}
                        onClick={() => {
                          window.location.href = ssoApi.oidcLoginURL(p.id);
                        }}
                      >
                        使用 {p.name} 登录
                      </Button>
                    ))}
                </Space>
              </>
            )}
          </Card>
        </div>
      </div>

      <div style={{ textAlign: 'center', paddingTop: 8, paddingBottom: 4 }}>
        <Text type="secondary" style={{ fontSize: 12 }}>
          <Text style={{ color: token.colorTextSecondary, fontWeight: 500 }}>系统运行部驱动</Text>
        </Text>
      </div>
    </div>
  );
}
