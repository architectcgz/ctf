<script setup lang="ts">
import { useRouter } from 'vue-router'
import ChallengePackageFormatGuidePanel from '@/components/platform/challenge/ChallengePackageFormatGuidePanel.vue'

const router = useRouter()
</script>

<template>
  <section
    class="workspace-shell journal-shell journal-shell-admin journal-hero flex min-h-full flex-1 flex-col"
  >
    <header class="workspace-topbar">
      <div class="topbar-leading">
        <span class="workspace-overline">Challenge Package</span>
        <span class="class-chip">上传示例</span>
      </div>
      <button
        class="nav-back"
        type="button"
        @click="router.push({ name: 'PlatformChallengeImportManage' })"
      >
        返回导入题目包
      </button>
    </header>

    <ChallengePackageFormatGuidePanel />
  </section>
</template>

<style scoped>
.journal-shell {
  --journal-divider-margin-block: 1.5rem;
  --journal-shell-hero-radial-strength: 7%;
  --journal-shell-hero-radial-size: 22rem;
  --journal-shell-hero-end: var(--journal-surface);
  --journal-shell-hero-shadow: 0 22px 50px var(--color-shadow-soft);
}

.workspace-overline {
  font-size: var(--font-size-0-70);
  font-weight: 700;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--journal-accent);
}

.nav-back {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.35rem;
  border-radius: 0.75rem;
  border: 1px solid var(--journal-border);
  padding: var(--space-2) var(--space-3-5);
  background: color-mix(in srgb, var(--journal-surface) 96%, var(--color-bg-base));
  color: var(--journal-ink);
  font-size: var(--font-size-0-88);
  font-weight: 600;
}

</style>
