// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import axios from 'axios';

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add request interceptor to include auth token
api.interceptors.request.use(
  (config: any) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  }
);

// Add response interceptor to handle 401 errors
api.interceptors.response.use(
  (response: any) => response,
  (error: any) => {
    if (error.response?.status === 401) {
      // Clear token and redirect to login
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export interface Rule {
  id: number;
  name: string;
  index_pattern: string;
  queries: QueryCondition[];
  enabled: boolean;
  interval: number;
  es_config_id?: number; // ES 数据源配置 ID
  es_config?: ESConfig; // ES 数据源配置关联
  lark_webhook: string; // 保留用于向后兼容
  lark_config_id?: number; // Lark 配置 ID
  lark_config?: LarkConfig; // Lark 配置关联
  description?: string;
  last_run_time?: string;
  run_count: number;
  alert_count: number;
  created_at: string;
  updated_at: string;
}

export interface QueryCondition {
  field: string;
  type?: string;
  value: any;
  op?: string;
  logic?: string;
}

export interface Alert {
  id: number;
  rule_id: number;
  rule?: Rule;
  index_name: string;
  log_count: number;
  logs: any[];
  time_range: string;
  status: 'sent' | 'failed';
  error_msg?: string;
  created_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_page: number;
  };
}

// Rules API
export const rulesApi = {
  getAll: () => api.get<{ data: Rule[] }>('/rules'),
  getById: (id: number) => api.get<{ data: Rule }>(`/rules/${id}`),
  create: (rule: Partial<Rule>) => api.post<{ data: Rule }>('/rules', rule),
  update: (id: number, rule: Partial<Rule>) => api.put<{ data: Rule }>(`/rules/${id}`, rule),
  delete: (id: number) => api.delete(`/rules/${id}`),
  toggleEnabled: (id: number) => api.post<{ data: Rule }>(`/rules/${id}/toggle`),
  clone: (id: number, name: string) => api.post<{ data: Rule }>(`/rules/${id}/clone`, { name }),
  test: (rule: Partial<Rule>) => api.post<{ success: boolean; data: { count: number; logs: any[]; time_range: { from: string; to: string } }; error?: string }>('/rules/test', rule),
  batchDelete: (ids: number[]) => api.post<{ success_count: number; error_count: number; errors: string[] }>('/rules/batch-delete', { ids }),
  export: () => api.get<{ version: string; exported_at: string; rules: Rule[] }>('/rules/export'),
  import: (rules: Rule[]) => api.post<{ 
    created_count: number; 
    updated_count: number; 
    skipped_count: number; 
    error_count: number; 
    errors: string[];
    details: string[];
  }>('/rules/import', { rules }),
};

export interface RuleAlertStats {
  rule_id: number;
  rule_name: string;
  total: number;
  sent: number;
  failed: number;
  last_alert?: string;
}

export interface TimeSeriesDataPoint {
  time: string;
  value: number;
}

export interface RuleTimeSeriesStats {
  rule_id: number;
  rule_name: string;
  total: number;
  data_points: TimeSeriesDataPoint[];
}

// Alerts API
export const alertsApi = {
  getAll: (page = 1, pageSize = 20) =>
    api.get<PaginatedResponse<Alert>>('/alerts', {
      params: { page, page_size: pageSize },
    }),
  getById: (id: number) => api.get<{ data: Alert }>(`/alerts/${id}`),
  getStats: (duration = '24h') =>
    api.get<{ data: { total: number; sent: number; failed: number } }>('/alerts/stats', {
      params: { duration },
    }),
  getRuleStats: (duration = '24h') =>
    api.get<{ data: RuleAlertStats[] }>('/alerts/rule-stats', {
      params: { duration },
    }),
  getRuleTimeSeries: (duration = '24h', interval = 60) =>
    api.get<{ data: RuleTimeSeriesStats[] }>('/alerts/rule-timeseries', {
      params: { duration, interval },
    }),
  delete: (id: number) => api.delete(`/alerts/${id}`),
  batchDelete: (ids: number[]) => api.post<{ success: boolean; deleted_count: number }>('/alerts/batch-delete', { ids }),
};

// Status API
export const statusApi = {
  getStatus: () =>
    api.get<{
      data: {
        rules: { total: number; enabled: number };
        elasticsearch: {
          status: string;
          total: number;
          success_count: number;
          failed_count: number;
          unknown_count: number;
          details: Array<{
            id: number;
            name: string;
            url: string;
            status: string;
          }>;
        };
        alerts_24h: { total: number; sent: number; failed: number };
      };
    }>('/status'),
};

export interface ESConfig {
  id: number;
  name: string;
  url: string;
  username?: string;
  password?: string;
  use_ssl?: boolean;
  skip_verify?: boolean;
  ca_certificate?: string;
  is_default: boolean;
  description?: string;
  enabled: boolean;
  last_test_at?: string;
  test_status: 'unknown' | 'success' | 'failed';
  test_error?: string;
  created_at: string;
  updated_at: string;
}

// ES Config API
export const esConfigApi = {
  getAll: () => api.get<{ data: ESConfig[] }>('/es-configs'),
  getById: (id: number) => api.get<{ data: ESConfig }>(`/es-configs/${id}`),
  create: (config: Partial<ESConfig>) => api.post<{ data: ESConfig }>('/es-configs', config),
  update: (id: number, config: Partial<ESConfig>) => api.put<{ data: ESConfig }>(`/es-configs/${id}`, config),
  delete: (id: number) => api.delete(`/es-configs/${id}`),
  test: (id: number) => api.post<{ success: boolean; error?: string; message?: string }>(`/es-configs/${id}/test`),
  setDefault: (id: number) => api.post<{ success: boolean }>(`/es-configs/${id}/set-default`),
};

export interface LarkConfig {
  id: number;
  name: string;
  webhook_url: string;
  is_default: boolean;
  description?: string;
  enabled: boolean;
  last_test_at?: string;
  test_status: 'unknown' | 'success' | 'failed';
  test_error?: string;
  created_at: string;
  updated_at: string;
}

// Lark Config API
export const larkConfigApi = {
  getAll: () => api.get<{ data: LarkConfig[] }>('/lark-configs'),
  getById: (id: number) => api.get<{ data: LarkConfig }>(`/lark-configs/${id}`),
  create: (config: Partial<LarkConfig>) => api.post<{ data: LarkConfig }>('/lark-configs', config),
  update: (id: number, config: Partial<LarkConfig>) => api.put<{ data: LarkConfig }>(`/lark-configs/${id}`, config),
  delete: (id: number) => api.delete(`/lark-configs/${id}`),
  test: (id: number) => api.post<{ success: boolean; error?: string; message?: string }>(`/lark-configs/${id}/test`),
  setDefault: (id: number) => api.post<{ success: boolean }>(`/lark-configs/${id}/set-default`),
};

// User interface
export interface User {
  id: number;
  username: string;
  email?: string;
  role: 'admin' | 'user';
  enabled: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

// Auth API
export const authApi = {
  login: (username: string, password: string) =>
    api.post<{ token: string; user: User; expires_at: string }>('/auth/login', {
      username,
      password,
    }),
  logout: () => api.post<{ message: string }>('/auth/logout'),
  getCurrentUser: () => api.get<{ data: User }>('/auth/me'),
  updatePassword: (oldPassword: string, newPassword: string) =>
    api.put<{ message: string }>('/auth/password', {
      old_password: oldPassword,
      new_password: newPassword,
    }),
};

// Cleanup Config interface
export interface CleanupConfig {
  enabled: boolean;
  hour: number;
  minute: number;
  retention_days: number;
  last_execution_status?: 'success' | 'failed' | 'never';
  last_execution_time?: string;
  last_execution_result?: string;
}

// System Config API
export const systemConfigApi = {
  getCleanupConfig: () => api.get<{ data: CleanupConfig }>('/system-config/cleanup'),
  updateCleanupConfig: (config: CleanupConfig) =>
    api.put<{ data: CleanupConfig }>('/system-config/cleanup', config),
  manualCleanup: () =>
    api.post<{ message: string; deleted_count: number; retention_days: number }>('/system-config/cleanup/manual'),
};

export default api;

