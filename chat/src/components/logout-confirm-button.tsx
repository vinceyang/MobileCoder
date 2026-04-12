'use client';

import { useRouter } from 'next/navigation';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';

interface LogoutConfirmButtonProps {
  children: React.ReactNode;
  className?: string;
}

export function LogoutConfirmButton({ children, className }: LogoutConfirmButtonProps) {
  const router = useRouter();

  const handleConfirm = () => {
    localStorage.clear();
    router.push('/login');
  };

  return (
    <Dialog>
      <DialogTrigger asChild>
        <button type="button" className={className}>
          {children}
        </button>
      </DialogTrigger>
      <DialogContent className="w-[calc(100vw-32px)] max-w-sm rounded-[28px] border border-cyan-400/10 bg-slate-950 p-5 text-slate-100 shadow-[0_24px_90px_rgba(0,0,0,0.55)]">
        <DialogHeader className="text-left">
          <DialogTitle className="text-xl font-black tracking-tight text-slate-50">确认退出登录？</DialogTitle>
          <DialogDescription className="text-sm leading-6 text-slate-400">
            退出后需要重新输入账号密码才能继续查看任务和终端。
          </DialogDescription>
        </DialogHeader>
        <DialogFooter className="mt-2 flex flex-col-reverse gap-2 sm:flex-row">
          <DialogTrigger asChild>
            <button className="rounded-2xl border border-cyan-400/10 bg-slate-900 px-4 py-3 text-sm font-semibold text-slate-200">
              取消
            </button>
          </DialogTrigger>
          <button
            onClick={handleConfirm}
            className="rounded-2xl bg-rose-400 px-4 py-3 text-sm font-black text-slate-950"
          >
            确认退出
          </button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
