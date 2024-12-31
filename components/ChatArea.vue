<template>
    <div class="flex flex-col h-[calc(100vh-350px)]">
      <ScrollArea ref="scrollArea" class="flex-grow mb-4">
        <div class="flex justify-center items-center h-full" v-if="chatLoading">
          <Loader size="80px" />
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="message in messages"
            :key="message.id"
            :class="[
              'max-w-[80%] p-2 rounded-lg',
              message.sender === 'me'
                ? 'ml-auto bg-primary text-primary-foreground'
                : 'bg-muted'
            ]"
          >
            {{ message.text }}
          </div>
        </div>
      </ScrollArea>
        
      <div class="flex gap-2">
        <Input
          v-model="localNewMessage"
          placeholder="Type a message..."
          @keyup.enter="sendMessage"
        />
        <Button @click="sendMessage">Send</Button>
      </div>
    </div>
</template>
  
<script setup>
  import { ref, watch, nextTick } from 'vue'
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { ScrollArea } from '@/components/ui/scroll-area'
  import Loader from '@/components/Loader'
  
  const props = defineProps({
    selectedFriend: Object,
    messages: Array,
    chatLoading: Boolean,
    newMessage: String,
  })
  
  const emit = defineEmits(['sendMessage', 'updateNewMessage'])
  
  const scrollArea = ref(null)
  const localNewMessage = ref(props.newMessage)
  
  watch(() => props.newMessage, (newVal) => {
    localNewMessage.value = newVal
  })
  
  watch(() => props.messages, () => {
    scrollToBottom()
  }, { deep: true })
  
  const scrollToBottom = () => {
    nextTick(() => {
      const viewportEl = scrollArea.value?.scrollAreaViewport?.$el
      if (viewportEl) {
        setTimeout(() => {
          viewportEl.parentElement.scrollTop = viewportEl.parentElement.scrollHeight
        }, 0)
      }
    })
  }
  
  const sendMessage = () => {
    emit('sendMessage')
    emit('updateNewMessage', '')
    localNewMessage.value = ''
  }
  
  watch(localNewMessage, (newVal) => {
    emit('updateNewMessage', newVal)
  })
</script>