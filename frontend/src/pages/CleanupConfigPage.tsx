// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { systemConfigApi, CleanupConfig } from '../services/api';
import { useToast } from '@/contexts/ToastContext';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { Skeleton } from '@/components/ui/skeleton';
import { Loader2, Save, Trash2, CheckCircle2, XCircle, Clock } from 'lucide-react';
import { useForm } from 'react-hook-form';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';

export default function CleanupConfigPage() {
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [isSaving, setIsSaving] = useState(false);
  const [isCleanupDialogOpen, setIsCleanupDialogOpen] = useState(false);
  const [isCleaning, setIsCleaning] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ['cleanup-config'],
    queryFn: () => systemConfigApi.getCleanupConfig().then(res => res.data.data),
  });

  const form = useForm<CleanupConfig>({
    defaultValues: {
      enabled: true,
      hour: 3,
      minute: 0,
      retention_days: 90,
    },
  });

  useEffect(() => {
    if (data) {
      form.reset({
        enabled: data.enabled,
        hour: data.hour,
        minute: data.minute,
        retention_days: data.retention_days,
      });
    }
  }, [data, form]);

  const updateMutation = useMutation({
    mutationFn: (config: CleanupConfig) => systemConfigApi.updateCleanupConfig(config),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['cleanup-config'] });
      toast({
        title: '保存成功',
        description: '清理任务配置已更新',
      });
      setIsSaving(false);
    },
    onError: (error: any) => {
      toast({
        title: '保存失败',
        description: error.response?.data?.error || '更新配置时发生错误',
        variant: 'error',
      });
      setIsSaving(false);
    },
  });

  const onSubmit = (formData: CleanupConfig) => {
    setIsSaving(true);
    updateMutation.mutate(formData);
  };

  const manualCleanupMutation = useMutation({
    mutationFn: () => systemConfigApi.manualCleanup(),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      queryClient.invalidateQueries({ queryKey: ['cleanup-config'] });
      toast({
        title: '清理完成',
        description: `已删除 ${response.data.deleted_count} 条超过 ${response.data.retention_days} 天的历史告警数据`,
      });
      setIsCleaning(false);
      setIsCleanupDialogOpen(false);
    },
    onError: (error: any) => {
      toast({
        title: '清理失败',
        description: error.response?.data?.error || '执行清理时发生错误',
        variant: 'error',
      });
      setIsCleaning(false);
    },
  });

  const handleManualCleanup = () => {
    setIsCleaning(true);
    manualCleanupMutation.mutate();
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">清理任务配置</h2>
          <p className="text-muted-foreground">配置定时清理历史告警数据</p>
        </div>
        <Card>
          <CardHeader>
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-96 mt-2" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">清理任务配置</h2>
          <p className="text-muted-foreground">配置定时清理历史告警数据</p>
        </div>
        <Button
          variant="destructive"
          onClick={() => setIsCleanupDialogOpen(true)}
          disabled={isCleaning}
        >
          {isCleaning ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              清理中...
            </>
          ) : (
            <>
              <Trash2 className="mr-2 h-4 w-4" />
              立即清理
            </>
          )}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>定时清理设置</CardTitle>
          <CardDescription>
            系统将自动删除超过保留期限的历史告警数据，以节省存储空间
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="enabled" className="text-base">
                  启用清理任务
                </Label>
                <p className="text-sm text-muted-foreground">
                  开启后，系统将按设定的时间自动清理历史数据
                </p>
              </div>
              <Switch
                id="enabled"
                checked={form.watch('enabled')}
                onCheckedChange={(checked) => form.setValue('enabled', checked)}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="hour">执行时间 - 小时</Label>
                <Input
                  id="hour"
                  type="number"
                  min="0"
                  max="23"
                  {...form.register('hour', {
                    required: '请输入小时',
                    min: { value: 0, message: '小时必须在 0-23 之间' },
                    max: { value: 23, message: '小时必须在 0-23 之间' },
                    valueAsNumber: true,
                  })}
                />
                {form.formState.errors.hour && (
                  <p className="text-sm text-red-600">{form.formState.errors.hour.message}</p>
                )}
                <p className="text-sm text-muted-foreground">0-23</p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="minute">执行时间 - 分钟</Label>
                <Input
                  id="minute"
                  type="number"
                  min="0"
                  max="59"
                  {...form.register('minute', {
                    required: '请输入分钟',
                    min: { value: 0, message: '分钟必须在 0-59 之间' },
                    max: { value: 59, message: '分钟必须在 0-59 之间' },
                    valueAsNumber: true,
                  })}
                />
                {form.formState.errors.minute && (
                  <p className="text-sm text-red-600">{form.formState.errors.minute.message}</p>
                )}
                <p className="text-sm text-muted-foreground">0-59</p>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="retention_days">数据保留天数</Label>
              <Input
                id="retention_days"
                type="number"
                min="1"
                {...form.register('retention_days', {
                  required: '请输入保留天数',
                  min: { value: 1, message: '保留天数必须至少为 1 天' },
                  valueAsNumber: true,
                })}
              />
              {form.formState.errors.retention_days && (
                <p className="text-sm text-red-600">{form.formState.errors.retention_days.message}</p>
              )}
              <p className="text-sm text-muted-foreground">
                超过此天数的告警数据将被自动删除。例如：90 天表示删除 3 个月前的数据
              </p>
            </div>

            <div className="flex justify-end gap-4">
              <Button type="submit" disabled={isSaving}>
                {isSaving ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    保存中...
                  </>
                ) : (
                  <>
                    <Save className="mr-2 h-4 w-4" />
                    保存配置
                  </>
                )}
              </Button>
            </div>
          </form>

          {form.watch('enabled') && (
            <div className="mt-6 p-4 bg-muted rounded-lg">
              <h4 className="font-medium mb-2">当前配置预览</h4>
              <p className="text-sm text-muted-foreground">
                系统将在每天 <strong>{String(form.watch('hour')).padStart(2, '0')}:{String(form.watch('minute')).padStart(2, '0')}</strong> 执行清理任务，
                删除 <strong>{form.watch('retention_days')} 天</strong>前的告警数据
              </p>
            </div>
          )}

          {/* Last Execution Status */}
          {data && (
            <div className="mt-6 p-4 border rounded-lg">
              <h4 className="font-medium mb-3 flex items-center gap-2">
                上次执行状态
              </h4>
              {!data.last_execution_status || data.last_execution_status === 'never' ? (
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Clock className="h-4 w-4" />
                  <span className="text-sm">尚未执行</span>
                </div>
              ) : data.last_execution_status === 'success' ? (
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-green-600">
                    <CheckCircle2 className="h-4 w-4" />
                    <span className="text-sm font-medium">执行成功</span>
                  </div>
                  {data.last_execution_time && (
                    <p className="text-sm text-muted-foreground ml-6">
                      执行时间: {new Date(data.last_execution_time).toLocaleString('zh-CN')}
                    </p>
                  )}
                  {data.last_execution_result && (
                    <p className="text-sm text-muted-foreground ml-6">
                      {data.last_execution_result}
                    </p>
                  )}
                </div>
              ) : data.last_execution_status === 'failed' ? (
                <div className="space-y-2">
                  <div className="flex items-center gap-2 text-red-600">
                    <XCircle className="h-4 w-4" />
                    <span className="text-sm font-medium">执行失败</span>
                  </div>
                  {data.last_execution_time && (
                    <p className="text-sm text-muted-foreground ml-6">
                      执行时间: {new Date(data.last_execution_time).toLocaleString('zh-CN')}
                    </p>
                  )}
                  {data.last_execution_result && (
                    <p className="text-sm text-red-600 ml-6">
                      {data.last_execution_result}
                    </p>
                  )}
                </div>
              ) : null}
            </div>
          )}
        </CardContent>
      </Card>

      <AlertDialog open={isCleanupDialogOpen} onOpenChange={setIsCleanupDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认立即清理</AlertDialogTitle>
            <AlertDialogDescription>
              此操作将立即删除超过 <strong>{form.watch('retention_days')} 天</strong> 的历史告警数据。
              <br />
              <br />
              此操作不可恢复，确定要继续吗？
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isCleaning}>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleManualCleanup}
              disabled={isCleaning}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isCleaning ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  清理中...
                </>
              ) : (
                '确认清理'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

