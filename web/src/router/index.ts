import { createRouter, createWebHistory } from 'vue-router'
import AlertsView from '../views/AlertsView.vue'
import SamplesView from '../views/SamplesView.vue'
import StationsView from '../views/StationsView.vue'

export default createRouter({
  history: createWebHistory(),
  linkActiveClass: 'is-active',
  linkExactActiveClass: 'is-active',
  routes: [
    { path: '/', redirect: { name: 'stations' } },
    { path: '/estacoes', name: 'stations', component: StationsView },
    { path: '/coletas', name: 'samples', component: SamplesView },
    { path: '/alertas', name: 'alerts', component: AlertsView },
  ],
})
