// Copyright (c) 2025 kk
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { authApi } from '../services/api';
import { useToast } from '@/contexts/ToastContext';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './ui/dialog';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Loader2, Lock } from 'lucide-react';

interface ChangePasswordDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface ChangePasswordForm {
  old_password: string;
  new_password: string;
  confirm_password: string;
}

export default function ChangePasswordDialog({ open, onOpenChange }: ChangePasswordDialogProps) {
  const { toast } = useToast();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const form = useForm<ChangePasswordForm>({
    defaultValues: {
      old_password: '',
      new_password: '',
      confirm_password: '',
    },
  });

  const onSubmit = async (data: ChangePasswordForm) => {
    // Validate password confirmation
    if (data.new_password !== data.confirm_password) {
      toast({
        title: '密码确认失败',
        description: '新密码和确认密码不一致',
        variant: 'error',
      });
      return;
    }

    setIsSubmitting(true);
    try {
      await authApi.updatePassword(data.old_password, data.new_password);
      toast({
        title: '密码修改成功',
        description: '您的密码已成功更新',
      });
      form.reset();
      onOpenChange(false);
    } catch (error: any) {
      toast({
        title: '密码修改失败',
        description: error.response?.data?.error || '修改密码时发生错误',
        variant: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>修改密码</DialogTitle>
          <DialogDescription>
            请输入当前密码和新密码来更新您的账户密码
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="old_password">当前密码</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                id="old_password"
                type="password"
                placeholder="请输入当前密码"
                className="pl-10"
                {...form.register('old_password', {
                  required: '请输入当前密码',
                })}
                disabled={isSubmitting}
              />
            </div>
            {form.formState.errors.old_password && (
              <p className="text-sm text-red-600">{form.formState.errors.old_password.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="new_password">新密码</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                id="new_password"
                type="password"
                placeholder="请输入新密码（至少6个字符）"
                className="pl-10"
                {...form.register('new_password', {
                  required: '请输入新密码',
                  minLength: {
                    value: 6,
                    message: '新密码长度至少为 6 个字符',
                  },
                })}
                disabled={isSubmitting}
              />
            </div>
            <p className="text-sm text-muted-foreground">密码长度至少为 6 个字符</p>
            {form.formState.errors.new_password && (
              <p className="text-sm text-red-600">{form.formState.errors.new_password.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirm_password">确认新密码</Label>
            <div className="relative">
              <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                id="confirm_password"
                type="password"
                placeholder="请再次输入新密码"
                className="pl-10"
                {...form.register('confirm_password', {
                  required: '请确认新密码',
                })}
                disabled={isSubmitting}
              />
            </div>
            {form.formState.errors.confirm_password && (
              <p className="text-sm text-red-600">{form.formState.errors.confirm_password.message}</p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => {
                form.reset();
                onOpenChange(false);
              }}
              disabled={isSubmitting}
            >
              取消
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  修改中...
                </>
              ) : (
                '确认修改'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

