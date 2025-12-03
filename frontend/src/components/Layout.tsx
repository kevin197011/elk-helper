// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, AlertCircle, Settings, Database, MessageSquare, LogOut, Shield, Trash2, Sparkles, Activity, Key } from 'lucide-react';
import { Sidebar, SidebarHeader, SidebarContent, SidebarFooter, SidebarItem } from './ui/sidebar';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from './ui/dropdown-menu';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import ChangePasswordDialog from './ChangePasswordDialog';

interface LayoutProps {
  children: React.ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const { user, logout } = useAuth();
  const { toast } = useToast();
  const navigate = useNavigate();
  const [changePasswordOpen, setChangePasswordOpen] = useState(false);

  const menuItems = [
    {
      path: '/',
      label: '仪表盘',
      icon: LayoutDashboard,
    },
    {
      path: '/rules',
      label: '规则管理',
      icon: Settings,
    },
    {
      path: '/alerts',
      label: '告警历史',
      icon: AlertCircle,
    },
    {
      path: '/es-configs',
      label: '数据源配置',
      icon: Database,
    },
    {
      path: '/lark-configs',
      label: 'Lark 配置',
      icon: MessageSquare,
    },
    {
      path: '/cleanup-config',
      label: '清理任务',
      icon: Trash2,
    },
  ];

  const handleLogout = () => {
    logout();
    toast({
      title: '已登出',
      description: '您已成功退出登录',
    });
    navigate('/login');
  };

  return (
    <div className="flex h-screen overflow-hidden bg-gradient-to-br from-slate-50 via-blue-50/30 to-slate-50">
      <Sidebar className="border-r border-border/40 bg-card/50 backdrop-blur-sm shadow-lg">
        <SidebarHeader className="border-b border-border/40 pb-4">
          <Link to="/" className="flex items-center gap-3 px-2 cursor-pointer hover:opacity-80 transition-opacity">
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-primary to-primary/80 text-primary-foreground shadow-lg shadow-primary/20">
              <Activity className="h-5 w-5" />
            </div>
            <div className="flex flex-col">
              <span className="font-bold text-lg bg-gradient-to-r from-primary to-primary/70 bg-clip-text text-transparent">
                ELK Helper
              </span>
              <span className="text-xs text-muted-foreground">智能告警系统</span>
            </div>
          </Link>
        </SidebarHeader>
        <SidebarContent className="pt-4">
          <nav className="space-y-1 px-2">
            {menuItems.map((item) => {
              const Icon = item.icon;
              const isActive = location.pathname === item.path ||
                              (item.path !== '/' && location.pathname.startsWith(item.path));
              return (
                <Link key={item.path} to={item.path}>
                  <SidebarItem active={isActive} className="rounded-lg transition-all duration-200 hover:bg-accent/50">
                    <Icon className={`h-5 w-5 transition-transform ${isActive ? 'scale-110' : ''}`} />
                    <span className="font-medium">{item.label}</span>
                  </SidebarItem>
                </Link>
              );
            })}
          </nav>
        </SidebarContent>
        <SidebarFooter className="border-t border-border/40 p-4">
          <div className="rounded-lg bg-gradient-to-r from-primary/5 to-primary/10 p-3 border border-primary/10">
            <p className="text-xs font-medium text-primary text-center">系统运维部驱动</p>
          </div>
        </SidebarFooter>
      </Sidebar>

      <div className="flex-1 flex flex-col overflow-hidden">
        <header className="flex-shrink-0 h-16 border-b border-border/40 flex items-center justify-between px-6 bg-card/80 backdrop-blur-sm shadow-sm">
          <div className="flex items-center gap-3">
            <h1 className="text-xl font-bold bg-gradient-to-r from-foreground to-foreground/70 bg-clip-text text-transparent">
              ELK 关键字告警系统
            </h1>
            <div className="h-4 w-px bg-border/60" />
            <span className="text-sm text-muted-foreground">实时监控 · 智能分析</span>
          </div>
          <div className="flex items-center gap-4">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" className="flex items-center gap-2 h-auto py-2 px-3 rounded-lg hover:bg-accent/50 transition-colors">
                  <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-primary to-primary/80 text-primary-foreground text-sm font-medium shadow-md shadow-primary/20 ring-2 ring-primary/10">
                    {user?.username?.charAt(0).toUpperCase() || 'U'}
                  </div>
                  <div className="flex flex-col items-start">
                    <span className="text-sm font-semibold">{user?.username || 'User'}</span>
                    <div className="flex items-center gap-1">
                      {user?.role === 'admin' && (
                        <Badge variant="secondary" className="h-4 px-1.5 text-xs bg-primary/10 text-primary border-primary/20">
                          <Shield className="h-3 w-3 mr-0.5" />
                          管理员
                        </Badge>
                      )}
                    </div>
                  </div>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-56 shadow-lg">
                <DropdownMenuLabel>
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">{user?.username}</p>
                    {user?.email && (
                      <p className="text-xs leading-none text-muted-foreground">{user.email}</p>
                    )}
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() => setChangePasswordOpen(true)}
                  className="cursor-pointer"
                >
                  <Key className="mr-2 h-4 w-4" />
                  <span>修改密码</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout} className="cursor-pointer text-red-600 focus:text-red-600 focus:bg-red-50">
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>退出登录</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <main className="flex-1 overflow-y-auto overflow-x-hidden p-6 bg-gradient-to-b from-transparent to-muted/20">
          <div className="max-w-7xl mx-auto">
            {children}
          </div>
        </main>
        <footer className="flex-shrink-0 border-t border-border/40 bg-card/50 backdrop-blur-sm px-6 py-3">
          <div className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
            <Sparkles className="h-4 w-4 text-primary/60" />
            <span>由 <span className="font-semibold text-primary">系统运维部</span> 驱动开发</span>
          </div>
        </footer>
      </div>
      <ChangePasswordDialog open={changePasswordOpen} onOpenChange={setChangePasswordOpen} />
    </div>
  );
}
