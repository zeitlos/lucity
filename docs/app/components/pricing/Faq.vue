<script setup lang="ts">
const questions = [
  { q: 'What happens when the trial ends?', a: 'After 14 days or €5 of usage, whichever comes first, the workspace is suspended until you choose a plan. Services stop and nothing can be deployed, but databases, volumes and buckets keep their contents, and everything comes back the way it was.' },
  { q: 'Do unused credits roll over?', a: 'No. Credit left over at the end of a billing period does not carry over to the next one.' },
  { q: 'Can I see what I am spending?', a: 'Yes. Billing in the dashboard shows the plan, when the billing period ends, what the workspace has used so far and the estimated total.' },
  { q: 'Can I switch plans?', a: 'Yes, at any time under Billing. The new plan applies immediately and the fee is prorated for the rest of the period. Only admins of a workspace can change its plan.' },
  { q: 'What if a payment fails?', a: 'The workspace is suspended until the payment method is updated. Your data stays put, and the workspace resumes on its own as soon as the payment goes through.' },
  { q: 'Which currency do you bill in?', a: 'In euros, the same as on your invoice. Prices exclude taxes such as VAT, which are added where they apply.' },
];
</script>

<template>
  <FaqSection
    id="billing-faq"
    title="Questions about billing"
    :items="questions"
    ask-label="Ask about billing"
    ask-href="mailto:christian@zeitlos.software?subject=Billing"
  />
</template>
