<template>
    <div class="flex gap-4">
      <Card class="w-1/3">
        <CardHeader>
          <CardTitle>Friends</CardTitle>
        </CardHeader>
        <CardContent>
          <ScrollArea class="h-[calc(100vh-200px)]">
            <div class="space-y-2">
              <Button
                v-for="friend in friends"
                :key="friend.id"
                :variant="selectedFriend?.id === friend.id ? 'secondary' : 'ghost'"
                class="w-full justify-start"
                @click="selectFriend(friend)"
              >
                {{ friend.name }}
              </Button>
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
  
      <Card class="w-2/3">
        <CardHeader>
          <CardTitle>
            {{ selectedFriend ? `Chat with ${selectedFriend.name}` : 'Select a friend' }}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div v-if="selectedFriend" class="flex flex-col h-[calc(100vh-300px)]">
            <ScrollArea class="flex-grow mb-4">
              <div class="space-y-2">
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
                v-model="newMessage"
                placeholder="Type a message..."
                @keyup.enter="sendMessage"
              />
              <Button @click="sendMessage">Send</Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  </template>
  
  <script setup>
  import { ref } from 'vue'
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { ScrollArea } from '@/components/ui/scroll-area'
  
  const friends = [
    { id: 1, name: 'Alice' },
    { id: 2, name: 'Bob' },
    { id: 3, name: 'Charlie' },
  ]
  
  const selectedFriend = ref(null)
  const messages = ref([])
  const newMessage = ref('')
  
  const selectFriend = (friend) => {
    selectedFriend.value = friend
    messages.value = [
      { id: 1, sender: 'me', text: 'Hey there!' },
      { id: 2, sender: friend.name, text: 'Hi! How are you?' },
    ]
  }
  
  const sendMessage = () => {
    if (newMessage.value.trim()) {
      messages.value.push({
        id: messages.value.length + 1,
        sender: 'me',
        text: newMessage.value,
      })
      newMessage.value = ''
    }
  }
  </script>