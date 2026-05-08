<template>
  <main class="page">
    <!-- blurred background dashboard -->
    <section class="dashboard-bg" aria-hidden="true">
      <header class="mock-topbar">
        <div class="mock-logo"></div>
        <div class="mock-title"></div>
        <div class="mock-actions">
          <span></span>
          <span></span>
          <span></span>
        </div>
      </header>

      <section class="mock-dashboard">
        <div class="mock-card card-large">
          <div class="mock-card-title"></div>
          <div class="mock-line-chart">
            <svg viewBox="0 0 320 120" class="chart-svg">
              <path
                d="M12 96 C42 70, 62 78, 91 58 S142 48, 171 36 S214 62, 244 43 S286 30, 310 46"
                fill="none"
                stroke="rgba(38, 137, 255, .42)"
                stroke-width="5"
                stroke-linecap="round"
              />
              <path
                d="M12 98 C42 72, 62 80, 91 60 S142 50, 171 38 S214 64, 244 45 S286 32, 310 48 L310 118 L12 118 Z"
                fill="url(#lineFill)"
              />
              <defs>
                <linearGradient id="lineFill" x1="0" y1="0" x2="0" y2="1">
                  <stop stop-color="rgba(31, 134, 255, .2)" />
                  <stop offset="1" stop-color="rgba(31, 134, 255, 0)" />
                </linearGradient>
              </defs>
            </svg>
          </div>
        </div>

        <div class="mock-list-card">
          <div class="mock-list-row" v-for="row in 5" :key="row">
            <span class="mock-dot"></span>
            <span class="mock-line"></span>
            <span class="mock-short"></span>
          </div>
        </div>

        <div class="mock-ring-card">
          <div class="mock-ring"></div>
        </div>

        <div class="mock-side-list">
          <div class="mock-side-item yellow"></div>
          <div class="mock-side-item green"></div>
        </div>
      </section>
    </section>

    <!-- full height notification center -->
    <aside class="notification-panel">
      <div class="panel-inner">
        <header class="panel-header">
          <div class="title-wrap">
            <div class="bell-wrap">
              <svg
                class="bell-icon"
                viewBox="0 0 32 32"
                aria-hidden="true"
              >
                <path
                  d="M16 27.2c2.1 0 3.5-1.1 4-2.8h-8c.5 1.7 1.9 2.8 4 2.8Z"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                />
                <path
                  d="M7.8 22.7h16.4c1 0 1.6-1.1 1-2-1.4-2-2.2-3.2-2.2-7.4 0-4.4-2.8-7.6-7-7.6s-7 3.2-7 7.6c0 4.2-.8 5.4-2.2 7.4-.6.9 0 2 1 2Z"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linejoin="round"
                />
                <path
                  d="M14.7 5.8V4.7c0-.8.6-1.4 1.3-1.4s1.3.6 1.3 1.4v1.1"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.8"
                  stroke-linecap="round"
                />
              </svg>
              <span class="bell-dot"></span>
            </div>

            <div>
              <h1>NOTIFICATIONS</h1>
              <p>通知中心</p>
            </div>
          </div>

          <button class="close-btn" type="button" aria-label="关闭通知中心">
            <svg viewBox="0 0 28 28" aria-hidden="true">
              <path
                d="M6.5 6.5 21.5 21.5M21.5 6.5 6.5 21.5"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
              />
            </svg>
          </button>
        </header>

        <section class="summary-row">
          <div class="summary-main">
            <span class="summary-number">1</span>
            <span class="summary-text">条未读通知待处理</span>
          </div>

          <nav class="summary-actions" aria-label="通知操作">
            <button class="text-action" type="button">实时同步</button>
            <span class="action-separator"></span>
            <button class="text-action" type="button">全部设为已读</button>
          </nav>
        </section>

        <section class="tabs" aria-label="通知筛选">
          <button
            v-for="tab in tabs"
            :key="tab.value"
            type="button"
            class="tab-btn"
            :class="{ 'is-active': activeTab === tab.value }"
          >
            {{ tab.label }}
          </button>
        </section>

        <div class="content-divider"></div>

        <section class="notification-list" aria-label="通知列表">
          <article
            v-for="item in notifications"
            :key="item.id"
            class="notice-card"
            :class="{ 'is-unread': item.unread, 'is-read': !item.unread }"
          >
            <div class="notice-icon">
              <svg viewBox="0 0 28 28" aria-hidden="true">
                <path
                  d="M8 23V6.5"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                />
                <path
                  d="M8 7.4c3.8-2.1 6.5 1.9 10.8-.3.7-.4 1.6.1 1.6.9v8.1c0 .5-.3 1-.8 1.2-4.3 1.9-7-1.7-11.6.4V7.4Z"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linejoin="round"
                />
              </svg>
            </div>

            <div class="notice-body">
              <div class="notice-category">{{ item.category }}</div>

              <div class="notice-title-row">
                <h2>{{ item.title }}</h2>
                <time>{{ item.time }}</time>
              </div>

              <p>{{ item.message }}</p>
            </div>

            <span
              v-if="item.unread"
              class="unread-dot"
              aria-label="未读"
            ></span>
          </article>
        </section>
      </div>

      <footer class="panel-footer">
        <button class="view-all-btn" type="button">
          <span class="footer-icon">
            <svg viewBox="0 0 28 28" aria-hidden="true">
              <path
                d="M10 8h12M10 14h12M10 20h12"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
              />
              <path
                d="M5.8 8h.1M5.8 14h.1M5.8 20h.1"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                stroke-linecap="round"
              />
            </svg>
          </span>

          <span>查看全部通知</span>

          <svg class="arrow-icon" viewBox="0 0 28 28" aria-hidden="true">
            <path
              d="m10.5 5.8 8 8.2-8 8.2"
              fill="none"
              stroke="currentColor"
              stroke-width="2.4"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </button>
      </footer>
    </aside>
  </main>
</template>

<script setup>
import { ref } from 'vue'

const activeTab = ref('all')

const tabs = [
  {
    label: '全部',
    value: 'all'
  },
  {
    label: '未读',
    value: 'unread'
  },
  {
    label: '已读',
    value: 'read'
  }
]

const notifications = [
  {
    id: 1,
    category: '训练',
    title: '题目解出',
    time: '2026/4/28 13:53:24',
    message: '你已成功提交题目 #11 的 Flag，获得 100 分。',
    unread: true
  },
  {
    id: 2,
    category: '训练',
    title: '题目解出',
    time: '2026/4/28 11:46:51',
    message: '你已成功提交题目 #6 的 Flag，获得 100 分。',
    unread: false
  }
]
</script>

<style scoped>
* {
  box-sizing: border-box;
}

.page {
  position: relative;
  width: 100vw;
  min-height: 100vh;
  overflow: hidden;
  color: #f4f7fb;
  background:
    radial-gradient(circle at 18% 22%, rgba(25, 75, 118, 0.22), transparent 36%),
    radial-gradient(circle at 85% 20%, rgba(25, 82, 129, 0.18), transparent 38%),
    linear-gradient(135deg, #06101b 0%, #09131f 44%, #0c1621 100%);
  font-family:
    Inter,
    ui-sans-serif,
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    "PingFang SC",
    "Microsoft YaHei",
    sans-serif;
}

/* =========================
   blurred left dashboard
   ========================= */

.dashboard-bg {
  position: absolute;
  inset: 0;
  padding: 74px 58vw 80px 52px;
  opacity: 0.58;
  filter: blur(7px);
  transform: scale(1.015);
  transform-origin: left center;
  pointer-events: none;
}

.dashboard-bg::before {
  content: "";
  position: absolute;
  inset: 0;
  background:
    linear-gradient(90deg, rgba(4, 10, 18, 0.3), rgba(5, 12, 21, 0.62) 52%, rgba(3, 9, 17, 0.86)),
    radial-gradient(circle at 26% 34%, rgba(46, 123, 214, 0.16), transparent 34%),
    radial-gradient(circle at 23% 76%, rgba(61, 224, 126, 0.12), transparent 26%);
  z-index: -1;
}

.mock-topbar {
  display: flex;
  align-items: center;
  height: 54px;
  gap: 18px;
  opacity: 0.45;
}

.mock-logo {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  background:
    linear-gradient(135deg, rgba(57, 169, 255, 0.35), rgba(21, 38, 59, 0.2)),
    rgba(20, 34, 50, 0.72);
  border: 1px solid rgba(100, 151, 210, 0.2);
}

.mock-title {
  width: 110px;
  height: 22px;
  border-radius: 8px;
  background: rgba(166, 187, 210, 0.16);
}

.mock-actions {
  margin-left: auto;
  display: flex;
  gap: 20px;
}

.mock-actions span {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(173, 190, 214, 0.14);
}

.mock-dashboard {
  margin-top: 138px;
  display: grid;
  gap: 28px;
  width: 420px;
}

.mock-card,
.mock-list-card,
.mock-ring-card {
  border-radius: 28px;
  background: rgba(11, 23, 36, 0.64);
  border: 1px solid rgba(113, 148, 185, 0.13);
  box-shadow: 0 24px 70px rgba(0, 0, 0, 0.28);
}

.card-large {
  height: 210px;
  padding: 34px;
}

.mock-card-title {
  width: 150px;
  height: 26px;
  border-radius: 12px;
  background: rgba(205, 218, 236, 0.17);
}

.mock-line-chart {
  margin-top: 28px;
  height: 118px;
}

.chart-svg {
  width: 100%;
  height: 100%;
}

.mock-list-card {
  height: 240px;
  padding: 34px 30px;
}

.mock-list-row {
  display: flex;
  align-items: center;
  gap: 18px;
  height: 36px;
  margin-bottom: 10px;
}

.mock-dot {
  width: 14px;
  height: 14px;
  border-radius: 4px;
  background: rgba(53, 230, 127, 0.32);
  box-shadow: 0 0 22px rgba(53, 230, 127, 0.2);
}

.mock-line {
  width: 190px;
  height: 16px;
  border-radius: 8px;
  background: rgba(190, 205, 225, 0.15);
}

.mock-short {
  width: 64px;
  height: 14px;
  border-radius: 7px;
  background: rgba(80, 150, 239, 0.18);
}

.mock-ring-card {
  height: 260px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mock-ring {
  width: 168px;
  height: 168px;
  border-radius: 50%;
  background:
    radial-gradient(circle, #07111c 0 56%, transparent 57%),
    conic-gradient(rgba(51, 224, 133, 0.54) 0 32%, rgba(57, 149, 255, 0.28) 32% 74%, rgba(47, 66, 88, 0.3) 74% 100%);
  box-shadow: 0 0 38px rgba(35, 122, 226, 0.12);
}

.mock-side-list {
  display: grid;
  gap: 52px;
  padding-left: 10px;
}

.mock-side-item {
  width: 38px;
  height: 38px;
  border-radius: 50%;
}

.mock-side-item.yellow {
  background: rgba(235, 192, 53, 0.34);
  box-shadow: 0 0 34px rgba(235, 192, 53, 0.28);
}

.mock-side-item.green {
  background: rgba(61, 227, 142, 0.24);
  box-shadow: 0 0 34px rgba(61, 227, 142, 0.24);
}

/* =========================
   right notification panel
   ========================= */

.notification-panel {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  width: 53.4vw;
  min-width: 500px;
  max-width: 540px;
  display: flex;
  flex-direction: column;
  background:
    radial-gradient(circle at 38% 4%, rgba(68, 122, 184, 0.14), transparent 28%),
    radial-gradient(circle at 88% 22%, rgba(31, 93, 156, 0.12), transparent 36%),
    linear-gradient(180deg, rgba(14, 23, 34, 0.98), rgba(9, 18, 29, 0.99));
  border-left: 1px solid rgba(151, 173, 202, 0.24);
  box-shadow:
    -28px 0 82px rgba(0, 0, 0, 0.34),
    inset 1px 0 0 rgba(255, 255, 255, 0.03);
}

.notification-panel::before {
  content: "";
  position: absolute;
  inset: 0;
  background:
    linear-gradient(90deg, rgba(255, 255, 255, 0.028), transparent 28%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.025), transparent 24%);
  pointer-events: none;
}

.panel-inner {
  position: relative;
  z-index: 1;
  flex: 1;
  padding: 101px 40px 180px 28px;
  overflow: hidden;
}

.panel-header {
  position: relative;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  min-height: 70px;
}

.title-wrap {
  display: flex;
  align-items: flex-start;
  gap: 22px;
}

.bell-wrap {
  position: relative;
  width: 50px;
  height: 50px;
  color: rgba(238, 244, 252, 0.93);
  display: flex;
  align-items: center;
  justify-content: center;
}

.bell-icon {
  width: 36px;
  height: 36px;
  filter: drop-shadow(0 0 7px rgba(90, 163, 255, 0.12));
}

.bell-dot {
  position: absolute;
  top: 5px;
  right: 7px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #42a5ff;
  box-shadow: 0 0 14px rgba(66, 165, 255, 0.7);
}

.panel-header h1 {
  margin: 0;
  color: #f4f7fb;
  font-size: 24px;
  line-height: 1.18;
  font-weight: 500;
  letter-spacing: 4px;
}

.panel-header p {
  margin: 12px 0 0;
  color: rgba(200, 210, 225, 0.72);
  font-size: 16px;
  line-height: 1.2;
  letter-spacing: 0.2px;
}

.close-btn {
  position: relative;
  top: 3px;
  width: 38px;
  height: 38px;
  padding: 0;
  color: rgba(231, 238, 249, 0.9);
  border: none;
  border-radius: 12px;
  background: transparent;
  cursor: pointer;
}

.close-btn svg {
  width: 30px;
  height: 30px;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.06);
}

.summary-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 22px;
  margin-top: 93px;
  min-height: 48px;
}

.summary-main {
  display: flex;
  align-items: baseline;
  gap: 12px;
  white-space: nowrap;
}

.summary-number {
  color: #45a6ff;
  font-size: 45px;
  line-height: 1;
  font-weight: 300;
  letter-spacing: -1px;
  text-shadow: 0 0 18px rgba(69, 166, 255, 0.32);
}

.summary-text {
  color: rgba(245, 248, 252, 0.94);
  font-size: 16px;
  line-height: 1;
  font-weight: 500;
}

.summary-actions {
  display: flex;
  align-items: center;
  gap: 16px;
  white-space: nowrap;
}

.text-action {
  padding: 0;
  border: 0;
  background: transparent;
  color: rgba(112, 173, 247, 0.82);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

.text-action:hover {
  color: #72b8ff;
}

.action-separator {
  width: 1px;
  height: 21px;
  background: rgba(152, 166, 184, 0.27);
}

.tabs {
  margin-top: 49px;
  height: 46px;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  align-items: center;
  column-gap: 12px;
}

.tab-btn {
  height: 45px;
  border: 0;
  border-radius: 18px;
  background: transparent;
  color: rgba(229, 236, 247, 0.88);
  font-size: 22px;
  font-weight: 400;
  cursor: pointer;
}

.tab-btn.is-active {
  color: #ffffff;
  font-weight: 600;
  background:
    linear-gradient(180deg, rgba(41, 54, 72, 0.92), rgba(31, 42, 58, 0.78));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.045),
    0 12px 28px rgba(0, 0, 0, 0.18);
}

.content-divider {
  height: 1px;
  margin-top: 20px;
  background: linear-gradient(90deg, transparent, rgba(150, 166, 188, 0.26) 8%, rgba(150, 166, 188, 0.22) 92%, transparent);
}

.notification-list {
  display: grid;
  gap: 16px;
  margin-top: 28px;
}

.notice-card {
  position: relative;
  min-height: 164px;
  display: grid;
  grid-template-columns: 74px 1fr;
  gap: 18px;
  padding: 27px 56px 24px 31px;
  border: 1px solid rgba(139, 158, 181, 0.17);
  border-radius: 22px;
  background:
    radial-gradient(circle at 16% 30%, rgba(70, 104, 135, 0.14), transparent 34%),
    linear-gradient(180deg, rgba(18, 31, 45, 0.66), rgba(13, 23, 35, 0.62));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.025),
    0 16px 42px rgba(0, 0, 0, 0.16);
}

.notice-card.is-unread {
  border-color: rgba(143, 167, 199, 0.21);
}

.notice-card.is-read {
  opacity: 0.95;
}

.notice-icon {
  width: 57px;
  height: 57px;
  margin-top: 2px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #61f083;
  background:
    radial-gradient(circle at 48% 43%, rgba(81, 240, 130, 0.2), rgba(58, 188, 101, 0.14) 52%, rgba(58, 188, 101, 0.09));
  box-shadow:
    0 0 28px rgba(67, 225, 121, 0.09),
    inset 0 0 0 1px rgba(76, 230, 126, 0.04);
}

.notice-icon svg {
  width: 31px;
  height: 31px;
}

.notice-body {
  min-width: 0;
}

.notice-category {
  margin-bottom: 7px;
  color: #3be26f;
  font-size: 15px;
  line-height: 1;
  font-weight: 600;
}

.notice-title-row {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
}

.notice-title-row h2 {
  margin: 0;
  flex: 0 0 auto;
  color: #f9fbff;
  font-size: 24px;
  line-height: 1.18;
  font-weight: 700;
  letter-spacing: 0.2px;
}

.notice-title-row time {
  min-width: 0;
  color: rgba(188, 199, 213, 0.78);
  font-size: 16px;
  line-height: 1.1;
  font-weight: 400;
  white-space: nowrap;
}

.notice-body p {
  width: 100%;
  max-width: 330px;
  margin: 18px 0 0;
  color: rgba(217, 225, 237, 0.84);
  font-size: 18px;
  line-height: 1.58;
  font-weight: 400;
  letter-spacing: 0.1px;
}

.unread-dot {
  position: absolute;
  top: 40px;
  right: 31px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #3ea7ff;
  box-shadow:
    0 0 18px rgba(62, 167, 255, 0.72),
    0 0 2px rgba(255, 255, 255, 0.4) inset;
}

.panel-footer {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 53.4vw;
  min-width: 500px;
  max-width: 540px;
  height: 162px;
  z-index: 2;
  border-top: 1px solid rgba(141, 158, 179, 0.14);
  background:
    linear-gradient(180deg, rgba(10, 19, 30, 0.88), rgba(8, 17, 27, 0.96));
  box-shadow: 0 -18px 38px rgba(0, 0, 0, 0.16);
}

.view-all-btn {
  width: 100%;
  height: 100%;
  display: grid;
  grid-template-columns: 58px 1fr 40px;
  align-items: center;
  column-gap: 22px;
  padding: 0 35px 0 33px;
  border: 0;
  background: transparent;
  color: #f1f6fd;
  cursor: pointer;
  text-align: left;
}

.footer-icon {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  color: rgba(229, 238, 249, 0.92);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 25, 39, 0.62);
  border: 1px solid rgba(138, 158, 184, 0.16);
}

.footer-icon svg {
  width: 28px;
  height: 28px;
}

.view-all-btn span:nth-child(2) {
  font-size: 22px;
  line-height: 1;
  font-weight: 500;
  letter-spacing: 0.1px;
}

.arrow-icon {
  justify-self: end;
  width: 34px;
  height: 34px;
  color: rgba(239, 246, 255, 0.9);
}

@media (max-width: 900px) {
  .notification-panel,
  .panel-footer {
    width: 56vw;
    min-width: 480px;
  }

  .panel-inner {
    padding-left: 26px;
    padding-right: 32px;
  }

  .summary-row {
    gap: 14px;
  }

  .summary-actions {
    gap: 12px;
  }

  .notice-card {
    padding-right: 48px;
  }
}

@media (max-width: 760px) {
  .dashboard-bg {
    display: none;
  }

  .notification-panel,
  .panel-footer {
    width: 100vw;
    min-width: 0;
    max-width: none;
  }

  .panel-inner {
    padding: 76px 24px 160px;
  }

  .summary-row {
    align-items: flex-start;
    flex-direction: column;
    margin-top: 64px;
  }

  .tabs {
    margin-top: 36px;
  }

  .notice-card {
    grid-template-columns: 56px 1fr;
    gap: 14px;
    min-height: 150px;
    padding: 24px 42px 22px 22px;
  }

  .notice-icon {
    width: 48px;
    height: 48px;
  }

  .notice-title-row h2 {
    font-size: 21px;
  }

  .notice-title-row time {
    font-size: 13px;
  }

  .notice-body p {
    font-size: 16px;
  }
}
</style>