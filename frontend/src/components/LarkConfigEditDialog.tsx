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
import { larkConfigApi, LarkConfig } from '../services/api';
import { useToast } from '@/contexts/ToastContext';

interface LarkConfigEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  config?: LarkConfig | null;
  onSuccess?: () => void;
}

export default function LarkConfigEditDialog({
  open,
  onOpenChange,
  config,
  onSuccess,
}: LarkConfigEditDialogProps) {
  const toast = useToast();
  const isEdit = !!config;

  const form = useForm<Partial<LarkConfig>>({
    defaultValues: {
      name: '',
      webhook_url: '',
      description: '',
      enabled: true,
      is_default: false,
    },
  });

  useEffect(() => {
    if (config) {
      form.reset(config);
    } else {
      form.reset({
        name: '',
        webhook_url: '',
        description: '',
        enabled: true,
        is_default: false,
      });
    }
  }, [config, form, open]);

  const createMutation = useMutation({
    mutationFn: (config: Partial<LarkConfig>) => larkConfigApi.create(config),
    onSuccess: () => {
      toast.success('创建成功');
      onSuccess?.();
    },
    onError: (error: any) => {
      toast.error('创建失败', error?.response?.data?.error || '未知错误');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, config }: { id: number; config: Partial<LarkConfig> }) =>
      larkConfigApi.update(id, config),
    onSuccess: () => {
      toast.success('更新成功');
      onSuccess?.();
    },
    onError: (error: any) => {
      toast.error('更新失败', error?.response?.data?.error || '未知错误');
    },
  });

  const onSubmit = (values: Partial<LarkConfig>) => {
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
          <DialogTitle>{isEdit ? '编辑 Lark 配置' : '新建 Lark 配置'}</DialogTitle>
          <DialogDescription>
            {isEdit ? '修改 Lark Webhook 配置' : '创建新的 Lark Webhook 配置'}
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
                    <Input placeholder="例如：生产环境通知" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="webhook_url"
              rules={{ required: '请输入 Webhook URL' }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Webhook URL</FormLabel>
                  <FormControl>
                    <Input placeholder="https://open.feishu.cn/open-apis/bot/v2/hook/..." {...field} />
                  </FormControl>
                  <FormDescription>
                    Lark 机器人 Webhook 地址
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

