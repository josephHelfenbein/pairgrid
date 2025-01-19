<template>
    <div class="flex flex-col" :style="{ height: `calc(${Math.max(calculatedHeight, 400)}px)` }">
      <ScrollArea ref="scrollArea" class="flex-grow mb-4">
        <div class="flex justify-center items-center h-full" v-if="chatLoading">
          <Loader size="80px" />
        </div>
        <div v-else>
          <div
            v-for="message in messages"
            :key="message.id"
            :class="[
              'flex items-start gap-2 p-2'
            ]"
          >
            <img
              :src="message.senderIcon"
              class="w-5 h-5 md:w-8 md:h-8 rounded-full object-cover"
              />
            <div>
              <div class="flex items-center gap-2">
                <p class="font-bold text-xs md:text-sm">{{ message.sender || 'Unknown' }}</p>
                <small class="text-xs text-gray-500">{{ formatTimestamp(message.id) }}</small>
              </div>
              <p class="text-sm">{{ message.text }}</p>
            </div>
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
  import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
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
  const calculatedHeight = ref(window.innerHeight - 300)
  
  const updateHeight = () => {
    calculatedHeight.value = window.innerHeight - 300
    }

	onMounted(() => {
    window.addEventListener('resize', updateHeight)
    updateHeight()
	})

	onUnmounted(() => {
    window.removeEventListener('resize', updateHeight)
	})

  watch(() => props.newMessage, (newVal) => {
    localNewMessage.value = newVal
  })

  function formatTimestamp(timestamp) {
    if (!timestamp) return ''
    const date = new Date(timestamp)
    return date.toLocaleString()
  }
  
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