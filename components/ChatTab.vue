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
                :key="friend.email"
                :variant="selectedFriend?.email === friend.email ? 'secondary' : 'ghost'"
                class="w-full justify-start flex items-center"
                @click="selectFriend(friend)"
              >
                <img :src="friend.profilePicture" class="w-16 h-16 rounded-full object-cover" />
                <p>
                  {{ friend.name }}
                </p>
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
  import { defineProps, defineEmits, onMounted } from 'vue'

  const props = defineProps({
    user: {
      type: Object,
      required: true,
    }
  })
  const user = props.user;

  const friends = ref([]);
  const error = ref(null);
  const emit = defineEmits(['toast-update']);
  const fetchFriends = async () =>{
    try{
      const response = await fetch(`https://www.pairgrid.com/api/getfriends/getfriends?user_id=${user.id}`, {
        method: 'GET',
      });
      if(!response.ok) throw new Error('Failed to fetch friends');
      const data = await response.json();
      friends.value = data;
    } catch (err) {
      console.error(err);
      error.value = err.message;
      emit('toast-update', 'Error fetching friends');
    }
  };
  onMounted(() => {
    fetchFriends();
  });
  
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