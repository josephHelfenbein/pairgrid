<script setup lang="ts">
import { DialogTrigger, type DialogTriggerProps } from 'radix-vue'

const props = defineProps<DialogTriggerProps>()
const emit = defineEmits<{
  (e: 'pointerdown', event: PointerEvent): void
}>()

const handleInteraction = (event: PointerEvent) => {
  emit('pointerdown', event)
}
</script>

<template>
  <DialogTrigger 
    v-bind="props"
    @pointerdown="handleInteraction"
  >
    <slot />
  </DialogTrigger>
</template>