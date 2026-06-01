// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect, useState } from 'react';
import { Button, Result, Spin } from 'antd';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export default function SsoCallbackPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { loginWithToken } = useAuth();
  const [errorMsg, setErrorMsg] = useState('');

  useEffect(() => {
    const err = params.get('error');
    if (err) {
      setErrorMsg(err);
      return;
    }
    const token = params.get('token');
    const username = params.get('username');
    const role = (params.get('role') || 'user') as 'admin' | 'user';
    if (!token || !username) {
      setErrorMsg('SSO callback is missing token or username');
      return;
    }
    void (async () => {
      try {
        await loginWithToken(token, username, role);
        navigate('/', { replace: true });
      } catch {
        setErrorMsg('Failed to complete SSO login');
      }
    })();
  }, [params, loginWithToken, navigate]);

  if (errorMsg) {
    return (
      <Result
        status="error"
        title="SSO login failed"
        subTitle={errorMsg}
        extra={
          <Button type="primary" onClick={() => navigate('/login', { replace: true })}>
            Back to login
          </Button>
        }
      />
    );
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 16,
      }}
    >
      <Spin size="large" />
      <span style={{ color: '#64748B' }}>Completing SSO login...</span>
    </div>
  );
}
