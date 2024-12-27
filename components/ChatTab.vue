<template>
    <div class="flex gap-4">
      <Card class="w-1/3">
        <CardHeader>
          <CardTitle>Friends</CardTitle>
        </CardHeader>
        <CardContent>
          <ScrollArea class="h-[calc(100vh-200px)]">
            <div class="space-y-2">
              <div
                v-for="request in requests"
                :key="request.email"
                class="w-full justify-between flex items-center"
              >
                <div class="flex items-center gap-2">
                <img :src="request.profile_picture" class="w-8 h-8 rounded-full object-cover" />
                  <p>
                    {{ request.name }}
                  </p>
                </div>
                <div class="flex gap-2">
                  <button @click="acceptRequest(request)" class="p-2 bg-green-500 text-white rounded-full">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  </button>
                  <button @click="denyRequest(request)" class="p-2 bg-red-500 text-white rounded-full">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>
              <Button
                v-for="friend in friends"
                :key="friend.email"
                :variant="selectedFriend?.email === friend.email ? 'secondary' : 'ghost'"
                class="w-full justify-start flex items-center"
                @click="selectFriend(friend)"
              >
                <img :src="friend.profile_picture" class="w-8 h-8 rounded-full object-cover" />
                <p>
                  {{ friend.name }}
                </p>
              </Button>
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
  
      <Card class="w-2/3">
        <CardHeader class="flex justify-between">
          <CardTitle>
            {{ selectedFriend ? `${selectedFriend.name}` : 'Select a friend' }}
          </CardTitle>
          <DropdownMenu v-if="selectedFriend">
            <DropdownMenuTrigger>
              <button class="p-2 bg-gray-200 rounded-full">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6h.01M12 12h.01M12 18h.01" />
                </svg>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuItem @click="viewProfile(selectedFriend)">View Profile</DropdownMenuItem>
              <DropdownMenuItem @click="denyRequest(selectedFriend)">Delete Friend</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
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
      <Dialog v-if="isDialogOpen" @close="isDialogOpen = false">
        <DialogTrigger></DialogTrigger>
        <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ selectedFriend?.name }}'s Profile</DialogTitle>
        </DialogHeader>
          
        <div class="space-y-2">
          <p><strong>Email:</strong> {{ selectedFriend?.email }}</p>
          <p><strong>Specialty:</strong> {{ friendProfile?.specialty }}</p>
          <p><strong>Occupation:</strong> {{ friendProfile?.occupation }}</p>
          <p><strong>Bio:</strong> {{ friendProfile?.bio }}</p>
          <div>
            <strong>Languages:</strong>
              <div class="flex flex-wrap space-x-2 text-sm">
                <p v-for="language in friendProfile?.language" :key="language" class="dark:bg-slate-800 bg-slate-200 rounded-lg pl-2 mb-1 pr-2">
                  {{ language }}
                </p>
              </div>
            </div>
            <div>
              <strong>Interests:</strong>
              <div class="flex flex-wrap space-x-2 text-sm">
                <p v-for="interest in friendProfile?.interests" :key="interest" class="dark:bg-blue-950 bg-blue-100 rounded-lg pl-2 mb-1 pr-2">
                  {{ interest }}
                </p>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button @click="isDialogOpen = false">Close</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  </template>
  
  <script setup>
  import { ref } from 'vue'
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
  import { ScrollArea } from '@/components/ui/scroll-area'
  import { defineProps, defineEmits, onMounted } from 'vue'
  import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'

  const props = defineProps({
    user: {
      type: Object,
      required: true,
    }
  })
  const user = props.user;

  const friends = ref([]);
  const requests = ref([]);
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
  const fetchRequests = async () => {
    try{
      const response = await fetch(`https://www.pairgrid.com/api/getrequests/getrequests?user_id=${user.id}`, {
        method: 'GET',
      });
      if(!response.ok) throw new Error('Failed to fetch friend requests');
      const data = await response.json();
      requests.value = data;
    } catch (err) {
      console.error(err);
      error.value = err.message;
      emit('toast-update', 'Error fetching friend requests');
    }
  }
  onMounted(() => {
    fetchFriends();
    fetchRequests();
  });
  const acceptRequest = async (request) => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${user.id}&friend_email=${request.email}`, {
        method: 'GET',
      });
      if (!response.ok) throw new Error('Failed to accept friend request');
      friends.value.push(request);
      requests.value = requests.value.filter((r) => r.email !== request.email);
      emit('toast-update', `Successfully connected with ${request.name}`);
    } catch (err) {
      console.error(err);
      emit('toast-update', 'Error accepting friend request');
    }
  };
  const denyRequest = async (request) => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/deletefriend/deletefriend?user_id=${user.id}&friend_email=${request.email}`, {
        method: 'GET',
      });
      if (!response.ok) throw new Error('Failed to deny friend request');
      requests.value = requests.value.filter((r) => r.email !== request.email);
      emit('toast-update', `${request.name}'s friend request denied`);
    } catch (err) {
      console.error(err);
      emit('toast-update', 'Error denying friend request');
    }
  };
  
  const selectedFriend = ref(null)
  const messages = ref([])
  const newMessage = ref('')
  const friendProfile = ref(null);

  const fetchFriendProfile = async (friend) => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/getuserinfo/getuserinfo?user_id=${friend.email}`, {
        method: 'GET',
      });
      if (!response.ok) throw new Error('Failed to fetch user profile');
      const data = await response.json();
      friendProfile.value = data;
    } catch (err) {
      console.error(err);
      emit('toast-update', 'Error fetching friend profile');
    }
  };
  const viewProfile = (friend) => {
    fetchFriendProfile(friend);
    isDialogOpen.value = true;
  }
  
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