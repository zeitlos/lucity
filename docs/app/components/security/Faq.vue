<script setup lang="ts">
const questions = [
  { q: 'Does a finding block my deploy?', a: 'No. The scan for leaked keys runs next to the build and reports what it found. The release ships either way, and a failed build never replaces the version that is already running.' },
  { q: 'Do you use our keys to check them?', a: "Only to ask the key's own provider whether it is valid, with a read-only request. Found keys are stored masked, and only members of your workspace see the findings." },
  { q: 'I rotated the key. Why is the finding still there?', a: 'A finding belongs to a commit, so it stays while that commit is in your history. The mark that a key still works is set when the key is found and is not checked again on later deploys yet. Rotating the key is what makes it harmless.' },
  { q: 'Which apps are checked?', a: 'Every app Lucity builds from your repository, in every environment. Prebuilt images from public registries are not scanned yet.' },
  { q: 'What leaves the EU?', a: 'Requests reach your apps through a global edge network, and some platform logs and metadata are analyzed with an AI provider in the United States under standard contractual clauses. The data processing agreement lists every sub-processor.' },
  { q: 'Can our security team read the code?', a: 'Yes. Lucity is open source, so they can check how scanning and isolation work instead of taking our word for it.' },
];
</script>

<template>
  <FaqSection
    id="security-faq"
    title="What reviewers ask us"
    :items="questions"
    ask-label="Ask us"
    ask-href="mailto:christian@zeitlos.software?subject=Security%20question"
  />
</template>
