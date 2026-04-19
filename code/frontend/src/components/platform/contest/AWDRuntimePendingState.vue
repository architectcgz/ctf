<template>
  <section class="awd-runtime-shell border p-6">
    <div class="awd-runtime-shell-overline flex flex-wrap items-center gap-2">
      <span>Operations</span>
      <span class="awd-runtime-shell-chip rounded-full px-2 py-1">待开赛</span>
    </div>
    <div class="mt-3 grid gap-4">
      <div>
        <h2 class="text-3xl font-semibold tracking-tight text-white">尚未进入运行阶段</h2>
        <p class="awd-runtime-shell-description mt-3 text-sm leading-7">
          当前赛事还不能进入轮次运行。需先通过赛前检查并开赛，随后才会接管创建轮次、服务巡检、攻击补录和当前轮态势。
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-3">
        <button
          id="awd-runtime-shell-create-round"
          type="button"
          class="ui-btn ui-btn--secondary"
          disabled
        >
          创建轮次
        </button>
        <button
          id="awd-runtime-shell-record-service"
          type="button"
          class="ui-btn ui-btn--secondary"
          disabled
        >
          录入服务检查
        </button>
        <button
          id="awd-runtime-shell-record-attack"
          type="button"
          class="ui-btn ui-btn--secondary"
          disabled
        >
          补录攻击日志
        </button>
        <button
          id="awd-runtime-shell-run-check"
          type="button"
          class="ui-btn ui-btn--primary"
          disabled
        >
          立即巡检当前轮
        </button>
      </div>

      <p class="awd-runtime-shell-hint text-xs">需先通过赛前检查并开赛</p>
    </div>
  </section>
</template>

<style scoped>
.awd-runtime-shell {
  border-radius: 1.75rem;
  box-shadow: 0 24px 70px var(--color-shadow-soft);
  background:
    linear-gradient(
      145deg,
      color-mix(in srgb, var(--color-surface-panel) 94%, white 6%),
      var(--color-surface-panel)
    ),
    radial-gradient(
      circle at top right,
      color-mix(in srgb, var(--color-primary) 18%, transparent),
      transparent 52%
    );
}

.awd-runtime-shell-overline {
  font-size: var(--font-size-11);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.22em;
  color: color-mix(in srgb, var(--color-primary-hover) 75%, transparent);
}

.awd-runtime-shell-chip {
  background: color-mix(in srgb, var(--color-warning) 18%, transparent);
  color: color-mix(in srgb, var(--color-warning) 74%, white 26%);
}

.awd-runtime-shell-description {
  color: color-mix(in srgb, var(--color-text-secondary) 90%, transparent);
}

.awd-runtime-shell-hint {
  color: color-mix(in srgb, var(--color-primary-hover) 75%, transparent);
}
</style>
