<script setup lang="ts">
import { ref } from 'vue';
import {
  Download,
  FileArchive,
  Loader2,
  Ship,
  GitBranch,
  Shield,
} from 'lucide-vue-next';
import { toast } from '@/components/ui/sonner';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { errorMessage } from '@/lib/utils';

const props = defineProps<{
  open: boolean;
  projectId: string;
  projectName: string;
}>();

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void;
}>();

const ejecting = ref(false);

const features = [
  { icon: Ship, label: 'Helm chart with all templates' },
  { icon: GitBranch, label: 'ArgoCD manifests per environment' },
  { icon: Shield, label: 'Zero Lucity dependency after export' },
];

async function handleEject() {
  ejecting.value = true;
  try {
    const url = `/api/eject/${encodeURIComponent(props.projectId)}`;
    const res = await fetch(url, { credentials: 'include' });

    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || 'Eject failed');
    }

    const blob = await res.blob();
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    const shortName = props.projectName.split('/').pop() || props.projectName;
    a.download = `${shortName}-ejected.zip`;
    a.click();
    URL.revokeObjectURL(a.href);

    emit('update:open', false);
    toast.success('Project ejected', {
      description: 'Your project has been exported as a self-contained zip file.',
    });
  } catch (e: unknown) {
    toast.error('Failed to eject project', { description: errorMessage(e) });
  } finally {
    ejecting.value = false;
  }
}
</script>

<template>
  <Dialog
    :open="open"
    @update:open="emit('update:open', $event)"
  >
    <DialogContent class="overflow-hidden p-0 sm:max-w-[480px]">
      <!-- Hero section with pattern background -->
      <div class="pattern-circles relative flex flex-col items-center bg-card px-8 pt-10 pb-6">
        <!-- Radial fade overlay -->
        <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_30%,var(--card)_75%)]" />

        <!-- Icon -->
        <div class="relative mb-5 flex h-14 w-14 items-center justify-center rounded-2xl border border-border bg-background shadow-sm">
          <FileArchive
            :size="26"
            class="text-primary"
          />
        </div>

        <DialogHeader class="relative space-y-2 text-center">
          <DialogTitle class="font-serif text-[1.75rem] leading-tight tracking-[-0.01em]">
            Eject project
          </DialogTitle>
          <DialogDescription class="text-sm leading-relaxed text-muted-foreground">
            Export <strong class="font-medium text-foreground">{{ projectName }}</strong>
            as a self-contained zip you can deploy anywhere.
          </DialogDescription>
        </DialogHeader>
      </div>

      <!-- Content -->
      <div class="space-y-5 px-8 pt-2 pb-8">
        <!-- Feature list -->
        <ul class="space-y-3">
          <li
            v-for="feature in features"
            :key="feature.label"
            class="flex items-center gap-3 text-sm"
          >
            <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-primary/10">
              <component
                :is="feature.icon"
                :size="14"
                class="text-primary"
              />
            </div>
            <span class="text-foreground">{{ feature.label }}</span>
          </li>
        </ul>

        <!-- Non-destructive notice -->
        <p class="rounded-lg border border-dashed border-border bg-muted/50 px-3 py-2 text-center text-xs text-muted-foreground">
          Non-destructive — your project keeps running on Lucity.
        </p>

        <!-- Actions -->
        <div class="flex gap-3">
          <Button
            variant="outline"
            class="flex-1"
            @click="emit('update:open', false)"
          >
            Cancel
          </Button>
          <Button
            class="flex-1"
            :disabled="ejecting"
            @click="handleEject"
          >
            <Loader2
              v-if="ejecting"
              :size="14"
              class="mr-1.5 animate-spin"
            />
            <Download
              v-else
              :size="14"
              class="mr-1.5"
            />
            {{ ejecting ? 'Exporting...' : 'Download .zip' }}
          </Button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
