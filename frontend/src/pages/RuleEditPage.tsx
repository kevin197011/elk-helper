// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { rulesApi, Rule, larkConfigApi, esConfigApi } from '../services/api';
import { useToast } from '@/contexts/ToastContext';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Play, Loader2 } from 'lucide-react';

export default function RuleEditPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const toast = useToast();
  const isEdit = !!id;
  const [testDialogOpen, setTestDialogOpen] = useState(false);
  const [testResult, setTestResult] = useState<any>(null);
  const [isTesting, setIsTesting] = useState(false);

  const { data: ruleData, isLoading } = useQuery({
    queryKey: ['rule', id],
    queryFn: () => rulesApi.getById(Number(id!)).then(res => res.data.data),
    enabled: isEdit,
  });

  const { data: larkConfigs, isLoading: larkConfigsLoading } = useQuery({
    queryKey: ['lark-configs'],
    queryFn: () => larkConfigApi.getAll().then(res => res.data.data),
    retry: 1,
    refetchOnWindowFocus: false,
    staleTime: 60000, // Cache for 1 minute
  });

  const { data: esConfigs, isLoading: esConfigsLoading } = useQuery({
    queryKey: ['es-configs'],
    queryFn: () => esConfigApi.getAll().then(res => res.data.data),
    retry: 1,
    refetchOnWindowFocus: false,
    staleTime: 60000, // Cache for 1 minute
  });

  const form = useForm<Partial<Rule>>({
    defaultValues: {
      enabled: true,
      interval: 60,
      queries: [],
      es_config_id: undefined,
      lark_webhook: '',
      lark_config_id: undefined,
    },
  });

  useEffect(() => {
    if (isEdit && ruleData) {
      // Parse queries field - handle different formats
      let queries: any[] = [];

      if (ruleData.queries !== null && ruleData.queries !== undefined) {
        if (typeof ruleData.queries === 'string') {
          try {
            const parsed = JSON.parse(ruleData.queries);
            if (Array.isArray(parsed)) {
              queries = parsed;
            }
          } catch (e) {
            // Failed to parse queries, use empty array
            queries = [];
          }
        } else if (Array.isArray(ruleData.queries)) {
          queries = ruleData.queries;
        }
      }

      form.reset({
        ...ruleData,
        queries: queries,
        es_config_id: ruleData.es_config_id || undefined,
        lark_config_id: ruleData.lark_config_id || undefined,
      });
    }
  }, [ruleData, isEdit, form]);

  const createMutation = useMutation({
    mutationFn: (rule: Partial<Rule>) => rulesApi.create(rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      toast.success('创建成功');
      navigate('/rules');
    },
    onError: (error: any) => {
      toast.error('创建失败', error?.response?.data?.error || '未知错误');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, rule }: { id: number; rule: Partial<Rule> }) =>
      rulesApi.update(id, rule),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      toast.success('更新成功');
      navigate('/rules');
    },
    onError: (error: any) => {
      toast.error('更新失败', error?.response?.data?.error || '未知错误');
    },
  });

  const testMutation = useMutation({
    mutationFn: (rule: Partial<Rule>) => rulesApi.test(rule),
    onSuccess: (data) => {
      setTestResult(data.data);
      setTestDialogOpen(true);
      if (data.data.success) {
        toast.success(`测试成功，找到 ${data.data.data.count} 条匹配日志`);
      } else {
        toast.error('测试失败', data.data.error);
      }
    },
    onError: (error: any) => {
      toast.error('测试失败', error?.response?.data?.error || '未知错误');
      setIsTesting(false);
    },
  });

  const handleTest = async () => {
    const values = form.getValues();

    // Validate required fields
    if (!values.index_pattern || !values.queries) {
      toast.error('请先填写索引模式和查询条件');
      return;
    }

    // Ensure queries is an array
    let queries = values.queries;
    if (typeof queries === 'string') {
      try {
        queries = JSON.parse(queries);
      } catch (e) {
        toast.error('查询条件 JSON 格式错误');
        return;
      }
    }

    setIsTesting(true);
    testMutation.mutate({
      ...values,
      queries,
    });
    setIsTesting(false);
  };

  const onSubmit = (values: Partial<Rule>) => {
    // Ensure queries is an array
    if (typeof values.queries === 'string') {
      try {
        const parsed = JSON.parse(values.queries);
        if (!Array.isArray(parsed)) {
          toast.error('查询条件必须是 JSON 数组格式。示例：[{"field": "response_code", "operator": "!=", "value": 200}]');
          return;
        }
        values.queries = parsed;
      } catch (e: any) {
        const errorMsg = e.message || 'JSON 格式错误';
        toast.error(
          `查询条件 JSON 格式错误: ${errorMsg}。正确格式示例：\n[{"field": "response_code", "operator": "!=", "value": 200}]`
        );
        return;
      }
    }

    // Validate queries format
    if (!Array.isArray(values.queries)) {
      toast.error('查询条件必须是数组格式');
      return;
    }

    // Validate each query condition
    for (let i = 0; i < values.queries.length; i++) {
      const q = values.queries[i] as any;
      if (!q.field) {
        toast.error(`查询条件第 ${i + 1} 项缺少 "field" 字段`);
        return;
      }
      if (!q.operator && !q.op) {
        toast.error(`查询条件第 ${i + 1} 项缺少操作符 "operator" 字段。支持的操作符：=, ==, !=, >, >=, <, <=, contains, not_contains, exists`);
        return;
      }
      if (q.value === undefined || q.value === null) {
        toast.error(`查询条件第 ${i + 1} 项缺少 "value" 字段`);
        return;
      }
    }

    if (isEdit) {
      updateMutation.mutate({ id: Number(id!), rule: values });
    } else {
      createMutation.mutate(values);
    }
  };

  // Show loading skeleton while loading rule data or configs
  if (isEdit && (isLoading || esConfigsLoading || larkConfigsLoading)) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-64" />
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-3xl font-bold tracking-tight">
          {isEdit ? '编辑规则' : '新建规则'}
        </h2>
        {!isEdit && (
          <Button
            variant="outline"
            onClick={handleTest}
            disabled={isTesting}
          >
            {isTesting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                测试中...
              </>
            ) : (
              <>
                <Play className="mr-2 h-4 w-4" />
                测试规则
              </>
            )}
          </Button>
        )}
      </div>

      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle>规则配置</CardTitle>
          <CardDescription>
            配置告警规则的查询条件和通知方式
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="name"
                rules={{ required: '请输入规则名称' }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>规则名称</FormLabel>
                    <FormControl>
                      <Input placeholder="例如：Nginx 错误日志告警" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="index_pattern"
                rules={{ required: '请输入索引模式' }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>索引模式</FormLabel>
                    <FormControl>
                      <Input placeholder="例如：prod-nginx-access-*-*-*" {...field} />
                    </FormControl>
                    <FormDescription>
                      支持通配符，如：prod-nginx-access-*-*-*
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>描述</FormLabel>
                    <FormControl>
                      <Textarea
                        rows={3}
                        placeholder="规则描述（可选）"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="enabled"
                render={({ field }) => (
                  <FormItem className="flex items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel>启用规则</FormLabel>
                      <FormDescription>
                        禁用后此规则将不会执行查询
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="interval"
                rules={{ required: '请输入查询间隔' }}
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>查询间隔（秒）</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        min={10}
                        max={3600}
                        {...field}
                        onChange={(e) => field.onChange(parseInt(e.target.value))}
                      />
                    </FormControl>
                    <FormDescription>
                      规则执行查询的时间间隔，建议 60 秒以上
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="es_config_id"
                render={({ field }) => {
                  // 获取当前选中的配置名称
                  const selectedConfig = esConfigs?.find((c) => c?.id === field.value);

                  return (
                    <FormItem>
                      <FormLabel>ES 数据源配置</FormLabel>
                      <Select
                        value={field.value ? String(field.value) : 'none'}
                        onValueChange={(value) => {
                          const numValue = value === 'none' ? undefined : Number(value);
                          field.onChange(numValue);
                        }}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="使用默认数据源">
                              {selectedConfig
                                ? `${selectedConfig.name}${selectedConfig.is_default ? ' (默认)' : ''}`
                                : field.value
                                  ? `配置 ID: ${field.value}`
                                  : '使用默认数据源'
                              }
                            </SelectValue>
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="none">使用默认数据源</SelectItem>
                          {esConfigs
                            ?.filter((c) => c?.enabled)
                            ?.map((config) => (
                              <SelectItem key={config.id} value={String(config.id)}>
                                {config.name} {config.is_default && '(默认)'}
                              </SelectItem>
                            ))}
                        </SelectContent>
                      </Select>
                      <FormDescription>
                        选择用于查询的 Elasticsearch 数据源配置，如不选择则使用默认配置
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  );
                }}
              />

              <FormField
                control={form.control}
                name="lark_config_id"
                render={({ field }) => {
                  // 获取当前选中的配置名称
                  const selectedConfig = larkConfigs?.find((c) => c?.id === field.value);

                  return (
                    <FormItem>
                      <FormLabel>Lark 告警配置</FormLabel>
                      <Select
                        value={field.value ? String(field.value) : 'none'}
                        onValueChange={(value) => {
                          const numValue = value === 'none' ? undefined : Number(value);
                          field.onChange(numValue);
                          // Clear direct webhook URL when selecting config
                          if (numValue) {
                            form.setValue('lark_webhook', '');
                          }
                        }}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder="不使用配置（直接输入 URL）">
                              {selectedConfig
                                ? `${selectedConfig.name}${selectedConfig.is_default ? ' (默认)' : ''}`
                                : field.value
                                  ? `配置 ID: ${field.value}`
                                  : '不使用配置（直接输入 URL）'
                              }
                            </SelectValue>
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          <SelectItem value="none">不使用配置（直接输入 URL）</SelectItem>
                          {larkConfigs
                            ?.filter((c) => c?.enabled)
                            ?.map((config) => (
                              <SelectItem key={config.id} value={String(config.id)}>
                                {config.name} {config.is_default && '(默认)'}
                              </SelectItem>
                            ))}
                        </SelectContent>
                      </Select>
                      <FormDescription>
                        选择已配置的 Lark 配置，或选择"不使用配置"直接输入 URL
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  );
                }}
              />

              <FormField
                control={form.control}
                name="lark_webhook"
                rules={{
                  validate: (value, formValues) => {
                    // Require either config or direct URL
                    if (!formValues.lark_config_id && !value) {
                      return '请选择 Lark 配置或输入 Webhook URL';
                    }
                    return true;
                  },
                }}
                render={({ field }) => {
                  const hasConfig = form.watch('lark_config_id');
                  return (
                    <FormItem>
                      <FormLabel>
                        Lark Webhook URL {hasConfig && '(可选，使用配置时可不填)'}
                      </FormLabel>
                      <FormControl>
                        <Input
                          placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..."
                          {...field}
                          disabled={!!hasConfig}
                          onChange={(e) => {
                            field.onChange(e);
                            // Clear config when typing direct URL
                            if (e.target.value && hasConfig) {
                              form.setValue('lark_config_id', undefined);
                            }
                          }}
                        />
                      </FormControl>
                      <FormDescription>
                        {hasConfig
                          ? '已选择配置，可直接使用配置中的 Webhook URL'
                          : '直接输入 Webhook URL（当未选择配置时必填）'}
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  );
                }}
              />

              <FormField
                control={form.control}
                name="queries"
                rules={{
                  required: '请配置查询条件',
                  validate: (value) => {
                    if (!value || (Array.isArray(value) && value.length === 0)) {
                      return '请至少配置一个查询条件';
                    }
                    if (typeof value === 'string') {
                      try {
                        const parsed = JSON.parse(value);
                        if (!Array.isArray(parsed) || parsed.length === 0) {
                          return '查询条件必须是非空数组';
                        }
                      } catch {
                        return '查询条件必须是有效的 JSON 格式';
                      }
                    }
                    return true;
                  },
                }}
                render={({ field }) => (
                  <FormItem>
                    <div className="flex justify-between items-center">
                      <FormLabel>查询条件 (JSON)</FormLabel>
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={handleTest}
                        disabled={isTesting}
                      >
                        {isTesting ? (
                          <>
                            <Loader2 className="mr-1 h-3 w-3 animate-spin" />
                            测试中
                          </>
                        ) : (
                          <>
                            <Play className="mr-1 h-3 w-3" />
                            测试
                          </>
                        )}
                      </Button>
                    </div>
                    <FormControl>
                      <Textarea
                        rows={15}
                        placeholder='请输入 JSON 格式的查询条件...'
                        {...field}
                        value={
                          typeof field.value === 'string'
                            ? field.value
                            : JSON.stringify(field.value || [], null, 2)
                        }
                        onChange={(e) => {
                          // 直接保存用户输入的字符串，不做自动转换
                          // 这样用户可以自由编辑，包括换行、缩进等
                          field.onChange(e.target.value);
                        }}
                        onKeyDown={(e) => {
                          // 支持 Tab 键缩进
                          if (e.key === 'Tab') {
                            e.preventDefault();
                            const start = e.currentTarget.selectionStart;
                            const end = e.currentTarget.selectionEnd;
                            const value = e.currentTarget.value;
                            const newValue = value.substring(0, start) + '  ' + value.substring(end);
                            e.currentTarget.value = newValue;
                            e.currentTarget.selectionStart = e.currentTarget.selectionEnd = start + 2;
                            field.onChange(newValue);
                          }
                        }}
                        className="font-mono text-sm resize-y"
                        style={{ minHeight: '300px' }}
                      />
                    </FormControl>
                    <div className="flex flex-wrap gap-2 mt-2">
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          try {
                            const current = form.getValues('queries');
                            const text = typeof current === 'string' ? current : JSON.stringify(current);
                            const parsed = JSON.parse(text);
                            const formatted = JSON.stringify(parsed, null, 2);
                            form.setValue('queries', formatted as any);
                            toast.success('格式化成功');
                          } catch (e: any) {
                            toast.error('JSON 格式错误', e.message);
                          }
                        }}
                      >
                        ✨ 格式化
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          const template = `[
  {
    "field": "response_code",
    "operator": "!=",
    "value": 200
  }
]`;
                          form.setValue('queries', template as any);
                        }}
                      >
                        📝 非200响应
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          const template = `[
  {
    "field": "response_code",
    "operator": ">=",
    "value": 400
  }
]`;
                          form.setValue('queries', template as any);
                        }}
                      >
                        🚫 4xx/5xx错误
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          const template = `[
  {
    "field": "response_code",
    "operator": "=",
    "value": 500,
    "logic": "or"
  },
  {
    "field": "response_code",
    "operator": "=",
    "value": 502,
    "logic": "or"
  },
  {
    "field": "response_code",
    "operator": "=",
    "value": 503,
    "logic": "or"
  }
]`;
                          form.setValue('queries', template as any);
                        }}
                      >
                        ⚠️ 5xx错误
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          const template = `[
  {
    "field": "responsetime",
    "operator": ">",
    "value": 3,
    "logic": "and"
  },
  {
    "field": "response_code",
    "operator": "=",
    "value": 200,
    "logic": "and"
  }
]`;
                          form.setValue('queries', template as any);
                        }}
                      >
                        🐌 慢查询
                      </Button>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={() => {
                          const template = `[
  {
    "field": "response_code",
    "operator": "=",
    "value": 499
  }
]`;
                          form.setValue('queries', template as any);
                        }}
                      >
                        🔌 499错误
                      </Button>
                    </div>
                    <FormDescription>
                      <div className="space-y-3 text-sm">
                        <p>
                          <strong>必填字段</strong>:
                          <code className="bg-muted px-1 rounded mx-1">field</code>
                          <code className="bg-muted px-1 rounded mx-1">operator</code>
                          <code className="bg-muted px-1 rounded mx-1">value</code>
                          |
                          <strong className="ml-1">可选</strong>:
                          <code className="bg-muted px-1 rounded mx-1">logic</code> (and/or)
                        </p>
                        <p className="text-xs text-muted-foreground">
                          💡 支持 <kbd className="px-1 py-0.5 text-xs border rounded">Tab</kbd> 键缩进 |
                          点击上方模板快速开始 |
                          "格式化"美化代码
                        </p>

                        <details className="border rounded-lg bg-muted/50">
                          <summary className="p-3 cursor-pointer font-semibold text-foreground hover:bg-muted/70 rounded-lg">
                            📚 支持的操作符 (operator) - 点击展开查看
                          </summary>
                          <div className="p-3 pt-0 space-y-2">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-2 text-xs">
                              <div>
                                <span className="font-medium">比较操作符：</span>
                                <ul className="ml-4 mt-1 space-y-0.5">
                                  <li><code className="bg-background px-1 rounded">=</code> / <code className="bg-background px-1 rounded">==</code> / <code className="bg-background px-1 rounded">equals</code> - 等于</li>
                                  <li><code className="bg-background px-1 rounded">!=</code> / <code className="bg-background px-1 rounded">not_equals</code> - 不等于</li>
                                  <li><code className="bg-background px-1 rounded">&gt;</code> / <code className="bg-background px-1 rounded">gt</code> - 大于</li>
                                  <li><code className="bg-background px-1 rounded">&gt;=</code> / <code className="bg-background px-1 rounded">gte</code> - 大于等于</li>
                                  <li><code className="bg-background px-1 rounded">&lt;</code> / <code className="bg-background px-1 rounded">lt</code> - 小于</li>
                                  <li><code className="bg-background px-1 rounded">&lt;=</code> / <code className="bg-background px-1 rounded">lte</code> - 小于等于</li>
                                </ul>
                              </div>
                              <div>
                                <span className="font-medium">文本/存在性：</span>
                                <ul className="ml-4 mt-1 space-y-0.5">
                                  <li><code className="bg-background px-1 rounded">contains</code> - 包含（文本匹配）</li>
                                  <li><code className="bg-background px-1 rounded">not_contains</code> - 不包含</li>
                                  <li><code className="bg-background px-1 rounded">exists</code> - 字段存在</li>
                                </ul>
                              </div>
                            </div>
                            <div className="pt-2 border-t text-xs">
                              <span className="font-medium">示例：</span>
                              <code className="block bg-background p-2 rounded mt-1 overflow-x-auto">
                                {`{"field": "response_code", "operator": "=", "value": 499}`}
                              </code>
                            </div>
                          </div>
                        </details>
                      </div>
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="flex gap-4 pt-6 mt-6 border-t">
                <Button
                  type="submit"
                  disabled={createMutation.isPending || updateMutation.isPending}
                  className="min-w-24"
                >
                  {createMutation.isPending || updateMutation.isPending
                    ? '保存中...'
                    : '保存'}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => navigate('/rules')}
                  className="min-w-24"
                >
                  取消
                </Button>
              </div>
            </form>
          </Form>
        </CardContent>
      </Card>

      {/* Test Result Dialog */}
      <Dialog open={testDialogOpen} onOpenChange={setTestDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>测试结果</DialogTitle>
            <DialogDescription>
              {testResult?.success
                ? `找到 ${testResult?.data?.count || 0} 条匹配日志`
                : '测试失败'}
            </DialogDescription>
          </DialogHeader>
          {testResult && (
            <div className="space-y-4">
              {testResult.success ? (
                <>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">匹配数量</p>
                      <p className="text-2xl font-bold">{testResult.data.count}</p>
                    </div>
                    <div>
                      <p className="text-sm font-medium text-muted-foreground">时间范围</p>
                      <p className="text-sm">
                        {testResult.data.time_range.from}
                        <br />
                        ~ {testResult.data.time_range.to}
                      </p>
                    </div>
                  </div>
                  {testResult.data.logs && testResult.data.logs.length > 0 && (
                    <div>
                      <p className="text-sm font-medium text-muted-foreground mb-2">
                        日志示例（最多显示 10 条）
                      </p>
                      <pre className="bg-muted p-4 rounded-md overflow-auto max-h-96 text-xs">
                        {JSON.stringify(
                          testResult.data.logs.slice(0, 10),
                          null,
                          2
                        )}
                      </pre>
                    </div>
                  )}
                </>
              ) : (
                <div className="text-destructive">
                  <p className="font-medium">错误信息：</p>
                  <p className="text-sm mt-2">{testResult.error}</p>
                </div>
              )}
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setTestDialogOpen(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
