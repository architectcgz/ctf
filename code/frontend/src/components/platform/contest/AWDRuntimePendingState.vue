<script setup lang="ts">
import { Activity, Clock, ShieldAlert, Sword, TimerReset, ClipboardCheck } from 'lucide-vue-next'
</script>

<template>
  <section class="studio-pending-shell">
    <div class="pending-identity">
      <div class="pending-overline">
        Operational Lifecycle
      </div>
      <h2 class="pending-title">
        尚未进入运行阶段
      </h2>
      <p class="pending-description">
        当前赛事尚未开启轮次调度。请先确保所有题目在“赛前就绪检查”中校验通过，随后在“题目编排”或“基础信息”中将赛事切换为“进行中”状态。
      </p>
    </div>

    <div class="pending-grid">
      <div class="pending-card">
        <div class="pending-card__icon">
          <TimerReset class="h-5 w-5" />
        </div>
        <div class="pending-card__body">
          <div class="card-label">
            轮次调度
          </div>
          <div class="card-status">
            等待开赛...
          </div>
        </div>
      </div>
      <div class="pending-card">
        <div class="pending-card__icon">
          <Sword class="h-5 w-5" />
        </div>
        <div class="pending-card__body">
          <div class="card-label">
            攻击流水
          </div>
          <div class="card-status">
            链路关闭
          </div>
        </div>
      </div>
      <div class="pending-card">
        <div class="pending-card__icon">
          <ShieldAlert class="h-5 w-5" />
        </div>
        <div class="pending-card__body">
          <div class="card-label">
            服务巡检
          </div>
          <div class="card-status">
            尚未就绪
          </div>
        </div>
      </div>
    </div>

    <div class="pending-actions">
      <div class="flex items-center gap-3">
        <button id="awd-runtime-shell-create-round" type="button" class="ui-btn ui-btn--secondary" disabled>
          <TimerReset class="h-3.5 w-3.5" /> 立即同步轮次
        </button>
        <button id="awd-runtime-shell-run-check" type="button" class="ui-btn ui-btn--primary" disabled>
          <ShieldAlert class="h-3.5 w-3.5" /> 补录检查
        </button>
      </div>
      <p class="pending-footer-hint">
        需先通过赛前检查并开赛，本面板才会切换到实时监控模式。
      </p>
    </div>
  </section>
</template>

<style scoped>
.studio-pending-shell {
  padding: 4rem 3rem;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 3rem;
}

.pending-overline { font-size: 10px; font-weight: 800; text-transform: uppercase; color: #3b82f6; letter-spacing: 0.2em; margin-bottom: 0.75rem; }
.pending-title { font-size: 2rem; font-weight: 900; color: #0f172a; margin: 0; }
.pending-description { font-size: 14px; color: #64748b; max-width: 36rem; line-height: 1.7; margin: 1rem auto 0; }

.pending-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
  width: 100%;
  max-width: 48rem;
}

.pending-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 1rem;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.pending-card__icon {
  width: 3rem;
  height: 3rem;
  border-radius: 0.75rem;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
  border: 1px solid #e2e8f0;
}

.card-label { font-size: 10px; font-weight: 800; text-transform: uppercase; color: #94a3b8; letter-spacing: 0.05em; }
.card-status { font-size: 14px; font-weight: 800; color: #475569; }

.pending-actions { display: flex; flex-direction: column; align-items: center; gap: 1rem; }

.ops-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  height: 2.5rem;
  padding: 0 1.25rem;
  border-radius: 0.85rem;
  font-size: 13px;
  font-weight: 700;
}

.ops-btn--disabled { background: #f1f5f9; color: #cbd5e1; cursor: not-allowed; border: 1px solid #e2e8f0; }

.pending-footer-hint { font-size: 12px; font-weight: 600; color: #94a3b8; }
</style>
