// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Form,
  Input,
  Button,
  Switch,
  Select,
  Card,
  Space,
  Typography,
  InputNumber,
  Modal,
  Spin,
  Collapse,
  App,
} from 'antd';
import { PlayCircleOutlined, LoadingOutlined } from '@ant-design/icons';
import { rulesApi, Rule, larkConfigApi, esConfigApi } from '../services/api';
import PageHeader from '../components/PageHeader';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

function escapeRegex(s: string) {
  return s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

function extractHitKeywords(queries: any[]): string[] {
  const values: string[] = [];
  for (const q of queries || []) {
    const op = q?.operator || q?.op;
    if (op !== 'contains') continue;
    if (typeof q?.value !== 'string') continue;
    const v = q.value.trim();
    if (!v) continue;
    values.push(v);
  }
  return Array.from(new Set(values)).sort((a, b) => b.length - a.length);
}

function highlightText(text: string, keywords: string[]) {
  if (!text || !keywords || keywords.length === 0) return text;
  const pattern = new RegExp(`(${keywords.map(escapeRegex).join('|')})`, 'gi');
  const parts = text.split(pattern);
  if (parts.length <= 1) return text;
  const keywordSet = new Set(keywords.map((k) => k.toLowerCase()));
  return parts.map((part, idx) => {
    const isHit = keywordSet.has(part.toLowerCase());
    return isHit ? (
      <span key={idx} className="app-highlight-hit">
        {part}
      </span>
    ) : (
      <span key={idx}>{part}</span>
    );
  });
}

export default function RuleEditPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const { message } = App.useApp();
  const [form] = Form.useForm();
  const isEdit = !!id;
  const [testModalOpen, setTestModalOpen] = useState(false);
  const [testResult, setTestResult] = useState<any>(null);
  const [isTesting, setIsTesting] = useState(false);
  const [testHitKeywords, setTestHitKeywords] = useState<string[]>([]);

  const { data: ruleData, isLoading } = useQuery({
    queryKey: ['rule', id],
    queryFn: () => rulesApi.getById(Number(id!)).then(res => res.data.data),
    enabled: isEdit,
  });

  const { data: larkConfigs } = useQuery({
    queryKey: ['lark-configs'],
    queryFn: () => larkConfigApi.getAll().then(res => res.data.data),
  });

  const { data: esConfigs } = useQuery({
    queryKey: ['es-configs'],
    queryFn: () => esConfigApi.getAll().then(res => res.data.data),
  });

  useEffect(() => {
    if (isEdit && ruleData) {
      let queries: any[] = [];
      if (ruleData.queries) {
        if (typeof ruleData.queries === 'string') {
          try {
            queries = JSON.parse(ruleData.queries);
          } catch (e) {
            queries = [];
          }
        } else if (Array.isArray(ruleData.queries)) {
          queries = ruleData.queries;
        }
      }

      form.setFieldsValue({
        ...ruleData,
        queries: JSON.stringify(queries, null, 2),
        es_config_id: ruleData.es_config_id || undefined,
        lark_config_id: ruleData.lark_config_id || undefined,
      });
    }
  }, [ruleData, isEdit, form]);

  const createMutation = useMutation({
    mutationFn: (rule: Partial<Rule>) => rulesApi.create(rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      message.success('创建成功');
      navigate('/rules');
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '创建失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, rule }: { id: number; rule: Partial<Rule> }) =>
      rulesApi.update(id, rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      message.success('更新成功');
      navigate('/rules');
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '更新失败');
    },
  });

  const handleTest = async () => {
    try {
      const values = await form.validateFields();

      let queries = values.queries;
      if (typeof queries === 'string') {
        try {
          queries = JSON.parse(queries);
        } catch (e) {
          message.error('查询条件 JSON 格式错误');
          return;
        }
      }

      setIsTesting(true);
      const response = await rulesApi.test({ ...values, queries });
      setTestResult(response.data);
      setTestHitKeywords(extractHitKeywords(queries));
      setTestModalOpen(true);

      if (response.data.success) {
        message.success(`测试成功，找到 ${response.data.data.count} 条匹配日志`);
      } else {
        message.error(response.data.error || '测试失败');
      }
    } catch (error: any) {
      message.error(error?.response?.data?.error || '测试失败');
    } finally {
      setIsTesting(false);
    }
  };

  const handleSubmit = async (values: any) => {
    let queries = values.queries;
    if (typeof queries === 'string') {
      try {
        queries = JSON.parse(queries);
        if (!Array.isArray(queries)) {
          message.error('查询条件必须是 JSON 数组格式');
          return;
        }
      } catch (e: any) {
        message.error(`查询条件 JSON 格式错误: ${e.message}`);
        return;
      }
    }

    // Validate each query condition
    for (let i = 0; i < queries.length; i++) {
      const q = queries[i];
      if (!q.field) {
        message.error(`查询条件第 ${i + 1} 项缺少 "field" 字段`);
        return;
      }
      if (!q.operator && !q.op) {
        message.error(`查询条件第 ${i + 1} 项缺少 "operator" 字段`);
        return;
      }
    }

    const submitData = { ...values, queries };

    if (isEdit) {
      updateMutation.mutate({ id: Number(id!), rule: submitData });
    } else {
      createMutation.mutate(submitData);
    }
  };

  const formatQueries = () => {
    const queries = form.getFieldValue('queries');
    try {
      const parsed = JSON.parse(queries);
      form.setFieldsValue({ queries: JSON.stringify(parsed, null, 2) });
      message.success('格式化成功');
    } catch (e: any) {
      message.error('JSON 格式错误: ' + e.message);
    }
  };

  const setTemplate = (template: string) => {
    form.setFieldsValue({ queries: template });
  };

  if (isEdit && isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin size="large" tip="加载中..." />
      </div>
    );
  }

  const queryTemplates = [
    {
      label: '非200响应',
      value: JSON.stringify([{ field: 'response_code', operator: '!=', value: 200 }], null, 2),
    },
    {
      label: '4xx/5xx错误',
      value: JSON.stringify([{ field: 'response_code', operator: '>=', value: 400 }], null, 2),
    },
    {
      label: '5xx错误',
      value: JSON.stringify([
        { field: 'response_code', operator: '=', value: 500, logic: 'or' },
        { field: 'response_code', operator: '=', value: 502, logic: 'or' },
        { field: 'response_code', operator: '=', value: 503, logic: 'or' },
      ], null, 2),
    },
    {
      label: '慢查询',
      value: JSON.stringify([
        { field: 'responsetime', operator: '>', value: 3, logic: 'and' },
        { field: 'response_code', operator: '=', value: 200, logic: 'and' },
      ], null, 2),
    },
    {
      label: '499错误',
      value: JSON.stringify([{ field: 'response_code', operator: '=', value: 499 }], null, 2),
    },
  ];

  return (
    <div>
      <PageHeader
        title={isEdit ? '编辑规则' : '新建规则'}
        description="配置索引、查询条件与告警渠道。"
        extra={
          <Button icon={isTesting ? <LoadingOutlined /> : <PlayCircleOutlined />} onClick={handleTest} loading={isTesting}>
            测试规则
          </Button>
        }
      />

      <Card style={{ maxWidth: 800 }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{ enabled: true, interval: 60 }}
          onFinish={handleSubmit}
        >
          <Form.Item
            name="name"
            label="规则名称"
            rules={[{ required: true, message: '请输入规则名称' }]}
          >
            <Input placeholder="例如：Nginx 错误日志告警" />
          </Form.Item>

          <Form.Item
            name="index_pattern"
            label="索引模式"
            rules={[{ required: true, message: '请输入索引模式' }]}
            extra="支持通配符，如：prod-nginx-access-*-*-*"
          >
            <Input placeholder="例如：prod-nginx-access-*-*-*" />
          </Form.Item>

          <Form.Item name="description" label="描述">
            <TextArea rows={2} placeholder="规则描述（可选）" />
          </Form.Item>

          <Form.Item
            name="enabled"
            label="启用规则"
            valuePropName="checked"
            extra="禁用后此规则将不会执行查询"
          >
            <Switch />
          </Form.Item>

          <Form.Item
            name="interval"
            label="查询间隔（秒）"
            rules={[{ required: true, message: '请输入查询间隔' }]}
            extra="规则执行查询的时间间隔，建议 60 秒以上"
          >
            <InputNumber min={10} max={3600} style={{ width: '100%' }} />
          </Form.Item>

          <Form.Item
            name="es_config_id"
            label="ES 数据源配置"
            extra="选择用于查询的 Elasticsearch 数据源配置，如不选择则使用默认配置"
          >
            <Select
              allowClear
              placeholder="使用默认数据源"
              options={esConfigs?.filter(c => c.enabled).map(config => ({
                value: config.id,
                label: `${config.name}${config.is_default ? ' (默认)' : ''}`,
              }))}
            />
          </Form.Item>

          <Form.Item
            name="lark_config_id"
            label="告警配置"
            extra="选择已配置的告警配置，或选择空直接输入 URL"
          >
            <Select
              allowClear
              placeholder="不使用配置（直接输入 URL）"
              options={larkConfigs?.filter(c => c.enabled).map(config => ({
                value: config.id,
                label: `${config.name}${config.is_default ? ' (默认)' : ''}`,
              }))}
              onChange={(value) => {
                if (value) {
                  form.setFieldsValue({ lark_webhook: '' });
                }
              }}
            />
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prev, cur) => prev.lark_config_id !== cur.lark_config_id}
          >
            {({ getFieldValue }) => {
              const hasConfig = getFieldValue('lark_config_id');
              return (
                <Form.Item
                  name="lark_webhook"
                  label={`Webhook URL ${hasConfig ? '(可选)' : ''}`}
                  rules={[{
                    required: !hasConfig,
                    message: '请选择告警配置或输入 Webhook URL',
                  }]}
                >
                  <Input
                    placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..."
                    disabled={!!hasConfig}
                  />
                </Form.Item>
              );
            }}
          </Form.Item>

          <Form.Item
            label={
              <Space>
                <span>查询条件 (JSON)</span>
                <Button size="small" onClick={handleTest} loading={isTesting}>
                  测试
                </Button>
                <Button size="small" onClick={formatQueries}>
                  格式化
                </Button>
              </Space>
            }
            name="queries"
            rules={[{ required: true, message: '请配置查询条件' }]}
          >
            <TextArea
              rows={12}
              placeholder="请输入 JSON 格式的查询条件..."
              style={{ fontFamily: 'monospace', fontSize: 13 }}
            />
          </Form.Item>

          <Space wrap style={{ marginBottom: 16 }}>
            {queryTemplates.map((tpl) => (
              <Button key={tpl.label} size="small" onClick={() => setTemplate(tpl.value)}>
                {tpl.label}
              </Button>
            ))}
          </Space>

          <Collapse
            size="small"
            items={[{
              key: '1',
              label: '支持的操作符 (operator)',
              children: (
                <div style={{ fontSize: 12 }}>
                  <Paragraph>
                    <Text strong>比较操作符：</Text>
                    <br />
                    <Text code>=</Text> / <Text code>==</Text> / <Text code>equals</Text> - 等于
                    <br />
                    <Text code>!=</Text> / <Text code>not_equals</Text> - 不等于
                    <br />
                    <Text code>&gt;</Text> / <Text code>gt</Text> - 大于
                    <br />
                    <Text code>&gt;=</Text> / <Text code>gte</Text> - 大于等于
                    <br />
                    <Text code>&lt;</Text> / <Text code>lt</Text> - 小于
                    <br />
                    <Text code>&lt;=</Text> / <Text code>lte</Text> - 小于等于
                  </Paragraph>
                  <Paragraph>
                    <Text strong>文本/存在性：</Text>
                    <br />
                    <Text code>contains</Text> - 包含（文本匹配）
                    <br />
                    <Text code>not_contains</Text> - 不包含
                    <br />
                    <Text code>exists</Text> - 字段存在
                  </Paragraph>
                </div>
              ),
            }]}
            style={{ marginBottom: 24 }}
          />

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={createMutation.isPending || updateMutation.isPending}
              >
                保存
              </Button>
              <Button onClick={() => navigate('/rules')}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {/* Test Result Modal */}
      <Modal
        title="测试结果"
        open={testModalOpen}
        onCancel={() => setTestModalOpen(false)}
        footer={<Button onClick={() => setTestModalOpen(false)}>关闭</Button>}
        width={800}
      >
        {testResult && (
          <div>
            {testResult.success ? (
              <>
                <div style={{ display: 'flex', gap: 48, marginBottom: 16 }}>
                  <div>
                    <Text type="secondary">匹配数量</Text>
                    <Title level={2} style={{ margin: 0 }}>{testResult.data.count}</Title>
                  </div>
                  <div>
                    <Text type="secondary">时间范围</Text>
                    <div>
                      {testResult.data.time_range.from}
                      <br />
                      ~ {testResult.data.time_range.to}
                    </div>
                  </div>
                </div>
                {testResult.data.logs?.length > 0 && (
                  <div>
                    <Text type="secondary">日志示例（最多显示 10 条）</Text>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 12, marginTop: 8 }}>
                      {testResult.data.logs.slice(0, 10).map((log: any, idx: number) => {
                        const msg = typeof log?.message === 'string' ? log.message : '';
                        return (
                          <div key={idx} className="app-surface-muted" style={{ padding: 12 }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', gap: 12, marginBottom: 6 }}>
                              <Text type="secondary">#{idx + 1}</Text>
                              {log?.['@timestamp'] ? <Text type="secondary">{String(log['@timestamp'])}</Text> : null}
                            </div>
                            {msg ? (
                              <>
                                <div style={{ fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace', fontSize: 12, lineHeight: 1.6 }}>
                                  {highlightText(msg, testHitKeywords)}
                                </div>
                                <details style={{ marginTop: 8 }}>
                                  <summary style={{ cursor: 'pointer', fontSize: 12 }}>查看原始 JSON</summary>
                                  <pre className="app-code-block" style={{ marginTop: 8, maxHeight: 260 }}>
                                    {JSON.stringify(log, null, 2)}
                                  </pre>
                                </details>
                              </>
                            ) : (
                              <pre className="app-code-block" style={{ maxHeight: 260 }}>
                                {JSON.stringify(log, null, 2)}
                              </pre>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  </div>
                )}
              </>
            ) : (
              <div style={{ color: '#ff4d4f' }}>
                <Text strong>错误信息：</Text>
                <Paragraph>{testResult.error}</Paragraph>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
}
