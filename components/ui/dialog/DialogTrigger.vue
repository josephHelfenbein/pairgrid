<script setup lang="ts">
import { DialogTrigger, type DialogTriggerProps } from 'radix-vue'

const props = defineProps<DialogTriggerProps>()
const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
  (e: 'touchstart', event: TouchEvent): void
}>()

const isMouseEvent = (event: Event): event is MouseEvent => {
  return event.type === 'click'
}

const handleInteraction = (event: MouseEvent | TouchEvent) => {
  if (isMouseEvent(event)) {
    emit('click', event)
  } else {
    emit('touchstart', event)
  }
}
</script>

<template>
  <DialogTrigger 
    v-bind="props"
    @click="handleInteraction"
    @touchstart="handleInteraction"
  >
    <slot />
  </DialogTrigger>
</template>