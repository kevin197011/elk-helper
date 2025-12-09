// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState, useMemo } from 'react';
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
import { Input } from '@/components/ui/input';
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
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery } from '@tanstack/react-query';
import { Eye, Search, X } from 'lucide-react';
import { alertsApi, Alert } from '../services/api';
import { useToast } from '@/contexts/ToastContext';

export default function AlertsPage() {
  const toast = useToast();
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'sent' | 'failed'>('all');

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['alerts', page],
    queryFn: () => alertsApi.getAll(page, pageSize).then(res => res.data),
    staleTime: 30000, // 30 秒内使用缓存，减少重复请求
    gcTime: 5 * 60 * 1000, // 缓存保留 5 分钟
    refetchOnWindowFocus: false, // 窗口聚焦时不自动刷新
  });

  // Filter alerts
  const filteredAlerts = useMemo(() => {
    if (!data?.data) return [];

    let filtered = data.data;

    // Filter by status
    if (statusFilter !== 'all') {
      filtered = filtered.filter(alert => alert.status === statusFilter);
    }

    // Filter by search query
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(alert =>
        alert.rule?.name?.toLowerCase().includes(query) ||
        alert.index_name.toLowerCase().includes(query) ||
        alert.time_range.toLowerCase().includes(query)
      );
    }

    return filtered;
  }, [data?.data, searchQuery, statusFilter]);

  const handleView = async (alert: Alert) => {
    // Fetch full alert details including logs
    try {
      const response = await alertsApi.getById(alert.id);
      setSelectedAlert(response.data.data);
      setDetailDialogOpen(true);
    } catch (error: any) {
      toast.error('获取告警详情失败', error?.response?.data?.error || '未知错误');
      // Fallback to using the alert from list if fetch fails
      setSelectedAlert(alert);
      setDetailDialogOpen(true);
    }
  };

  const totalPages = data?.pagination.total_page || 1;
  const displayAlerts = filteredAlerts.length > 0 ? filteredAlerts : (data?.data || []);

  // Generate page numbers to display (max 10 pages)
  const getPageNumbers = (): (number | 'ellipsis')[] => {
    if (totalPages <= 10) {
      // Show all pages if total is 10 or less
      return Array.from({ length: totalPages }, (_, i) => i + 1);
    }

    const maxVisible = 10;
    const pages: (number | 'ellipsis')[] = [];

    if (page <= 6) {
      // Show first 10 pages
      for (let i = 1; i <= maxVisible; i++) {
        pages.push(i);
      }
      if (totalPages > maxVisible) {
        pages.push('ellipsis');
        pages.push(totalPages);
      }
    } else if (page >= totalPages - 5) {
      // Show last 10 pages
      if (totalPages > maxVisible) {
        pages.push(1);
        pages.push('ellipsis');
      }
      for (let i = totalPages - maxVisible + 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Show pages around current page
      pages.push(1);
      pages.push('ellipsis');
      const start = Math.max(2, page - 3);
      const end = Math.min(totalPages - 1, page + 3);
      for (let i = start; i <= end; i++) {
        pages.push(i);
      }
      if (end < totalPages - 1) {
        pages.push('ellipsis');
      }
      pages.push(totalPages);
    }

    return pages;
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-3xl font-bold tracking-tight">告警历史</h2>
        {isFetching && !isLoading && (
          <div className="flex items-center text-sm text-muted-foreground">
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary mr-2"></div>
            刷新中...
          </div>
        )}
      </div>

      {/* Search and Filter */}
      <div className="flex gap-4 mb-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="搜索规则名称、索引或时间范围..."
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
            <SelectItem value="sent">已发送</SelectItem>
            <SelectItem value="failed">失败</SelectItem>
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
        <>
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[80px]">ID</TableHead>
                  <TableHead>规则名称</TableHead>
                  <TableHead>索引名称</TableHead>
                  <TableHead>日志数量</TableHead>
                  <TableHead>时间范围</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {displayAlerts.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                      {searchQuery || statusFilter !== 'all'
                        ? '没有找到匹配的告警'
                        : '暂无告警记录'}
                    </TableCell>
                  </TableRow>
                ) : (
                  displayAlerts.map((alert) => (
                    <TableRow key={alert.id}>
                      <TableCell>{alert.id}</TableCell>
                      <TableCell className="font-medium">
                        {alert.rule?.name || '-'}
                      </TableCell>
                      <TableCell className="font-mono text-sm">{alert.index_name}</TableCell>
                      <TableCell>{alert.log_count}</TableCell>
                      <TableCell className="text-sm">{alert.time_range}</TableCell>
                      <TableCell>
                        <Badge variant={alert.status === 'sent' ? 'default' : 'destructive'}>
                          {alert.status === 'sent' ? '已发送' : '失败'}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm">
                        {new Date(alert.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleView(alert)}
                          title="查看详情"
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {totalPages > 1 && !searchQuery && statusFilter === 'all' && (
            <div className="mt-4">
              <Pagination>
                <PaginationContent>
                  <PaginationItem>
                    <PaginationPrevious
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      className={page === 1 ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
                    />
                  </PaginationItem>
                  {getPageNumbers().map((p, index) => (
                    <PaginationItem key={p === 'ellipsis' ? `ellipsis-${index}` : p}>
                      {p === 'ellipsis' ? (
                        <PaginationEllipsis />
                      ) : (
                        <PaginationLink
                          onClick={() => setPage(p)}
                          isActive={p === page}
                          className="cursor-pointer"
                        >
                          {p}
                        </PaginationLink>
                      )}
                    </PaginationItem>
                  ))}
                  <PaginationItem>
                    <PaginationNext
                      onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                      className={page === totalPages ? 'pointer-events-none opacity-50' : 'cursor-pointer'}
                    />
                  </PaginationItem>
                </PaginationContent>
              </Pagination>
            </div>
          )}
        </>
      )}

      {/* Alert Detail Dialog */}
      <Dialog open={detailDialogOpen} onOpenChange={setDetailDialogOpen}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>告警详情 #{selectedAlert?.id}</DialogTitle>
            <DialogDescription>
              {selectedAlert?.rule?.name || '规则名称'}
            </DialogDescription>
          </DialogHeader>
          {selectedAlert && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">索引名称</p>
                  <p className="font-mono">{selectedAlert.index_name}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">日志数量</p>
                  <p>{selectedAlert.log_count}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">时间范围</p>
                  <p>{selectedAlert.time_range}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">状态</p>
                  <Badge variant={selectedAlert.status === 'sent' ? 'default' : 'destructive'}>
                    {selectedAlert.status === 'sent' ? '已发送' : '失败'}
                  </Badge>
                </div>
                {selectedAlert.error_msg && (
                  <div className="col-span-2">
                    <p className="text-sm font-medium text-muted-foreground">错误信息</p>
                    <p className="text-sm text-destructive">{selectedAlert.error_msg}</p>
                  </div>
                )}
              </div>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <p className="text-sm font-medium text-muted-foreground">
                    日志详情
                    {selectedAlert.log_count > 0 && (
                      <span className="ml-2">
                        (共 {selectedAlert.log_count} 条，显示前 10 条)
                      </span>
                    )}
                  </p>
                  {selectedAlert.logs && selectedAlert.logs.length > 0 && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        const jsonStr = JSON.stringify(selectedAlert.logs, null, 2);
                        navigator.clipboard.writeText(jsonStr);
                        toast.success('已复制到剪贴板');
                      }}
                    >
                      📋 复制 JSON
                    </Button>
                  )}
                </div>
                {selectedAlert.logs && Array.isArray(selectedAlert.logs) && selectedAlert.logs.length > 0 ? (
                  <>
                    <pre className="bg-muted p-4 rounded-md overflow-auto max-h-96 text-xs font-mono">
                      {JSON.stringify(selectedAlert.logs, null, 2)}
                    </pre>
                    {selectedAlert.log_count > 10 && (
                      <p className="text-xs text-muted-foreground mt-2">
                        💡 为了性能考虑，详情仅显示前 10 条日志。完整日志已发送到 Lark/飞书通知。
                      </p>
                    )}
                  </>
                ) : (
                  <div className="bg-muted p-4 rounded-md text-sm text-muted-foreground">
                    {selectedAlert.log_count > 0
                      ? '日志数据为空，可能已被清理或数据格式异常'
                      : '暂无日志数据'}
                  </div>
                )}
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDetailDialogOpen(false)}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
