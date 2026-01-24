// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useMemo, useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Layout as AntLayout, Menu, Dropdown, Avatar, Button, Space, Typography, App, theme } from 'antd';
import type { MenuProps } from 'antd';
import {
  DashboardOutlined,
  AlertOutlined,
  SettingOutlined,
  DatabaseOutlined,
  MessageOutlined,
  DeleteOutlined,
  LogoutOutlined,
  UserOutlined,
  KeyOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
} from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';
import ChangePasswordDialog from './ChangePasswordDialog';

const { Header, Sider, Content, Footer } = AntLayout;
const { Text } = Typography;

interface LayoutProps {
  children: React.ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const { user, logout } = useAuth();
  const { message } = App.useApp();
  const navigate = useNavigate();
  const [collapsed, setCollapsed] = useState(false);
  const [changePasswordOpen, setChangePasswordOpen] = useState(false);
  const { token } = theme.useToken();

  const menuItems: MenuProps['items'] = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: <Link to="/">仪表盘</Link>,
    },
    {
      key: '/rules',
      icon: <SettingOutlined />,
      label: <Link to="/rules">规则管理</Link>,
    },
    {
      key: '/alerts',
      icon: <AlertOutlined />,
      label: <Link to="/alerts">告警历史</Link>,
    },
    {
      key: '/es-configs',
      icon: <DatabaseOutlined />,
      label: <Link to="/es-configs">数据源配置</Link>,
    },
    {
      key: '/lark-configs',
      icon: <MessageOutlined />,
      label: <Link to="/lark-configs">告警配置</Link>,
    },
    {
      key: '/cleanup-config',
      icon: <DeleteOutlined />,
      label: <Link to="/cleanup-config">清理任务</Link>,
    },
  ];

  const getSelectedKeys = () => {
    const path = location.pathname;
    if (path === '/') return ['/'];
    // Match parent path for sub-routes
    const parentPath = menuItems.find(item => 
      item?.key !== '/' && path.startsWith(item?.key as string)
    );
    return parentPath ? [parentPath.key as string] : [path];
  };

  const selectedKeys = useMemo(() => getSelectedKeys(), [location.pathname]);

  const handleLogout = () => {
    logout();
    message.success('已成功退出登录');
    navigate('/login');
  };

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'user-info',
      label: (
        <div style={{ padding: '4px 0' }}>
          <div style={{ fontWeight: 500 }}>{user?.username}</div>
          {user?.email && <div style={{ fontSize: 12, color: '#999' }}>{user.email}</div>}
        </div>
      ),
      disabled: true,
    },
    { type: 'divider' },
    {
      key: 'change-password',
      icon: <KeyOutlined />,
      label: '修改密码',
      onClick: () => setChangePasswordOpen(true),
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
      onClick: handleLogout,
    },
  ];

  return (
    <AntLayout style={{ minHeight: '100vh', background: token.colorBgLayout }}>
      <Sider 
        trigger={null} 
        collapsible 
        collapsed={collapsed}
        theme="light"
        width={240}
        style={{
          background: token.colorBgContainer,
          borderRight: `1px solid ${token.colorBorderSecondary}`,
        }}
      >
        <div
          style={{
            padding: collapsed ? '24px 8px' : '24px 20px',
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            margin: collapsed ? '8px 8px 0' : '8px 8px 0',
          }}
        >
          <Link to="/" className="app-link" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            {collapsed ? (
              <img src="/favicon.svg" alt="ELK Helper" style={{ width: 36, height: 36 }} />
            ) : (
              <img src="/logo.svg" alt="ELK Helper" style={{ height: 36, width: 'auto' }} />
            )}
          </Link>
        </div>
        <Menu
          theme="light"
          mode="inline"
          selectedKeys={selectedKeys}
          items={menuItems}
          style={{ borderRight: 0, background: 'transparent', padding: 8 }}
        />
        <div style={{
          position: 'absolute',
          bottom: 16,
          left: 0,
          right: 0,
          textAlign: 'center',
          color: token.colorTextSecondary,
          fontSize: 11
        }}>
          {/* intentionally empty */}
        </div>
      </Sider>
      <AntLayout>
        <Header style={{ 
          padding: '0 32px', 
          background: token.colorBgContainer, 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'space-between',
          borderBottom: `1px solid ${token.colorBorderSecondary}`,
          height: 64,
          position: 'sticky',
          top: 0,
          zIndex: 100,
        }}>
          <Space>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{ fontSize: 16, width: 48, height: 48 }}
            />
            <Text strong style={{ fontSize: 16 }}>ELK智能告警系统</Text>
          </Space>
          <Dropdown menu={{ items: userMenuItems }} trigger={['click']} placement="bottomRight">
            <Button type="text" style={{ height: 'auto', padding: '4px 8px' }}>
              <Space>
                <Avatar 
                  style={{ backgroundColor: token.colorPrimary }}
                  icon={<UserOutlined />}
                >
                  {user?.username?.charAt(0).toUpperCase()}
                </Avatar>
                <span>{user?.username}</span>
              </Space>
            </Button>
          </Dropdown>
        </Header>
        <Content style={{ 
          margin: 24,
          minHeight: 280,
          overflow: 'auto'
        }}>
          <div className="app-content">
            <div className="app-page-card app-page-card--padded">{children}</div>
          </div>
        </Content>
        <Footer style={{ textAlign: 'center', padding: '20px 24px', color: token.colorTextSecondary, background: token.colorBgContainer, borderTop: `1px solid ${token.colorBorderSecondary}` }}>
          <Text style={{ color: token.colorTextSecondary, fontWeight: 500 }}>系统运行部驱动</Text>
        </Footer>
      </AntLayout>
      <ChangePasswordDialog open={changePasswordOpen} onOpenChange={setChangePasswordOpen} />
    </AntLayout>
  );
}
