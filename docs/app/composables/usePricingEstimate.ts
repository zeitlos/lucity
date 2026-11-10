const CPU_RATE = 20;
const MEMORY_RATE = 10;
const DISK_RATE = 0.1;
const EGRESS_RATE = 0.02;
const VCPU_PER_GB = 0.5;
const DATABASE_INSTANCES = 2;
const DATABASE_MEMORY_SHARE = 0.5;

const INSTANCES = [1, 2, 4, 8];

const MEMORY = [
  { label: '256 MB', gb: 0.25 },
  { label: '512 MB', gb: 0.5 },
  { label: '1 GB', gb: 1 },
  { label: '2 GB', gb: 2 },
  { label: '4 GB', gb: 4 },
];

const BUSY = [
  { label: 'A little', memory: 0.3, cpu: 0.05 },
  { label: 'About half', memory: 0.5, cpu: 0.2 },
  { label: 'Most of it', memory: 0.75, cpu: 0.5 },
];

const DATABASE = [
  { label: 'None', gb: 0, cpu: 0, disk: 0 },
  { label: '1 GB', gb: 1, cpu: 0.05, disk: 10 },
  { label: '2 GB', gb: 2, cpu: 0.1, disk: 20 },
  { label: '8 GB', gb: 8, cpu: 0.25, disk: 100 },
];

const EGRESS = [
  { label: '10 GB', gb: 10 },
  { label: '100 GB', gb: 100 },
  { label: '1 TB', gb: 1000 },
];

const TEAM = ['Just me', 'A team'];

const HOBBY = { name: 'Hobby', fee: 5, instances: 3, databaseStorage: 32 };
const PRO = { name: 'Pro', fee: 25 };

export interface EstimateLine {
  label: string;
  detail: string;
  amount: string;
}

function round(n: number) {
  return Math.round(n * 100) / 100;
}

function money(n: number) {
  return n.toFixed(2);
}

function amount(gb: number) {
  return gb < 1 ? `${Math.round(gb * 1024)} MB` : `${Math.round(gb * 100) / 100} GB`;
}

function cores(n: number) {
  return `${Math.round(n * 1000) / 1000} vCPU`;
}

export function usePricingEstimate() {
  const instances = ref(1);
  const memory = ref(2);
  const busy = ref(1);
  const database = ref(2);
  const egress = ref(0);
  const team = ref(0);

  const groups = [
    { label: 'Instances of your app', options: INSTANCES.map(String), selected: instances },
    { label: 'Memory per instance', options: MEMORY.map(option => option.label), selected: memory },
    { label: 'How much of that it uses', options: BUSY.map(option => option.label), selected: busy },
    { label: 'PostgreSQL database', options: DATABASE.map(option => option.label), selected: database },
    { label: 'Traffic sent out, per month', options: EGRESS.map(option => option.label), selected: egress },
    { label: 'Who works on it', options: TEAM, selected: team },
  ];

  const estimate = computed(() => {
    const count = INSTANCES[instances.value]!;
    const size = MEMORY[memory.value]!;
    const use = BUSY[busy.value]!;
    const db = DATABASE[database.value]!;
    const out = EGRESS[egress.value]!;

    const appCpu = count * size.gb * VCPU_PER_GB * use.cpu;
    const appMemory = count * size.gb * use.memory;
    const app = round(appCpu * CPU_RATE + appMemory * MEMORY_RATE);

    const lines: EstimateLine[] = [{
      label: `${count} ${count > 1 ? 'instances' : 'instance'} of your app`,
      detail: `${cores(appCpu)} and ${amount(appMemory)} in use`,
      amount: money(app),
    }];

    let databaseCost = 0;

    if (db.gb) {
      const databaseCpu = DATABASE_INSTANCES * db.cpu;
      const databaseMemory = DATABASE_INSTANCES * db.gb * DATABASE_MEMORY_SHARE;
      const databaseDisk = DATABASE_INSTANCES * db.disk;
      databaseCost = round(databaseCpu * CPU_RATE + databaseMemory * MEMORY_RATE + databaseDisk * DISK_RATE);
      lines.push({
        label: 'PostgreSQL with a standby',
        detail: `${cores(databaseCpu)} and ${amount(databaseMemory)} in use, ${databaseDisk} GB of disk`,
        amount: money(databaseCost),
      });
    }

    const egressCost = round(out.gb * EGRESS_RATE);
    lines.push({ label: 'Egress', detail: `${out.label} sent out`, amount: money(egressCost) });

    const usage = round(app + databaseCost + egressCost);

    const isTeam = team.value === 1;
    const tooManyInstances = count > HOBBY.instances;
    const tooMuchStorage = db.disk > HOBBY.databaseStorage;
    const needsPro = isTeam || tooManyInstances || tooMuchStorage;
    const plan = needsPro || usage >= PRO.fee ? PRO : HOBBY;
    const credit = Math.min(usage, plan.fee);
    const total = Math.max(usage, plan.fee);

    let note = 'The usage goes past the CHF 5 that Hobby includes, so the bill is the usage.';

    if (isTeam) {
      note = 'Pro, because Hobby is for one person.';
    } else if (tooManyInstances) {
      note = 'Pro, because Hobby runs up to three instances per service.';
    } else if (tooMuchStorage) {
      note = 'Pro, because Hobby databases go up to 32 GB of storage.';
    } else if (plan === PRO) {
      note = 'From CHF 25 of usage, Pro costs the same as Hobby and gives you more room.';
    } else if (usage <= plan.fee) {
      note = 'The usage stays inside what Hobby includes, so the plan is the whole bill.';
    }

    return {
      lines,
      plan: plan.name,
      usage: money(usage),
      fee: money(plan.fee),
      credit: `−${money(credit)}`,
      total: `CHF ${money(total)}`,
      note,
      memoryShare: `${Math.round(use.memory * 100)}%`,
      cpuShare: `${Math.round(use.cpu * 100)}%`,
    };
  });

  return { groups, estimate };
}
