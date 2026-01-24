// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  App,
  Button,
  Card,
  Collapse,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Spin,
  Switch,
  Typography,
  theme,
} from 'antd';
import { LoadingOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { esConfigApi, larkConfigApi, rulesApi, Rule } from '../services/api';

const { Text, Paragraph, Title } = Typography;
const { TextArea } = Input;

export interface RuleEditDialogProps {
  open: boolean;
  ruleId?: number;
  onOpenChange: (open: boolean) => void;
}

export default function RuleEditDialog({ open, ruleId, onOpenChange }: RuleEditDialogProps) {
  const isEdit = typeof ruleId === 'number';
  const queryClient = useQueryClient();
  const { message } = App.useApp();
  const { token } = theme.useToken();
  const [form] = Form.useForm();

  const [testModalOpen, setTestModalOpen] = useState(false);
  const [testResult, setTestResult] = useState<any>(null);
  const [isTesting, setIsTesting] = useState(false);

  const { data: ruleData, isLoading: ruleLoading } = useQuery({
    queryKey: ['rule', ruleId],
    queryFn: () => rulesApi.getById(Number(ruleId)).then((res) => res.data.data),
    enabled: open && isEdit,
  });

  const { data: larkConfigs, isLoading: larkLoading } = useQuery({
    queryKey: ['lark-configs'],
    queryFn: () => larkConfigApi.getAll().then((res) => res.data.data),
    enabled: open,
  });

  const { data: esConfigs, isLoading: esLoading } = useQuery({
    queryKey: ['es-configs'],
    queryFn: () => esConfigApi.getAll().then((res) => res.data.data),
    enabled: open,
  });

  const queryTemplates = useMemo(
    () => [
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
        value: JSON.stringify(
          [
            { field: 'response_code', operator: '=', value: 500, logic: 'or' },
            { field: 'response_code', operator: '=', value: 502, logic: 'or' },
            { field: 'response_code', operator: '=', value: 503, logic: 'or' },
          ],
          null,
          2
        ),
      },
      {
        label: '慢查询',
        value: JSON.stringify(
          [
            { field: 'responsetime', operator: '>', value: 3, logic: 'and' },
            { field: 'response_code', operator: '=', value: 200, logic: 'and' },
          ],
          null,
          2
        ),
      },
      {
        label: '499错误',
        value: JSON.stringify([{ field: 'response_code', operator: '=', value: 499 }], null, 2),
      },
    ],
    []
  );

  // When opening dialog, initialize form values
  useEffect(() => {
    if (!open) return;

    if (!isEdit) {
      form.resetFields();
      form.setFieldsValue({
        enabled: true,
        interval: 60,
        queries: '[]',
      });
      return;
    }

    if (!ruleData) return;

    let queries: any[] = [];
    if ((ruleData as any).queries) {
      if (typeof (ruleData as any).queries === 'string') {
        try {
          queries = JSON.parse((ruleData as any).queries);
        } catch {
          queries = [];
        }
      } else if (Array.isArray((ruleData as any).queries)) {
        queries = (ruleData as any).queries;
      }
    }

    form.setFieldsValue({
      ...ruleData,
      queries: JSON.stringify(queries, null, 2),
      es_config_id: ruleData.es_config_id || undefined,
      lark_config_id: ruleData.lark_config_id || undefined,
    });
  }, [open, isEdit, ruleData, form]);

  const createMutation = useMutation({
    mutationFn: (rule: Partial<Rule>) => rulesApi.create(rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      message.success('创建成功');
      onOpenChange(false);
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '创建失败');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, rule }: { id: number; rule: Partial<Rule> }) => rulesApi.update(id, rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      message.success('更新成功');
      onOpenChange(false);
    },
    onError: (error: any) => {
      message.error(error?.response?.data?.error || '更新失败');
    },
  });

  const parseQueries = (raw: any) => {
    let queries = raw;
    if (typeof queries === 'string') {
      try {
        queries = JSON.parse(queries);
      } catch (e: any) {
        throw new Error(`查询条件 JSON 格式错误: ${e.message}`);
      }
    }
    if (!Array.isArray(queries)) {
      throw new Error('查询条件必须是 JSON 数组格式');
    }
    for (let i = 0; i < queries.length; i++) {
      const q = queries[i];
      if (!q.field) {
        throw new Error(`查询条件第 ${i + 1} 项缺少 "field" 字段`);
      }
      if (!q.operator && !q.op) {
        throw new Error(`查询条件第 ${i + 1} 项缺少 "operator" 字段`);
      }
    }
    return queries;
  };

  const handleSubmit = async (values: any) => {
    try {
      const queries = parseQueries(values.queries);
      const submitData = { ...values, queries };

      if (isEdit) {
        updateMutation.mutate({ id: Number(ruleId), rule: submitData });
      } else {
        createMutation.mutate(submitData);
      }
    } catch (e: any) {
      message.error(e?.message || '保存失败');
    }
  };

  const formatQueries = () => {
    const raw = form.getFieldValue('queries');
    try {
      const parsed = JSON.parse(raw);
      form.setFieldsValue({ queries: JSON.stringify(parsed, null, 2) });
      message.success('格式化成功');
    } catch (e: any) {
      message.error('JSON 格式错误: ' + e.message);
    }
  };

  const handleTest = async () => {
    try {
      const values = await form.validateFields();
      const queries = parseQueries(values.queries);

      setIsTesting(true);
      const response = await rulesApi.test({ ...values, queries });
      setTestResult(response.data);
      setTestModalOpen(true);

      if (response.data.success) {
        message.success(`测试成功，找到 ${response.data.data.count} 条匹配日志`);
      } else {
        message.error(response.data.error || '测试失败');
      }
    } catch (error: any) {
      message.error(error?.message || error?.response?.data?.error || '测试失败');
    } finally {
      setIsTesting(false);
    }
  };

  const isBusy = ruleLoading || larkLoading || esLoading;

  return (
    <>
      <Modal
        open={open}
        onCancel={() => onOpenChange(false)}
        title={isEdit ? '编辑规则' : '新建规则'}
        width={900}
        destroyOnClose
        footer={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 12 }}>
            <Button
              icon={isTesting ? <LoadingOutlined /> : <PlayCircleOutlined />}
              onClick={handleTest}
              loading={isTesting}
              disabled={isBusy || createMutation.isPending || updateMutation.isPending}
            >
              测试规则
            </Button>
            <Space>
              <Button onClick={() => onOpenChange(false)} disabled={createMutation.isPending || updateMutation.isPending}>
                取消
              </Button>
              <Button
                type="primary"
                onClick={() => form.submit()}
                loading={createMutation.isPending || updateMutation.isPending}
              >
                保存
              </Button>
            </Space>
          </div>
        }
        styles={{
          body: { paddingTop: 12, maxHeight: '70vh', overflow: 'auto' },
        }}
      >
        {isBusy ? (
          <div style={{ textAlign: 'center', padding: 48 }}>
            <Spin size="large" tip="加载中..." />
          </div>
        ) : (
          <Card bordered={false} style={{ background: 'transparent' }} styles={{ body: { padding: 0 } }}>
            <Form form={form} layout="vertical" initialValues={{ enabled: true, interval: 60 }} onFinish={handleSubmit}>
              <Form.Item name="name" label="规则名称" rules={[{ required: true, message: '请输入规则名称' }]}>
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

              <Form.Item name="enabled" label="启用规则" valuePropName="checked" extra="禁用后此规则将不会执行查询">
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

              <Form.Item name="es_config_id" label="ES 数据源配置" extra="不选择则使用默认配置">
                <Select
                  allowClear
                  placeholder="使用默认数据源"
                  options={esConfigs
                    ?.filter((c) => c.enabled)
                    .map((config) => ({ value: config.id, label: `${config.name}${config.is_default ? ' (默认)' : ''}` }))}
                />
              </Form.Item>

              <Form.Item name="lark_config_id" label="告警配置" extra="选择已配置的告警配置，或选择空直接输入 URL">
                <Select
                  allowClear
                  placeholder="不使用配置（直接输入 URL）"
                  options={larkConfigs
                    ?.filter((c) => c.enabled)
                    .map((config) => ({ value: config.id, label: `${config.name}${config.is_default ? ' (默认)' : ''}` }))}
                  onChange={(value) => {
                    if (value) {
                      form.setFieldsValue({ lark_webhook: '' });
                    }
                  }}
                />
              </Form.Item>

              <Form.Item noStyle shouldUpdate={(prev, cur) => prev.lark_config_id !== cur.lark_config_id}>
                {({ getFieldValue }) => {
                  const hasConfig = getFieldValue('lark_config_id');
                  return (
                    <Form.Item
                      name="lark_webhook"
                      label={`Webhook URL ${hasConfig ? '(可选)' : ''}`}
                      rules={[
                        {
                          required: !hasConfig,
                          message: '请选择告警配置或输入 Webhook URL',
                        },
                      ]}
                    >
                      <Input placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..." disabled={!!hasConfig} />
                    </Form.Item>
                  );
                }}
              </Form.Item>

              <Form.Item
                label={
                  <Space>
                    <span>查询条件 (JSON)</span>
                    <Button size="small" onClick={formatQueries}>
                      格式化
                    </Button>
                  </Space>
                }
                name="queries"
                rules={[{ required: true, message: '请配置查询条件' }]}
              >
                <TextArea rows={12} placeholder="请输入 JSON 格式的查询条件..." style={{ fontFamily: 'monospace', fontSize: 13 }} />
              </Form.Item>

              <Space wrap style={{ marginBottom: 16 }}>
                {queryTemplates.map((tpl) => (
                  <Button key={tpl.label} size="small" onClick={() => form.setFieldsValue({ queries: tpl.value })}>
                    {tpl.label}
                  </Button>
                ))}
              </Space>

              <Collapse
                size="small"
                items={[
                  {
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
                  },
                ]}
                style={{ marginBottom: 8 }}
              />

              <div className="app-muted" style={{ fontSize: 12 }}>
                提示：保存后规则将立即生效（若启用）。
              </div>
            </Form>
          </Card>
        )}
      </Modal>

      <Modal
        title="测试结果"
        open={testModalOpen}
        onCancel={() => setTestModalOpen(false)}
        footer={<Button onClick={() => setTestModalOpen(false)}>关闭</Button>}
        width={860}
      >
        {testResult && (
          <div>
            {testResult.success ? (
              <>
                <div style={{ display: 'flex', gap: 48, marginBottom: 16, flexWrap: 'wrap' }}>
                  <div>
                    <Text type="secondary">匹配数量</Text>
                    <Title level={2} style={{ margin: 0 }}>
                      {testResult.data.count}
                    </Title>
                  </div>
                  <div>
                    <Text type="secondary">时间范围</Text>
                    <div style={{ color: token.colorText }}>
                      {testResult.data.time_range.from}
                      <br />~ {testResult.data.time_range.to}
                    </div>
                  </div>
                </div>
                {testResult.data.logs?.length > 0 && (
                  <div>
                    <Text type="secondary">日志示例（最多显示 10 条）</Text>
                    <pre className="app-code-block" style={{ maxHeight: 400 }}>
                      {JSON.stringify(testResult.data.logs.slice(0, 10), null, 2)}
                    </pre>
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
    </>
  );
}

