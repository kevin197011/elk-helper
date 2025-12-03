// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useMutation } from '@tanstack/react-query';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { esConfigApi, ESConfig } from '../services/api';
import { useToast } from '@/contexts/ToastContext';

interface ESConfigEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config?: ESConfig | null;
  onSuccess?: () => void;
}

export default function ESConfigEditDialog({
  open,
  onOpenChange,
  config,
  onSuccess,
}: ESConfigEditDialogProps) {
  const toast = useToast();
  const isEdit = !!config;

  const form = useForm<Partial<ESConfig>>({
    defaultValues: {
      name: '',
      url: '',
      username: '',
      password: '',
      use_ssl: false,
      skip_verify: false,
      ca_certificate: '',
      description: '',
      enabled: true,
      is_default: false,
    },
  });

  useEffect(() => {
    if (config) {
      form.reset({
        ...config,
        password: '', // Don't show password when editing
        ca_certificate: '', // Don't show CA cert when editing
      });
    } else {
      form.reset({
        name: '',
        url: '',
        username: '',
        password: '',
        use_ssl: false,
        skip_verify: false,
        ca_certificate: '',
        description: '',
        enabled: true,
        is_default: false,
      });
    }
  }, [config, form, open]);

  const createMutation = useMutation({
    mutationFn: (config: Partial<ESConfig>) => esConfigApi.create(config),
    onSuccess: () => {
      toast.success('创建成功');
      onSuccess?.();
    },
    onError: (error: any) => {
      toast.error('创建失败', error?.response?.data?.error || '未知错误');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, config }: { id: number; config: Partial<ESConfig> }) =>
      esConfigApi.update(id, config),
    onSuccess: () => {
      toast.success('更新成功');
      onSuccess?.();
    },
    onError: (error: any) => {
      toast.error('更新失败', error?.response?.data?.error || '未知错误');
    },
  });

  const onSubmit = (values: Partial<ESConfig>) => {
    if (isEdit && config) {
      updateMutation.mutate({ id: config.id, config: values });
    } else {
      createMutation.mutate(values);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? '编辑数据源配置' : '新建数据源配置'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '修改 Elasticsearch 数据源连接配置' : '创建新的 Elasticsearch 数据源连接配置'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              rules={{ required: '请输入配置名称' }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>配置名称</FormLabel>
                  <FormControl>
                    <Input placeholder="例如：生产环境 ES" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="url"
              rules={{
                required: '请输入 ES 地址',
                validate: (value) => {
                  // Validate URL format
                  if (!value) return '请输入 ES 地址';

                  const urls = value.split(';').map(u => u.trim()).filter(u => u);
                  if (urls.length === 0) return '请输入有效的 ES 地址';

                  // Validate each URL
                  for (const url of urls) {
                    if (!url.startsWith('http://') && !url.startsWith('https://')) {
                      return `地址必须以 http:// 或 https:// 开头: ${url}`;
                    }
                  }

                  return true;
                }
              }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>ES 地址</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={3}
                      placeholder="单个节点: https://10.170.1.54:9200&#10;多个节点（分号分隔，轮询查询）:&#10;https://10.170.1.54:9200;https://10.170.1.55:9200;https://10.170.1.56:9200"
                      {...field}
                      onChange={(e) => {
                        field.onChange(e);
                        // Auto-enable SSL if any URL starts with https://
                        const value = e.target.value.trim().toLowerCase();
                        if (value.includes('https://')) {
                          form.setValue('use_ssl', true);
                        } else if (value.startsWith('http://') && !value.includes('https://')) {
                          form.setValue('use_ssl', false);
                        }
                      }}
                    />
                  </FormControl>
                  <FormDescription>
                    <div className="space-y-1">
                      <div>• <strong>单节点</strong>: 直接输入地址，如 https://10.170.1.54:9200</div>
                      <div>• <strong>多节点</strong>: 用分号分隔多个地址，系统会轮询查询以实现负载均衡</div>
                      <div className="text-xs text-muted-foreground mt-1">
                        示例: https://es1.com:9200;https://es2.com:9200;https://es3.com:9200
                      </div>
                    </div>
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="username"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>用户名（可选）</FormLabel>
                  <FormControl>
                    <Input placeholder="留空则使用无认证连接" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="password"
              render={({ field }) => {
                // Show placeholder dots if editing and password field is empty (meaning password exists in DB but not shown)
                const placeholder = isEdit && !field.value ? '••••••••' : '密码';
                return (
                  <FormItem>
                    <FormLabel>密码{isEdit ? '（留空则不修改）' : ''}</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        placeholder={placeholder}
                        value={field.value || ''}
                        onChange={(e) => {
                          field.onChange(e.target.value);
                        }}
                        onFocus={(e) => {
                          // Clear placeholder when user starts typing
                          if (isEdit && e.target.value === '') {
                            e.target.placeholder = '输入新密码';
                          }
                        }}
                        onBlur={(e) => {
                          // Restore placeholder if still empty after blur
                          if (isEdit && e.target.value === '') {
                            e.target.placeholder = '••••••••';
                          }
                        }}
                      />
                    </FormControl>
                    {isEdit && (
                      <FormDescription>
                        当前已配置密码。留空则保持原密码不变，输入新密码则更新
                      </FormDescription>
                    )}
                    <FormMessage />
                  </FormItem>
                );
              }}
            />

            <FormField
              control={form.control}
              name="use_ssl"
              render={({ field }) => (
                <FormItem className="flex items-center justify-between rounded-lg border p-4">
                  <div className="space-y-0.5">
                    <FormLabel>启用 SSL/TLS</FormLabel>
                    <FormDescription>
                      启用后使用 HTTPS 加密连接
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch checked={field.value || false} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />

            {form.watch('use_ssl') && (
              <>
                <FormField
                  control={form.control}
                  name="skip_verify"
                  render={({ field }) => (
                    <FormItem className="flex items-center justify-between rounded-lg border p-4">
                      <div className="space-y-0.5">
                        <FormLabel>跳过证书验证</FormLabel>
                        <FormDescription>
                          启用后不验证服务器证书，适用于自签名证书或内部环境。⚠️ 仅用于开发/测试环境，生产环境不建议使用
                        </FormDescription>
                      </div>
                      <FormControl>
                        <Switch checked={field.value || false} onCheckedChange={field.onChange} />
                      </FormControl>
                    </FormItem>
                  )}
                />

                {!form.watch('skip_verify') && (
                  <FormField
                    control={form.control}
                    name="ca_certificate"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>CA 证书{isEdit ? '（留空则不修改）' : ''}（可选）</FormLabel>
                        <FormControl>
                          <Textarea
                            rows={6}
                            placeholder="PEM 格式的 CA 证书内容，留空则使用系统默认证书。如果跳过证书验证，则不需要填写"
                            {...field}
                          />
                        </FormControl>
                        <FormDescription>
                          用于验证服务器证书的 CA 证书（PEM 格式），通常用于自签名证书。如果已启用"跳过证书验证"，则不需要填写此字段
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                )}
              </>
            )}

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>描述</FormLabel>
                  <FormControl>
                    <Textarea rows={2} placeholder="配置描述（可选）" {...field} />
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
                    <FormLabel>启用配置</FormLabel>
                    <FormDescription>
                      禁用后此配置将不会被使用
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="is_default"
              render={({ field }) => (
                <FormItem className="flex items-center justify-between rounded-lg border p-4">
                  <div className="space-y-0.5">
                    <FormLabel>设为默认</FormLabel>
                    <FormDescription>
                      设为默认后，其他配置的默认状态将被取消
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch checked={field.value} onCheckedChange={field.onChange} />
                  </FormControl>
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                取消
              </Button>
              <Button
                type="submit"
                disabled={createMutation.isPending || updateMutation.isPending}
              >
                {createMutation.isPending || updateMutation.isPending ? '保存中...' : '保存'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

