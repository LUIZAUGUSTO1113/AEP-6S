<script setup lang="ts">
import { useToastsStore } from '../stores/toasts'

const toasts = useToastsStore()
</script>

<template>
  <ul
    class="toasts"
    aria-live="assertive"
  >
    <li
      v-for="item in toasts.items"
      :key="item.id"
      class="toast"
    >
      {{ item.text }}
    </li>
  </ul>
</template>

<style scoped>
.toasts {
  position: fixed;
  right: var(--space-5);
  bottom: var(--space-5);
  z-index: 1100;
  display: flex;
  flex-direction: column-reverse;
  gap: var(--space-2);
  width: min(16rem, calc(100% - var(--space-5) * 2));
  margin: 0;
  padding: 0;
  list-style: none;
}

.toast {
  margin: 0;
  padding: var(--space-2);
  background: var(--alert-bg);
  color: var(--alert-fg);
  font-size: 0.8125rem;
  font-weight: 600;
  line-height: 1.4;
  border: 1px solid var(--alert-fg);
}

@media (max-width: 960px) {
  .toasts {
    left: var(--space-4);
    right: var(--space-4);
    bottom: calc(var(--space-4) + env(safe-area-inset-bottom));
    width: auto;
  }

  .toast {
    font-size: 0.875rem;
  }
}
</style>
