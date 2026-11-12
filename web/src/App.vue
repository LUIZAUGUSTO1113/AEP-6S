<script setup lang="ts">
import { useTemplateRef, watch } from 'vue'
import { useRoute } from 'vue-router'
import ToastHost from './components/ToastHost.vue'

const route = useRoute()
const backToTop = useTemplateRef<HTMLElement>('backToTop')

watch(
  () => route.path,
  () => {
    backToTop.value?.focus()
  },
)
</script>

<template>
  <span
    id="topo"
    ref="backToTop"
    tabindex="-1"
  />
  <ul class="skip-links">
    <li>
      <a
        class="skip-link"
        href="#conteudo"
      >Ir para o conteúdo</a>
    </li>
  </ul>
  <header class="site-header">
    <div class="shell">
      <p class="brand">Monitoramento de água</p>
      <nav
        class="site-nav"
        aria-label="Seções"
      >
        <RouterLink :to="{ name: 'stations' }">Estações</RouterLink>
        <RouterLink :to="{ name: 'samples' }">Coletas</RouterLink>
        <RouterLink :to="{ name: 'alerts' }">Alertas</RouterLink>
      </nav>
    </div>
  </header>
  <ToastHost />
  <main id="conteudo">
    <div class="shell">
      <RouterView />
    </div>
  </main>
</template>
