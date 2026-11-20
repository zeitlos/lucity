<script setup lang="ts">
const github = 'M12 .3a12 12 0 0 0-3.8 23.4c.6.1.8-.3.8-.6v-2c-3.3.7-4-1.6-4-1.6-.6-1.4-1.4-1.8-1.4-1.8-1-.7.1-.7.1-.7 1.2.1 1.8 1.2 1.8 1.2 1 1.8 2.8 1.3 3.5 1 0-.8.4-1.3.7-1.6-2.7-.3-5.5-1.3-5.5-6 0-1.2.5-2.3 1.3-3.1-.2-.4-.6-1.6 0-3.2 0 0 1-.3 3.4 1.2a11.5 11.5 0 0 1 6 0c2.3-1.5 3.3-1.2 3.3-1.2.6 1.6.2 2.8.1 3.2.8.8 1.3 1.9 1.3 3.2 0 4.6-2.8 5.6-5.5 5.9.5.4.9 1 .9 2.2v3.3c0 .3.1.7.8.6A12 12 0 0 0 12 .3';
</script>

<template>
  <div class="ag">
    <div class="ag-note">
      <h3 class="font-display text-2xl/tight text-stone-900 dark:text-stone-100">Your agent deploys. You stay in charge.</h3>
      <p class="mt-2.5 text-base text-[oklch(0.36_0.04_300)] dark:text-[oklch(0.85_0.04_300)]">
        The plugin works with Claude Code, Codex, Cursor and any other MCP client. It asks before it creates anything paid, and it sees the names of your variables, never their values.
      </p>
      <p class="mt-3 text-base">
        <NuxtLink to="/docs/claude-plugin" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">Set up the plugin</NuxtLink>
      </p>
    </div>

    <div class="ag-chat" role="img" aria-label="You ask your agent to deploy the project. It sets up a service, then asks before it creates a paid database. After your yes, the database appears on the canvas next to the service, and the app goes live.">
      <div class="ag-you ag-chip" aria-hidden="true">deploy this to Lucity</div>
      <div class="ag-row" aria-hidden="true">
        <img class="day" src="/img/landing/robot.svg" alt="" width="300" height="280">
        <img class="night" src="/img/landing/robot-night.svg" alt="" width="300" height="280">
        <div class="ag-stack">
          <div class="bubble ag-a"><p>A Next.js app that reads DATABASE_URL. I'm setting up a project and a service.</p></div>
          <div class="bubble ag-b">
            <p>This adds a PostgreSQL database, billed by usage. Should I create it?</p>
            <div class="ag-btns">
              <span class="bb bb-primary ag-press">Create it</span>
              <span class="bb">Not now</span>
              <svg class="ag-cur" viewBox="0 0 24 24"><path d="M4 3l15 7.5-6.4 1.9L10.2 19z" fill="oklch(0.25 0.03 55)" stroke="oklch(0.99 0.006 80)" stroke-width="1.6" stroke-linejoin="round" /></svg>
            </div>
          </div>
          <div class="bubble ag-c"><p>Created. DATABASE_URL is linked to it, and I never see the value.</p></div>
          <div class="bubble ag-d"><p>Built and rolled out. It's live at vouch-production-e2qn5.lucity.app.</p></div>
        </div>
      </div>
    </div>

    <div class="ag-ui" aria-hidden="true">
      <div class="ag-top">
        <svg width="20" height="20" viewBox="0 0 20 20" class="shrink-0"><rect width="20" height="20" rx="5" fill="#8843db" /></svg>
        <span>vouch</span><em>/</em><span>production</span>
        <span class="ag-new">+ New</span>
      </div>

      <div class="ag-node ag-n1 ag-svc">
        <b><svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" class="shrink-0"><path :d="github" /></svg> web</b>
        <small>
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0"><circle cx="12" cy="12" r="10" /><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" /><path d="M2 12h20" /></svg>
          <span>vouch-production-e2qn5.lucity.app</span>
        </small>
        <small style="margin-top: 4px">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0"><path d="M6 3v12" /><circle cx="18" cy="6" r="3" /><circle cx="6" cy="18" r="3" /><path d="M18 9a9 9 0 0 1-9 9" /></svg>
          <span>zeitlos/vouch</span>
        </small>
        <div class="ag-foot stack">
          <span class="pill ag-s-build"><i class="dot-build" />Building</span>
          <span class="pill ag-s-deploy"><i class="dot-deploy" />Deploying</span>
          <span class="pill ag-s-on"><i class="dot-on" />Online</span>
        </div>
      </div>

      <div class="ag-node ag-n2 ag-ghost">
        <b class="opacity-60">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="oklch(0.50 0.13 250)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0"><ellipse cx="12" cy="5" rx="9" ry="3" /><path d="M3 5v14a9 3 0 0 0 18 0V5" /><path d="M3 12a9 3 0 0 0 18 0" /></svg>
          feedback
        </b>
        <div class="ag-foot ag-foot-dashed"><span class="pill pill-ask">Waiting for your OK</span></div>
      </div>

      <div class="ag-node ag-n2 ag-db">
        <b>
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="oklch(0.50 0.13 250)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="shrink-0"><ellipse cx="12" cy="5" rx="9" ry="3" /><path d="M3 5v14a9 3 0 0 0 18 0V5" /><path d="M3 12a9 3 0 0 0 18 0" /></svg>
          feedback
        </b>
        <div class="ag-foot stack">
          <span class="pill ag-db-start"><i class="dot-build" />Starting</span>
          <span class="pill ag-db-on"><i class="dot-on" />Online</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.ag {
  position: relative;
  margin-top: 28px;
  height: 560px;
}

.ag-note {
  position: absolute;
  left: 0;
  top: 0;
  width: 33%;
}

.ag-chat {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 38%;
  z-index: 2;
}

.ag-you {
  box-sizing: border-box;
  width: max-content;
  max-width: 100%;
  margin: 0 0 14px auto;
  padding: 10px 16px;
  border-radius: 16px 16px 4px 16px;
  background: oklch(0.25 0.03 55);
  color: oklch(0.95 0.02 80);
  font-size: 16px;
  font-weight: 500;
}

.ag-row {
  display: flex;
  align-items: flex-end;
  gap: 2px;
}

.ag-row > img.night,
.dark .ag-row > img.day {
  display: none;
}

.dark .ag-row > img.night {
  display: block;
}

.ag-row > img {
  flex: none;
  display: block;
  width: 168px;
  height: auto;
}

.ag-stack {
  flex: none;
  width: 330px;
  margin-bottom: 56px;
  display: grid;
  align-items: end;
}

.ag-stack > .bubble {
  grid-area: 1 / 1;
}

.bubble {
  position: relative;
  box-sizing: border-box;
  padding: 18px 20px 16px;
  border-radius: 18px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.bubble::after {
  content: "";
  position: absolute;
  left: -7px;
  bottom: 30px;
  width: 15px;
  height: 15px;
  transform: rotate(45deg);
  background: oklch(0.99 0.006 80);
  box-shadow: -1px 1px 0 oklch(0.25 0.04 55 / .10);
}

.bubble p {
  margin: 0;
  font-size: 17px;
  line-height: 1.45;
  color: oklch(0.25 0.03 55);
}

.ag-btns {
  position: relative;
  margin-top: 14px;
  display: flex;
  gap: 10px;
}

.bb {
  display: inline-flex;
  align-items: center;
  min-height: 42px;
  padding: 0 18px;
  border-radius: 8px;
  font-size: 15px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
  box-shadow: inset 0 0 0 1px oklch(0.85 0.03 75);
}

.bb-primary {
  background: oklch(0.25 0.03 55);
  color: oklch(0.95 0.02 80);
  box-shadow: none;
}

.ag-cur {
  position: absolute;
  left: 62px;
  top: 24px;
  width: 22px;
  height: 22px;
  opacity: 0;
  transform-origin: 4px 3px;
}

.ag-ui {
  position: absolute;
  left: 36%;
  right: -60px;
  top: 36px;
  bottom: -40px;
  border-radius: 18px 0 0 0;
  overflow: hidden;
  background-color: oklch(0.955 0.018 80);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .16) 1px, transparent 0);
  background-size: 20px 20px;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
  font-family: 'Sora', 'Mona Sans', system-ui, sans-serif;
  color: #301c0e;
}

.ag-top {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 54px;
  padding: 0 96px 0 22px;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 1px 0 oklch(0.25 0.04 55 / .10);
  font-size: 14.5px;
  font-weight: 500;
}

.ag-top em {
  font-style: normal;
  color: oklch(0.62 0.03 60);
}

.ag-new {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  font-size: 13px;
  font-weight: 600;
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .16);
}

.ag-node {
  position: absolute;
  width: 236px;
  box-sizing: border-box;
  padding: 16px 18px;
  border-radius: 16px;
  background: oklch(0.975 0.012 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .14), 0 10px 24px -14px oklch(0.25 0.04 55 / .4);
}

.ag-node b {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
}

.ag-node small {
  display: flex;
  align-items: center;
  gap: 7px;
  margin-top: 8px;
  font-size: 12px;
  color: oklch(0.42 0.03 55);
  white-space: nowrap;
}

.ag-node small span {
  overflow: hidden;
  text-overflow: ellipsis;
}

.ag-foot {
  margin-top: 14px;
  padding-top: 12px;
  border-top: 1px solid oklch(0.90 0.015 78);
}

.ag-foot-dashed {
  border-top-style: dashed;
  border-top-color: oklch(0.55 0.15 300 / .4);
}

.ag-ghost {
  background: oklch(0.975 0.012 80 / .5);
  box-shadow: none;
  outline: 1.5px dashed oklch(0.55 0.15 300 / .8);
  outline-offset: -1px;
}

.ag-n1 {
  left: 20%;
  top: 172px;
}

.ag-n2 {
  left: 53%;
  top: 306px;
}

.stack {
  display: grid;
  justify-items: start;
  align-items: center;
}

.stack > * {
  grid-area: 1 / 1;
}

.pill {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  padding: 3px 10px;
  border-radius: 999px;
  background: oklch(0.90 0.015 75);
  font-size: 12px;
  font-weight: 500;
  color: oklch(0.35 0.03 55);
}

.pill i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.pill-ask {
  background: oklch(0.93 0.045 300);
  color: oklch(0.42 0.15 300);
}

.dot-on {
  background: oklch(0.70 0.19 150);
}

.dot-build {
  background: oklch(0.75 0.15 75);
}

.dot-deploy {
  background: oklch(0.62 0.18 300);
}

.ag-chip { animation: ag-chip 18s linear infinite; }
.ag-a { animation: ag-a 18s linear infinite; opacity: 0; }
.ag-b { animation: ag-b 18s linear infinite; }
.ag-ghost { animation: ag-ghost 18s linear infinite; }
.ag-c { animation: ag-c 18s linear infinite; opacity: 0; }
.ag-d { animation: ag-d 18s linear infinite; opacity: 0; }
.ag-s-build { animation: ag-s-build 18s linear infinite; }
.ag-s-deploy { animation: ag-s-deploy 18s linear infinite; opacity: 0; }
.ag-s-on { animation: ag-s-on 18s linear infinite; opacity: 0; }
.ag-db-start { animation: ag-db-start 18s linear infinite; }
.ag-db-on { animation: ag-db-on 18s linear infinite; opacity: 0; }
.ag-svc { animation: ag-svc 18s ease-out infinite; }
.ag-db { animation: ag-db 18s ease-out infinite; opacity: 0; }
.ag-cur { animation: ag-cur 18s ease-in-out infinite; }
.ag-press { animation: ag-press 18s ease-out infinite; }

@keyframes ag-chip { 0% { opacity: 0; } 1.56% { opacity: 0; } 2.78% { opacity: 1; } 96.11% { opacity: 1; } 97.33% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-a { 0% { opacity: 0; } 7.67% { opacity: 0; } 8.89% { opacity: 1; } 27.78% { opacity: 1; } 29% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-b { 0% { opacity: 0; } 26.56% { opacity: 0; } 27.78% { opacity: 1; } 50% { opacity: 1; } 51.22% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-ghost { 0% { opacity: 0; } 26.56% { opacity: 0; } 27.78% { opacity: 1; } 50% { opacity: 1; } 51.22% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-c { 0% { opacity: 0; } 48.78% { opacity: 0; } 50% { opacity: 1; } 75% { opacity: 1; } 76.22% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-d { 0% { opacity: 0; } 73.78% { opacity: 0; } 75% { opacity: 1; } 96.11% { opacity: 1; } 97.33% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-s-build { 0% { opacity: 1; } 63.89% { opacity: 1; } 65.11% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-s-deploy { 0% { opacity: 0; } 62.67% { opacity: 0; } 63.89% { opacity: 1; } 75% { opacity: 1; } 76.22% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-s-on { 0% { opacity: 0; } 73.78% { opacity: 0; } 75% { opacity: 1; } 100% { opacity: 1; } }
@keyframes ag-db-start { 0% { opacity: 1; } 61.11% { opacity: 1; } 62.33% { opacity: 0; } 100% { opacity: 0; } }
@keyframes ag-db-on { 0% { opacity: 0; } 59.89% { opacity: 0; } 61.11% { opacity: 1; } 100% { opacity: 1; } }
@keyframes ag-svc { 0%, 14.44% { opacity: 0; transform: translateY(10px) scale(.95); } 16.94% { opacity: 1; transform: none; } 96.11% { opacity: 1; transform: none; } 97.78%, 100% { opacity: 0; transform: none; } }
@keyframes ag-db { 0%, 50% { opacity: 0; transform: translateY(10px) scale(.95); } 52.5% { opacity: 1; transform: none; } 96.11% { opacity: 1; transform: none; } 97.78%, 100% { opacity: 0; transform: none; } }
@keyframes ag-cur { 0%, 36.67% { opacity: 0; transform: translate(150px, 104px); } 37.78% { opacity: 1; transform: translate(150px, 104px); } 45% { opacity: 1; transform: none; } 46.39% { opacity: 1; transform: scale(.84); } 47.78% { opacity: 1; transform: none; } 49.17% { opacity: 1; transform: none; } 50%, 100% { opacity: 0; transform: none; } }
@keyframes ag-press { 0%, 45.83% { transform: none; } 46.67% { transform: scale(.94); } 48.33%, 100% { transform: none; } }

@media (max-width: 1100px) {
  .ag {
    height: auto;
    display: flex;
    flex-direction: column;
    gap: 28px;
  }

  .ag-note,
  .ag-chat {
    position: static;
    width: auto;
  }

  .ag-you {
    margin-left: 0;
  }

  .ag-row > img {
    width: 112px;
  }

  .ag-stack {
    flex: 1 1 0;
    width: auto;
    min-width: 0;
    margin-bottom: 36px;
  }

  .ag-ui {
    position: relative;
    left: auto;
    right: auto;
    top: auto;
    bottom: auto;
    height: 400px;
    margin: 0 -16px -40px;
    border-radius: 18px 18px 0 0;
  }

  .ag-top {
    padding-right: 22px;
  }

  .ag-n1 {
    left: 16px;
    top: 78px;
  }

  .ag-n2 {
    left: auto;
    right: 16px;
    top: 232px;
  }
}

@media (max-width: 520px) {
  .ag-row > img {
    width: 76px;
  }

  .bb {
    padding: 0 14px;
    white-space: nowrap;
  }
}

@media (prefers-reduced-motion: reduce) {
  .ag * {
    animation: none !important;
  }
}
</style>
