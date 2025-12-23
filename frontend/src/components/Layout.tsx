// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { Layout as AntLayout, Menu, Dropdown, Avatar, Button, Space, Typography, App } from 'antd';
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
      label: <Link to="/lark-configs">Lark 配置</Link>,
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
    <AntLayout style={{ minHeight: '100vh' }}>
      <Sider 
        trigger={null} 
        collapsible 
        collapsed={collapsed}
        theme="dark"
        width={220}
      >
        <div style={{ 
          height: 64, 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: collapsed ? 'center' : 'flex-start',
          padding: collapsed ? 0 : '0 16px',
          borderBottom: '1px solid rgba(255,255,255,0.1)'
        }}>
          <Link to="/" style={{ display: 'flex', alignItems: 'center', gap: 10, textDecoration: 'none' }}>
            {collapsed ? (
              <img src="/favicon.svg" alt="ELK Helper" style={{ width: 36, height: 36 }} />
            ) : (
              <img src="/logo.svg" alt="ELK Helper" style={{ height: 36, width: 'auto' }} />
            )}
          </Link>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={getSelectedKeys()}
          items={menuItems}
          style={{ borderRight: 0, marginTop: 8 }}
        />
        <div style={{
          position: 'absolute',
          bottom: 16,
          left: 0,
          right: 0,
          textAlign: 'center',
          color: 'rgba(255,255,255,0.45)',
          fontSize: 11
        }}>
          {!collapsed && '系统运维部驱动'}
        </div>
      </Sider>
      <AntLayout>
        <Header style={{ 
          padding: '0 24px', 
          background: '#fff', 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'space-between',
          boxShadow: '0 1px 4px rgba(0,0,0,0.08)'
        }}>
          <Space>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{ fontSize: 16, width: 48, height: 48 }}
            />
            <Text strong style={{ fontSize: 16 }}>ELK 关键字告警系统</Text>
            <Text type="secondary" style={{ fontSize: 13 }}>实时监控 · 智能分析</Text>
          </Space>
          <Dropdown menu={{ items: userMenuItems }} trigger={['click']} placement="bottomRight">
            <Button type="text" style={{ height: 'auto', padding: '4px 8px' }}>
              <Space>
                <Avatar 
                  style={{ backgroundColor: '#1677ff' }}
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
          padding: 24, 
          background: '#fff', 
          borderRadius: 8,
          minHeight: 280,
          overflow: 'auto'
        }}>
          {children}
        </Content>
        <Footer style={{ textAlign: 'center', padding: '12px 24px', color: '#999' }}>
          由 <Text strong style={{ color: '#1677ff' }}>系统运维部</Text> 驱动开发
        </Footer>
      </AntLayout>
      <ChangePasswordDialog open={changePasswordOpen} onOpenChange={setChangePasswordOpen} />
    </AntLayout>
  );
}
