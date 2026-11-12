<script setup lang="ts">
import BaseButton from './BaseButton.vue'

defineEmits<{
  edit: []
  remove: []
}>()
</script>

<template>
  <div class="row-actions">
    <BaseButton @click="$emit('edit')">Editar</BaseButton>
    <BaseButton @click="$emit('remove')">Excluir</BaseButton>
  </div>
</template>
