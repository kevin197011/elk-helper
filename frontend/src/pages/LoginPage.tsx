// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Form, Input, Button, Typography, App } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';

const { Title, Text } = Typography;

export default function LoginPage() {
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const { message } = App.useApp();
  const navigate = useNavigate();

  const handleSubmit = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功，欢迎回来！');
      navigate('/');
    } catch (error: any) {
      message.error(error.response?.data?.error || '用户名或密码错误');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      padding: 24,
      position: 'relative',
      overflow: 'hidden'
    }}>
      {/* DevOps 扫描风格背景 */}
      <style>{`
        @keyframes scanLine {
          0% { transform: translateY(-100vh); opacity: 0; }
          5% { opacity: 1; }
          95% { opacity: 1; }
          100% { transform: translateY(100vh); opacity: 0; }
        }
        @keyframes scanLineHorizontal {
          0% { transform: translateX(-100vw); opacity: 0; }
          5% { opacity: 1; }
          95% { opacity: 1; }
          100% { transform: translateX(100vw); opacity: 0; }
        }
        @keyframes radarScan {
          0% { transform: rotate(0deg); opacity: 0.8; }
          50% { opacity: 1; }
          100% { transform: rotate(360deg); opacity: 0.8; }
        }
        @keyframes gridScan {
          0% { background-position: 0 0; }
          100% { background-position: 0 100px; }
        }
        @keyframes pulse {
          0%, 100% { opacity: 0.3; transform: scale(1); }
          50% { opacity: 0.7; transform: scale(1.1); }
        }
        @keyframes codeFlow {
          0% { transform: translateY(-100vh) translateX(0); opacity: 0; }
          10% { opacity: 1; }
          90% { opacity: 1; }
          100% { transform: translateY(100vh) translateX(0); opacity: 0; }
        }
        @keyframes shimmer {
          0% { background-position: -200% 0; }
          100% { background-position: 200% 0; }
        }
        .bg-scan-line {
          animation: scanLine 3s linear infinite;
        }
        .bg-scan-horizontal {
          animation: scanLineHorizontal 4s linear infinite;
        }
        .bg-radar-scan {
          animation: radarScan 8s linear infinite;
        }
        .bg-grid-scan {
          background-image: linear-gradient(rgba(22, 119, 255, 0.1) 1px, transparent 1px),
                            linear-gradient(90deg, rgba(22, 119, 255, 0.1) 1px, transparent 1px);
          background-size: 50px 50px;
          animation: gridScan 20s linear infinite;
        }
        .bg-pulse {
          animation: pulse 4s ease-in-out infinite;
        }
        .bg-code-flow {
          animation: codeFlow 8s linear infinite;
        }
        .bg-shimmer {
          background: linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent);
          background-size: 200% 100%;
          animation: shimmer 3s ease-in-out infinite;
        }
      `}</style>

      {/* 扫描网格背景 */}
      <div
        className="bg-grid-scan"
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          opacity: 0.3,
          pointerEvents: 'none'
        }}
      />

      {/* 垂直扫描线 */}
      {Array.from({ length: 3 }).map((_, i) => (
        <div
          key={`scan-v-${i}`}
          className="bg-scan-line"
          style={{
            position: 'absolute',
            left: `${25 + i * 25}%`,
            top: 0,
            width: '2px',
            height: '100%',
            background: 'linear-gradient(to bottom, transparent, rgba(22, 119, 255, 0.6), transparent)',
            boxShadow: '0 0 10px rgba(22, 119, 255, 0.5)',
            animationDelay: `${i * 1}s`,
            animationDuration: '3s',
            pointerEvents: 'none'
          }}
        />
      ))}

      {/* 水平扫描线 */}
      {Array.from({ length: 2 }).map((_, i) => (
        <div
          key={`scan-h-${i}`}
          className="bg-scan-horizontal"
          style={{
            position: 'absolute',
            top: `${30 + i * 40}%`,
            left: 0,
            width: '100%',
            height: '2px',
            background: 'linear-gradient(to right, transparent, rgba(114, 46, 209, 0.6), transparent)',
            boxShadow: '0 0 10px rgba(114, 46, 209, 0.5)',
            animationDelay: `${i * 1.5}s`,
            animationDuration: '4s',
            pointerEvents: 'none'
          }}
        />
      ))}

      {/* 雷达扫描效果 */}
      <svg
        className="bg-radar-scan"
        style={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          transform: 'translate(-50%, -50%)',
          width: '400px',
          height: '400px',
          opacity: 0.15,
          pointerEvents: 'none'
        }}
        viewBox="0 0 200 200"
      >
        <defs>
          <linearGradient id="radarGradient" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="rgba(22, 119, 255, 0.8)" stopOpacity={1} />
            <stop offset="100%" stopColor="rgba(22, 119, 255, 0)" stopOpacity={0} />
          </linearGradient>
        </defs>
        {/* 扫描扇形 */}
        <path
          d="M 100 100 L 100 0 A 100 100 0 0 1 170 50 Z"
          fill="url(#radarGradient)"
          transform-origin="100 100"
        />
        {/* 同心圆 */}
        <circle cx="100" cy="100" r="80" fill="none" stroke="rgba(22, 119, 255, 0.2)" strokeWidth="1"/>
        <circle cx="100" cy="100" r="60" fill="none" stroke="rgba(22, 119, 255, 0.2)" strokeWidth="1"/>
        <circle cx="100" cy="100" r="40" fill="none" stroke="rgba(22, 119, 255, 0.2)" strokeWidth="1"/>
        <circle cx="100" cy="100" r="20" fill="none" stroke="rgba(22, 119, 255, 0.2)" strokeWidth="1"/>
      </svg>

      {/* 终端扫描数据流 */}
      {Array.from({ length: 8 }).map((_, i) => (
        <div
          key={`code-${i}`}
          className="bg-code-flow"
          style={{
            position: 'absolute',
            left: `${15 + i * 10}%`,
            fontFamily: 'monospace',
            fontSize: '11px',
            color: 'rgba(0, 255, 136, 0.4)',
            whiteSpace: 'nowrap',
            textShadow: '0 0 5px rgba(0, 255, 136, 0.5)',
            animationDelay: `${i * 0.6}s`,
            animationDuration: `${6 + Math.random() * 3}s`,
            pointerEvents: 'none',
            fontWeight: 500
          }}
        >
          {i % 4 === 0 ? '▶ SCANNING...' : 
           i % 4 === 1 ? '▶ ANALYZING...' : 
           i % 4 === 2 ? '▶ DEPLOYING...' : '▶ BUILDING...'}
        </div>
      ))}

      {/* 扫描检测点 */}
      {Array.from({ length: 20 }).map((_, i) => (
        <div
          key={`scan-point-${i}`}
          className="bg-pulse"
          style={{
            position: 'absolute',
            width: 8,
            height: 8,
            borderRadius: '50%',
            background: `rgba(0, 255, 136, ${0.3 + Math.random() * 0.4})`,
            left: `${Math.random() * 100}%`,
            top: `${Math.random() * 100}%`,
            animationDelay: `${Math.random() * 4}s`,
            boxShadow: `0 0 ${6 + Math.random() * 6}px rgba(0, 255, 136, 0.6)`,
            pointerEvents: 'none'
          }}
        />
      ))}

      {/* 动态光晕 */}
      <div
        className="bg-pulse"
        style={{
          position: 'absolute',
          top: '10%',
          right: '10%',
          width: 300,
          height: 300,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(255,255,255,0.15) 0%, transparent 70%)',
          filter: 'blur(40px)',
          animationDelay: '0s'
        }}
      />
      <div
        className="bg-pulse"
        style={{
          position: 'absolute',
          bottom: '15%',
          left: '15%',
          width: 250,
          height: 250,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(255,255,255,0.12) 0%, transparent 70%)',
          filter: 'blur(40px)',
          animationDelay: '2s'
        }}
      />
      <div
        className="bg-pulse"
        style={{
          position: 'absolute',
          top: '50%',
          left: '50%',
          width: 200,
          height: 200,
          borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 70%)',
          filter: 'blur(40px)',
          transform: 'translate(-50%, -50%)',
          animationDelay: '1s'
        }}
      />

      {/* 扫描数据连接线 */}
      <svg
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          opacity: 0.2,
          pointerEvents: 'none'
        }}
      >
        {/* 扫描路径线 */}
        {Array.from({ length: 12 }).map((_, i) => {
          const x1 = (i % 4) * 25 + 10;
          const y1 = Math.floor(i / 4) * 30 + 15;
          const x2 = ((i + 1) % 4) * 25 + 10;
          const y2 = Math.floor((i + 1) / 4) * 30 + 15;
          return (
            <line
              key={`scan-path-${i}`}
              x1={`${x1}%`}
              y1={`${y1}%`}
              x2={`${x2}%`}
              y2={`${y2}%`}
              stroke="rgba(22, 119, 255, 0.3)"
              strokeWidth="1"
              strokeDasharray="5 5"
            />
          );
        })}
      </svg>


      {/* 闪烁效果层 */}
      <div
        className="bg-shimmer"
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '100%',
          height: '100%',
          pointerEvents: 'none'
        }}
      />

      <div style={{
        width: '100%',
        maxWidth: 400,
        position: 'relative',
        zIndex: 1
      }}>
        {/* Logo 和标题区域 */}
        <div style={{
          textAlign: 'center',
          marginBottom: 48
        }}>
          <div style={{
            width: 80,
            height: 80,
            borderRadius: 16,
            background: '#fff',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 24px',
            boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
            padding: 12
          }}>
            <img src="/favicon.svg" alt="ELK Helper" style={{ width: '100%', height: '100%' }} />
          </div>
          <Title level={2} style={{
            marginBottom: 8,
            color: '#fff',
            fontWeight: 600
          }}>
            ELK Helper
          </Title>
          <Text style={{
            color: 'rgba(255, 255, 255, 0.85)',
            fontSize: 16
          }}>
            智能告警系统
          </Text>
        </div>

        {/* 登录表单卡片 */}
        <div style={{
          background: '#fff',
          borderRadius: 8,
          padding: 40,
          boxShadow: '0 8px 24px rgba(0,0,0,0.12)'
        }}>
          <Title level={4} style={{
            marginBottom: 32,
            textAlign: 'center',
            fontWeight: 500
          }}>
            账户登录
          </Title>

          <Form
            name="login"
            onFinish={handleSubmit}
            size="large"
            autoComplete="off"
            layout="vertical"
          >
            <Form.Item
              name="username"
              label="用户名"
              rules={[{ required: true, message: '请输入用户名' }]}
              style={{ marginBottom: 24 }}
            >
              <Input
                prefix={<UserOutlined style={{ color: '#bfbfbf' }} />}
                placeholder="请输入用户名"
                disabled={loading}
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="密码"
              rules={[{ required: true, message: '请输入密码' }]}
              style={{ marginBottom: 32 }}
            >
              <Input.Password
                prefix={<LockOutlined style={{ color: '#bfbfbf' }} />}
                placeholder="请输入密码"
                disabled={loading}
              />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                block
                size="large"
                style={{
                  height: 44,
                  fontSize: 16,
                  fontWeight: 500
                }}
              >
                登录
              </Button>
            </Form.Item>
          </Form>
        </div>

        {/* 页脚 */}
        <div style={{
          textAlign: 'center',
          marginTop: 32,
          position: 'relative',
          zIndex: 1
        }}>
          <Text style={{
            color: 'rgba(255, 255, 255, 0.65)',
            fontSize: 14
          }}>
            由 <Text strong style={{ color: '#fff' }}>系统运维部</Text> 驱动开发
          </Text>
        </div>
      </div>
    </div>
  );
}
