// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import { Switch } from '@/components/ui/switch';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Plus, Edit, Trash2, Search, X, Download, Upload } from 'lucide-react';
import { rulesApi } from '../services/api';
import { useToast } from '@/contexts/ToastContext';

export default function RulesPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToast();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [ruleToDelete, setRuleToDelete] = useState<{ id: number; name: string } | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all');
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['rules'],
    queryFn: () => rulesApi.getAll().then(res => res.data.data),
    refetchInterval: false, // Disable auto refetch to prevent jumping
    staleTime: 30000, // Consider data fresh for 30 seconds
  });

  // Filter and sort rules
  const filteredRules = useMemo(() => {
    if (!data) return [];

    let filtered = [...data]; // Create a copy to avoid mutating original data

    // Filter by status
    if (statusFilter !== 'all') {
      filtered = filtered.filter(rule =>
        statusFilter === 'enabled' ? rule.enabled : !rule.enabled
      );
    }

    // Filter by search query
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(rule =>
        rule.name.toLowerCase().includes(query) ||
        rule.index_pattern.toLowerCase().includes(query) ||
        (rule.description && rule.description.toLowerCase().includes(query))
      );
    }

    // Sort by ID descending (newest first)
    filtered.sort((a, b) => b.id - a.id);

    return filtered;
  }, [data, searchQuery, statusFilter]);

  const deleteMutation = useMutation({
    mutationFn: (id: number) => rulesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      toast.success('删除成功');
      setDeleteDialogOpen(false);
    },
    onError: () => {
      toast.error('删除失败');
    },
  });

  const toggleMutation = useMutation({
    mutationFn: (id: number) => rulesApi.toggleEnabled(id),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      toast.success(
        data.data.data.enabled ? '规则已启用' : '规则已禁用'
      );
    },
    onError: () => {
      toast.error('操作失败');
    },
  });

  const exportMutation = useMutation({
    mutationFn: async () => {
      const response = await rulesApi.export();
      return response.data;
    },
    onSuccess: (data) => {
      // Create blob and download
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `rules_export_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.json`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      toast.success('规则导出成功');
    },
    onError: () => {
      toast.error('规则导出失败');
    },
  });

  const importMutation = useMutation({
    mutationFn: (rules: any[]) => rulesApi.import(rules),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['rules'] });
      const { success_count, error_count, errors } = response.data;
      if (error_count === 0) {
        toast.success(`成功导入 ${success_count} 条规则`);
      } else {
        toast.error(`导入完成：成功 ${success_count} 条，失败 ${error_count} 条`);
        if (errors.length > 0) {
          console.error('Import errors:', errors);
        }
      }
      setImportDialogOpen(false);
      setImportFile(null);
    },
    onError: (error: any) => {
      toast.error('规则导入失败', error?.response?.data?.error || '未知错误');
    },
  });

  const handleDelete = (id: number, name: string) => {
    setRuleToDelete({ id, name });
    setDeleteDialogOpen(true);
  };

  const confirmDelete = () => {
    if (ruleToDelete) {
      deleteMutation.mutate(ruleToDelete.id);
    }
  };

  const handleExport = () => {
    exportMutation.mutate();
  };

  const handleImportFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setImportFile(file);
    }
  };

  const handleImport = () => {
    if (!importFile) {
      toast.error('请选择要导入的文件');
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      try {
        const text = e.target?.result as string;
        const data = JSON.parse(text);

        // Handle both direct array and wrapped object formats
        let rules: any[] = [];
        if (Array.isArray(data)) {
          rules = data;
        } else if (data.rules && Array.isArray(data.rules)) {
          rules = data.rules;
        } else {
          toast.error('文件格式错误：无法识别规则数据');
          return;
        }

        if (rules.length === 0) {
          toast.error('文件中没有规则数据');
          return;
        }

        importMutation.mutate(rules);
      } catch (error) {
        toast.error('文件解析失败：' + (error instanceof Error ? error.message : '未知错误'));
      }
    };
    reader.readAsText(importFile);
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-3xl font-bold tracking-tight">规则管理</h2>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleExport} disabled={exportMutation.isPending}>
            <Download className="mr-2 h-4 w-4" />
            {exportMutation.isPending ? '导出中...' : '导出'}
          </Button>
          <Button variant="outline" onClick={() => setImportDialogOpen(true)}>
            <Upload className="mr-2 h-4 w-4" />
            导入
          </Button>
          <Button onClick={() => navigate('/rules/new')}>
            <Plus className="mr-2 h-4 w-4" />
            新建规则
          </Button>
        </div>
      </div>

      {/* Search and Filter */}
      <div className="flex gap-4 mb-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索规则名称、索引模式或描述..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
          {searchQuery && (
            <Button
              variant="ghost"
              size="sm"
              className="absolute right-1 top-1/2 transform -translate-y-1/2 h-6 w-6 p-0"
              onClick={() => setSearchQuery('')}
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>
        <Select value={statusFilter} onValueChange={(value: any) => setStatusFilter(value)}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="筛选状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="enabled">已启用</SelectItem>
            <SelectItem value="disabled">已禁用</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <div className="space-y-4">
          {[1, 2, 3, 4, 5].map((i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      ) : (
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[80px]">ID</TableHead>
                <TableHead>规则名称</TableHead>
                <TableHead>索引模式</TableHead>
                <TableHead>数据源</TableHead>
                <TableHead>Lark配置</TableHead>
                <TableHead>查询间隔</TableHead>
                <TableHead>执行次数</TableHead>
                <TableHead>告警次数</TableHead>
                <TableHead>上次执行</TableHead>
                <TableHead>状态</TableHead>
                <TableHead className="text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredRules.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={11} className="text-center py-8 text-muted-foreground">
                    {searchQuery || statusFilter !== 'all'
                      ? '没有找到匹配的规则'
                      : '暂无规则，点击"新建规则"创建第一个规则'}
                  </TableCell>
                </TableRow>
              ) : (
                filteredRules.map((rule) => (
                  <TableRow key={rule.id}>
                    <TableCell>{rule.id}</TableCell>
                    <TableCell className="font-medium">{rule.name}</TableCell>
                    <TableCell className="font-mono text-sm">{rule.index_pattern}</TableCell>
                    <TableCell>
                      {rule.es_config ? (
                        <div className="flex items-center gap-2">
                          <Badge variant="outline" className="font-normal">
                            {rule.es_config.name}
                          </Badge>
                          {rule.es_config.is_default && (
                            <span className="text-xs text-muted-foreground">(默认)</span>
                          )}
                        </div>
                      ) : (
                        <span className="text-muted-foreground text-sm">-</span>
                      )}
                    </TableCell>
                    <TableCell>
                      {rule.lark_config ? (
                        <div className="flex items-center gap-2">
                          <Badge variant="outline" className="font-normal">
                            {rule.lark_config.name}
                          </Badge>
                          {rule.lark_config.is_default && (
                            <span className="text-xs text-muted-foreground">(默认)</span>
                          )}
                        </div>
                      ) : rule.lark_webhook ? (
                        <Badge variant="outline" className="font-normal text-xs">
                          自定义Webhook
                        </Badge>
                      ) : (
                        <span className="text-muted-foreground text-sm">-</span>
                      )}
                    </TableCell>
                    <TableCell>{rule.interval}秒</TableCell>
                    <TableCell>{rule.run_count}</TableCell>
                    <TableCell>{rule.alert_count}</TableCell>
                    <TableCell className="text-sm">
                      {rule.last_run_time ? (
                        <div className="flex flex-col">
                          <span>{new Date(rule.last_run_time).toLocaleDateString()}</span>
                          <span className="text-xs text-muted-foreground">
                            {new Date(rule.last_run_time).toLocaleTimeString()}
                          </span>
                        </div>
                      ) : (
                        <span className="text-muted-foreground">未执行</span>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Switch
                          checked={rule.enabled}
                          onCheckedChange={() => toggleMutation.mutate(rule.id)}
                          disabled={toggleMutation.isPending}
                        />
                        <Badge variant={rule.enabled ? 'default' : 'secondary'}>
                          {rule.enabled ? '启用' : '禁用'}
                        </Badge>
                      </div>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => navigate(`/rules/${rule.id}/edit`)}
                          title="编辑"
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(rule.id, rule.name)}
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

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              <div className="space-y-2">
                <p>确定要删除规则 <strong>"{ruleToDelete?.name}"</strong> 吗？</p>
                <p className="text-destructive font-medium">⚠️ 此操作为物理删除，数据将永久移除且无法恢复！</p>
                <p className="text-sm text-muted-foreground">建议：如果只是暂时不需要，可以选择"禁用"而非删除。</p>
              </div>
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

      <Dialog open={importDialogOpen} onOpenChange={setImportDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>导入规则</DialogTitle>
            <DialogDescription>
              请选择要导入的规则 JSON 文件。如果规则名称已存在，将会更新现有规则。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <label className="block text-sm font-medium mb-2">选择文件</label>
              <Input
                type="file"
                accept=".json,application/json"
                onChange={handleImportFileChange}
                disabled={importMutation.isPending}
              />
              {importFile && (
                <p className="mt-2 text-sm text-muted-foreground">
                  已选择：{importFile.name}
                </p>
              )}
            </div>
            <div className="text-sm text-muted-foreground">
              <p>支持格式：</p>
              <ul className="list-disc list-inside mt-1 space-y-1">
                <li>规则数组格式：<code className="bg-muted px-1 rounded">[{"{"}...{"}"}]</code></li>
                <li>导出文件格式：<code className="bg-muted px-1 rounded">{"{"}"version": "...", "rules": [...]{"}"}</code></li>
              </ul>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setImportDialogOpen(false);
              setImportFile(null);
            }} disabled={importMutation.isPending}>
              取消
            </Button>
            <Button onClick={handleImport} disabled={importMutation.isPending || !importFile}>
              {importMutation.isPending ? '导入中...' : '确认导入'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
