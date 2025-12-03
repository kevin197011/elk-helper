// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Edit, Trash2, Play, Star, StarOff, Loader2 } from 'lucide-react';
import { larkConfigApi, LarkConfig } from '../services/api';
import { useToast } from '@/contexts/ToastContext';
import LarkConfigEditDialog from '../components/LarkConfigEditDialog';

export default function LarkConfigPage() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedConfig, setSelectedConfig] = useState<LarkConfig | null>(null);
  const [configToDelete, setConfigToDelete] = useState<LarkConfig | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['lark-configs'],
    queryFn: () => larkConfigApi.getAll().then(res => res.data.data),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => larkConfigApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
      toast.success('删除成功');
      setDeleteDialogOpen(false);
    },
    onError: (error: any) => {
      toast.error('删除失败', error?.response?.data?.error || '未知错误');
    },
  });

  const testMutation = useMutation({
    mutationFn: (id: number) => larkConfigApi.test(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
      if (data.data.success) {
        toast.success('连接测试成功');
      } else {
        toast.error('连接测试失败', data.data.error);
      }
    },
    onError: () => {
      toast.error('测试失败');
    },
  });

  const setDefaultMutation = useMutation({
    mutationFn: (id: number) => larkConfigApi.setDefault(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
      toast.success('已设置为默认配置');
    },
    onError: () => {
      toast.error('设置失败');
    },
  });

  const handleCreate = () => {
    setSelectedConfig(null);
    setEditDialogOpen(true);
  };

  const handleEdit = (config: LarkConfig) => {
    setSelectedConfig(config);
    setEditDialogOpen(true);
  };

  const handleDelete = (config: LarkConfig) => {
    setConfigToDelete(config);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = () => {
    if (configToDelete) {
      deleteMutation.mutate(configToDelete.id);
    }
  };

  const handleTest = (id: number) => {
    testMutation.mutate(id);
  };

  const handleSetDefault = (id: number) => {
    setDefaultMutation.mutate(id);
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-3xl font-bold tracking-tight">Lark 告警配置</h2>
        <Button onClick={handleCreate}>
          <Plus className="mr-2 h-4 w-4" />
          新建配置
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-4">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      ) : (
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[80px]">ID</TableHead>
                <TableHead>配置名称</TableHead>
                <TableHead>Webhook URL</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>测试状态</TableHead>
                <TableHead>默认</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data && data.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                    暂无配置，点击"新建配置"创建第一个 Lark 配置
                  </TableCell>
                </TableRow>
              ) : (
                data?.map((config) => (
                  <TableRow key={config.id}>
                    <TableCell>{config.id}</TableCell>
                    <TableCell className="font-medium">{config.name}</TableCell>
                    <TableCell className="font-mono text-sm max-w-md truncate" title={config.webhook_url}>
                      {config.webhook_url}
                    </TableCell>
                    <TableCell>
                      <Badge variant={config.enabled ? 'default' : 'secondary'}>
                        {config.enabled ? '启用' : '禁用'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {config.test_status === 'success' ? (
                        <Badge variant="default">成功</Badge>
                      ) : config.test_status === 'failed' ? (
                        <Badge variant="destructive">失败</Badge>
                      ) : (
                        <Badge variant="secondary">未测试</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {config.is_default ? (
                        <Star className="h-4 w-4 text-yellow-500 fill-yellow-500" />
                      ) : (
                        <StarOff className="h-4 w-4 text-muted-foreground" />
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleTest(config.id)}
                          disabled={testMutation.isPending}
                          title="测试连接"
                        >
                          {testMutation.isPending ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Play className="h-4 w-4" />
                          )}
                        </Button>
                        {!config.is_default && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleSetDefault(config.id)}
                            disabled={setDefaultMutation.isPending}
                            title="设为默认"
                          >
                            <Star className="h-4 w-4" />
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleEdit(config)}
                          title="编辑"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(config)}
                          title="删除"
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </div>
      )}

      <LarkConfigEditDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        config={selectedConfig}
        onSuccess={() => {
          setEditDialogOpen(false);
          queryClient.invalidateQueries({ queryKey: ['lark-configs'] });
        }}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              确定要删除 Lark 配置 "{configToDelete?.name}" 吗？如果有规则使用此配置，将无法删除。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              取消
            </Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={deleteMutation.isPending}>
              {deleteMutation.isPending ? '删除中...' : '确认删除'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

