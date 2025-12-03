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
import { esConfigApi, ESConfig } from '../services/api';
import { useToast } from '@/contexts/ToastContext';
import ESConfigEditDialog from '../components/ESConfigEditDialog';

export default function ESConfigPage() {
  const queryClient = useQueryClient();
  const toast = useToast();
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedConfig, setSelectedConfig] = useState<ESConfig | null>(null);
  const [configToDelete, setConfigToDelete] = useState<ESConfig | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['es-configs'],
    queryFn: () => esConfigApi.getAll().then(res => res.data.data),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: number) => esConfigApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['es-configs'] });
      toast.success('删除成功');
      setDeleteDialogOpen(false);
    },
    onError: () => {
      toast.error('删除失败');
    },
  });

  const testMutation = useMutation({
    mutationFn: (id: number) => esConfigApi.test(id),
    onSuccess: (data) => {
      // 无论成功失败都刷新数据，确保测试状态更新
      queryClient.invalidateQueries({ queryKey: ['es-configs'] });
      if (data.data.success) {
        toast.success('连接测试成功');
      } else {
        toast.error('连接测试失败', data.data.error || '未知错误');
      }
    },
    onError: (error: any) => {
      // 即使请求失败也刷新数据
      queryClient.invalidateQueries({ queryKey: ['es-configs'] });
      const errorMsg = error?.response?.data?.error || error?.message || '测试请求失败';
      toast.error('测试失败', errorMsg);
    },
  });

  const setDefaultMutation = useMutation({
    mutationFn: (id: number) => esConfigApi.setDefault(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['es-configs'] });
      toast.success('已设置为默认数据源');
    },
    onError: () => {
      toast.error('设置失败');
    },
  });

  const handleCreate = () => {
    setSelectedConfig(null);
    setEditDialogOpen(true);
  };

  const handleEdit = (config: ESConfig) => {
    setSelectedConfig(config);
    setEditDialogOpen(true);
  };

  const handleDelete = (config: ESConfig) => {
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
        <h2 className="text-3xl font-bold tracking-tight">ES 数据源配置</h2>
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
                <TableHead>ES 地址</TableHead>
                <TableHead>用户名</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>测试状态</TableHead>
                <TableHead>默认</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data && data.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                    暂无配置，点击"新建配置"创建第一个数据源配置
                  </TableCell>
                </TableRow>
              ) : (
                data?.map((config) => (
                  <TableRow key={config.id}>
                    <TableCell>{config.id}</TableCell>
                    <TableCell className="font-medium">{config.name}</TableCell>
                    <TableCell className="font-mono text-sm">{config.url}</TableCell>
                    <TableCell>{config.username || '-'}</TableCell>
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

      <ESConfigEditDialog
        open={editDialogOpen}
        onOpenChange={setEditDialogOpen}
        config={selectedConfig}
        onSuccess={() => {
          setEditDialogOpen(false);
          queryClient.invalidateQueries({ queryKey: ['es-configs'] });
        }}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              确定要删除数据源配置 "{configToDelete?.name}" 吗？此操作不可恢复。
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

