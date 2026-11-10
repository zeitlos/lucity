<script setup lang="ts">
interface Message {
  avatar: 'agent' | 'lucity' | 'alert' | 'person';
  who: string;
  when: string;
  text: string;
  buttons?: [string, string];
  approved?: string;
}

interface Scene {
  tab: string;
  heading: string;
  text: string;
  app: { icon: string; color: string; channel: string; thread: string };
  messages: Message[];
  day: { title: string; blips: number[]; note: string };
}

const scenes: Scene[] = [
  {
    tab: 'A deploy fails',
    heading: 'A failed deploy wakes it.',
    text: "Lucity's own events are triggers: failed builds and deploys, crash loops, full disks, new vulnerabilities and expiring certificates.",
    app: { icon: 'mattermost', color: '#1e58c9', channel: 'Deploys', thread: 'api, release 214' },
    messages: [
      { avatar: 'lucity', who: 'Lucity', when: '02:14', text: 'The deploy of api failed while building.' },
      { avatar: 'agent', who: 'Deploy doctor', when: '02:14', text: 'On it. Reading the build log and the last three commits.' },
      { avatar: 'agent', who: 'Deploy doctor', when: '02:16', text: 'The build needs STRIPE_WEBHOOK_SECRET since commit 4f2a91c, and it is not set at build time. The previous release still serves traffic. Should I add the variable from the shared group and redeploy?', buttons: ['Add and redeploy', 'Not now'] },
    ],
    day: { title: 'Deploy doctor, yesterday', blips: [9.3, 58.7, 73.0], note: 'Awake for 11 minutes. Asleep, and nearly free, the rest of the day.' },
  },
  {
    tab: 'An alert at night',
    heading: 'Your monitoring wakes it.',
    text: 'Alerts, new errors and new issues from the tools you already use start a conversation. Events about the same thing land in the same thread.',
    app: { icon: 'telegram', color: '#1c93d1', channel: 'Shop on call', thread: 'web out of memory' },
    messages: [
      { avatar: 'alert', who: 'Alert', when: '03:12', text: 'web ran out of memory and restarted three times in ten minutes.' },
      { avatar: 'agent', who: 'Night watch', when: '03:13', text: 'Memory has climbed since the 01:50 deploy. The image resizer keeps uploads in memory. Raising the limit to 1 GB stops the restarts. May I?', approved: 'Approved by Noa at 03:15' },
      { avatar: 'agent', who: 'Night watch', when: '03:15', text: 'Done. web has been stable since. I filed the leak as SHOP-482 for the morning.' },
    ],
    day: { title: 'Night watch, yesterday', blips: [13.3, 98.6], note: 'Awake for 6 minutes. Asleep, and nearly free, the rest of the day.' },
  },
  {
    tab: 'Every morning',
    heading: 'The clock wakes it.',
    text: 'Any schedule, in your time zone. It reads what happened, writes its report and goes back to sleep.',
    app: { icon: 'mattermost', color: '#1e58c9', channel: 'Ops', thread: 'Morning report' },
    messages: [
      { avatar: 'agent', who: 'Morning report', when: '06:00', text: 'Good morning. Four deploys last night, one failed and was fixed at 02:21. web ran out of memory at 03:12, and Night watch raised it with a yes from Noa.' },
      { avatar: 'agent', who: 'Morning report', when: '06:00', text: 'One thing needs a person today: the certificate for shop.example renews in six days and its DNS record is gone.' },
    ],
    day: { title: 'Morning report, yesterday', blips: [25.0], note: 'Awake for 2 minutes. Asleep, and nearly free, the rest of the day.' },
  },
  {
    tab: 'A teammate asks',
    heading: 'Your team wakes it.',
    text: 'Mention it in any chat it was added to. It knows who is asking, because teammates link their account once.',
    app: { icon: 'slack', color: '#4a154b', channel: 'shop-dev', thread: 'checkout is slow' },
    messages: [
      { avatar: 'person', who: 'Mara', when: '13:42', text: '@Triage why is checkout slow since lunch?' },
      { avatar: 'agent', who: 'Triage', when: '13:43', text: 'Response times went from 180 ms to 1.4 s at 12:40, right after release 214. The new query on orders runs without an index. Want an issue with the details?' },
      { avatar: 'person', who: 'Mara', when: '13:44', text: 'Yes please.' },
      { avatar: 'agent', who: 'Triage', when: '13:44', text: 'Filed as SHOP-483, with the slow query and a suggested index.' },
    ],
    day: { title: 'Triage, yesterday', blips: [38.2, 47.9, 57.1], note: 'Awake for 14 minutes. Asleep, and nearly free, the rest of the day.' },
  },
];

const features = [
  { name: 'Asleep when idle', text: 'No work, no bill worth mentioning. A trigger wakes it in seconds.' },
  { name: 'Picks up where it left off', text: 'Files, memory and conversations survive restarts, deploys and sleep.' },
  { name: 'A sandbox of its own', text: "Each agent runs isolated, on infrastructure built for other people's code." },
  { name: 'Answers where you talk', text: 'Telegram, Mattermost, Slack, Teams or the dashboard, approvals included.' },
];

const tabs = scenes.map(scene => scene.tab);
const way = ref(0);
const scene = computed(() => scenes[way.value]!);
</script>

<template>
  <section class="mx-auto max-w-content px-6" aria-labelledby="on">
    <LandingChapterHeading id="on" beat="joy" promise="Always on" rest="long after you log off.">
      A hosted agent does not wait for your terminal. It sleeps until something happens, does the work, and tells your team in the chat they already use.
    </LandingChapterHeading>

    <LandingStage tint="joy" class="on-stage mt-14 px-4 pt-7 pb-10 sm:px-9">
      <LandingTabs v-model="way" :options="tabs" label="What wakes an agent" id-prefix="wake" />
      <div id="wake-panel" class="wk" role="tabpanel" :aria-labelledby="`wake-tab-${way}`">
        <div class="wk-note">
          <h3 class="font-display text-2xl/tight text-stone-900 dark:text-stone-100">{{ scene.heading }}</h3>
          <p class="mt-2.5 text-base text-[oklch(0.36_0.04_300)] dark:text-stone-300">{{ scene.text }}</p>
          <p class="mt-3 text-base">
            <NuxtLink to="/docs/agents/triggers" class="font-semibold text-stone-900 underline decoration-1 underline-offset-4 dark:text-stone-100">Every trigger</NuxtLink>
          </p>
        </div>

        <div class="day">
          <b>{{ scene.day.title }}</b>
          <div class="day-bar" role="img" aria-label="A day from midnight to midnight with the few minutes the agent was awake marked">
            <i v-for="blip in scene.day.blips" :key="blip" :style="{ left: `${blip}%` }" />
          </div>
          <div class="day-x" aria-hidden="true"><span>00:00</span><span>06:00</span><span>12:00</span><span>18:00</span><span>24:00</span></div>
          <p>{{ scene.day.note }}</p>
        </div>

        <div class="ch">
          <div class="ch-top">
            <AgentsIcon :name="scene.app.icon" :size="22" :style="{ color: scene.app.color }" />
            <span>{{ scene.app.channel }}</span>
            <small>{{ scene.app.thread }}</small>
          </div>
          <div class="ch-body">
            <div v-for="(message, index) in scene.messages" :key="index" class="msg">
              <span class="av" :class="`av-${message.avatar}`" aria-hidden="true">
                <AgentsIcon v-if="message.avatar === 'agent'" name="bot" :size="22" />
                <AgentsIcon v-else-if="message.avatar === 'lucity'" name="lucity" :size="20" />
                <AgentsIcon v-else-if="message.avatar === 'alert'" name="bell" :size="20" />
                <template v-else>{{ message.who[0] }}</template>
              </span>
              <div>
                <b>{{ message.who }}</b><time>{{ message.when }}</time>
                <p>{{ message.text }}</p>
                <div v-if="message.buttons" class="acts">
                  <span class="bb bb-primary">{{ message.buttons[0] }}</span>
                  <span class="bb">{{ message.buttons[1] }}</span>
                </div>
                <p v-if="message.approved" class="okd">
                  <svg width="18" height="18" viewBox="0 0 24 24" aria-hidden="true"><path d="M4 12.5 L10 18 L20 6" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" /></svg>
                  <span>{{ message.approved }}</span>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </LandingStage>

    <div class="mt-10 grid grid-cols-[repeat(auto-fit,minmax(min(250px,100%),1fr))] gap-x-10 border-b border-stone-300 dark:border-stone-800">
      <div v-for="feature in features" :key="feature.name" class="border-t border-stone-300 py-5.5 dark:border-stone-800">
        <b class="block text-lg font-semibold text-stone-900 dark:text-stone-100">{{ feature.name }}</b>
        <span class="mt-1 block text-base text-stone-700 dark:text-stone-400">{{ feature.text }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.on-stage :deep(.tab) {
  --tab-off: oklch(0.38 0.05 300);
}

.dark .on-stage :deep(.tab) {
  --tab-off: oklch(0.84 0.04 300);
}

.wk {
  position: relative;
  margin-top: 28px;
  height: 472px;
}

.wk-note {
  position: absolute;
  left: 0;
  top: 0;
  width: 30%;
}

.day {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 40%;
  z-index: 2;
  box-sizing: border-box;
  padding: 20px 24px 18px;
  border-radius: 14px;
  background: oklch(0.99 0.006 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .10), 0 24px 40px -24px oklch(0.25 0.04 55 / .35);
}

.day b {
  display: block;
  font-size: 16px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.day p {
  margin: 12px 0 0;
  font-size: 14.5px;
  line-height: 1.45;
  color: oklch(0.45 0.04 55);
}

.day-bar {
  position: relative;
  height: 38px;
  margin-top: 14px;
  box-shadow: 0 1px 0 oklch(0.25 0.03 55);
  background-color: oklch(0.25 0.04 55 / .035);
  background-image: radial-gradient(circle at 1px 1px, oklch(0.25 0.04 55 / .30) 1px, transparent 0);
  background-size: 7px 7px;
}

.day-bar i {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 5px;
  margin-left: -2px;
  background: oklch(0.55 0.15 300);
}

.day-x {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
  font-size: 13px;
  color: oklch(0.47 0.04 55);
}

.ch {
  position: absolute;
  left: 38%;
  right: -60px;
  top: 0;
  bottom: -40px;
  border-radius: 18px 0 0 0;
  overflow: hidden;
  background: oklch(0.985 0.008 80);
  box-shadow: 0 0 0 1px oklch(0.25 0.04 55 / .12), 0 28px 56px -22px oklch(0.25 0.04 55 / .42);
}

.ch-top {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 58px;
  padding: 0 96px 0 44px;
  border-bottom: 1px solid oklch(0.91 0.015 80);
  font-size: 16.5px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.ch-top small {
  font-size: 14.5px;
  font-weight: 400;
  color: oklch(0.47 0.04 55);
}

.ch-body {
  display: flex;
  flex-direction: column;
  gap: 22px;
  padding: 24px 96px 0 44px;
}

.msg {
  display: grid;
  grid-template-columns: 40px minmax(0, 1fr);
  gap: 0 14px;
}

.av {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 10px;
}

.av-agent {
  background: #8843db;
  color: white;
}

.av-lucity {
  background: oklch(0.93 0.045 300);
  color: #8843db;
}

.av-alert {
  background: oklch(0.93 0.08 85);
  color: oklch(0.42 0.10 70);
}

.av-person {
  border-radius: 50%;
  background: oklch(0.88 0.05 200);
  color: oklch(0.25 0.03 55);
  font-weight: 600;
}

.msg b {
  font-size: 15.5px;
  font-weight: 600;
  color: oklch(0.25 0.03 55);
}

.msg time {
  margin-left: 8px;
  font-size: 13px;
  color: oklch(0.47 0.04 55);
}

.msg p {
  max-width: 620px;
  margin: 3px 0 0;
  font-size: 16.5px;
  line-height: 1.5;
  color: oklch(0.35 0.03 55);
}

.acts {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 12px;
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
  white-space: nowrap;
  box-shadow: inset 0 0 0 1px oklch(0.85 0.03 75);
}

.bb-primary {
  background: oklch(0.25 0.03 55);
  color: oklch(0.95 0.02 80);
  box-shadow: none;
}

.msg .okd {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
  font-size: 15px;
  font-weight: 600;
  color: oklch(0.43 0.12 160);
}

@media (max-width: 1000px) {
  .wk {
    display: flex;
    flex-direction: column;
    gap: 24px;
    height: auto;
  }

  .wk-note,
  .day,
  .ch {
    position: static;
    width: auto;
  }

  .ch {
    margin: 0 -16px -40px;
    padding-bottom: 40px;
    border-radius: 18px 18px 0 0;
  }

  .ch-top,
  .ch-body {
    padding-right: 20px;
    padding-left: 20px;
  }
}

@media (max-width: 640px) {
  .on-stage :deep(.tabs) {
    grid-auto-flow: row;
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
