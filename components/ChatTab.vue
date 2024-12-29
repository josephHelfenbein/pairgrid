<template>
    <div class="flex gap-4">
      <Card class="w-1/3">
        <CardHeader>
          <CardTitle>Friends</CardTitle>
        </CardHeader>
        <CardContent>
          <ScrollArea class="h-[calc(100vh-300px)]">
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
                <div>
                  <p class="text-left">
                  {{ friend.name }}
                  </p>
                  <p class="text-sm text-left text-gray-500">{{ getLastSeenText(friend.last_seen) }}</p>
                </div>
              </Button>
            </div>
          </ScrollArea>
        </CardContent>
      </Card>
  
      <Card class="w-2/3">
        <CardHeader class="flex flex-row justify-between items-center">
          <CardTitle class="flex-shrink-0 flex items-center">
            {{ selectedFriend ? `${selectedFriend.name}` : 'Select a friend' }}
          </CardTitle>
          <Dialog>
            <DropdownMenu v-if="selectedFriend" class="w-8">
              <DropdownMenuTrigger>
                <button class="p-2 w-6 flex-shrink-0">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" color="#505050" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6h.01M12 12h.01M12 18h.01" />
                  </svg>
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DialogTrigger asChild>
                  <DropdownMenuItem>View Profile</DropdownMenuItem>
                </DialogTrigger>
                <DropdownMenuItem @click="removeFriend(selectedFriend)">Remove Friend</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
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
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardHeader>
        <CardContent>
          <div v-if="selectedFriend" class="flex flex-col h-[calc(100vh-350px)]">
            <ScrollArea ref="scrollArea" class="flex-grow mb-4">
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
  import { Button } from '@/components/ui/button'
  import { Input } from '@/components/ui/input'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
  import { ScrollArea } from '@/components/ui/scroll-area'
  import { ref, defineProps, defineEmits, onMounted, onBeforeUnmount, nextTick } from 'vue'
  import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
  import Pusher from 'pusher-js'
  import { useRuntimeConfig } from '#app'
  import CryptoJS from 'crypto-js'

  const props = defineProps({
    user: {
      type: Object,
      required: true,
    }
  })
  const user = props.user;

  const getLastSeenText = (lastSeen) => {
    const now = new Date();
    const lastSeenDate = new Date(lastSeen);
    const diffInSeconds = Math.floor((now - lastSeenDate) / 1000);

    if (diffInSeconds < 60) {
      return `Last seen just now`;
    }
    const diffInMinutes = Math.floor(diffInSeconds / 60);
    if (diffInMinutes < 60) {
      return `Last seen ${diffInMinutes} minute${diffInMinutes > 1 ? 's' : ''} ago`;
    }
    const diffInHours = Math.floor(diffInMinutes / 60);
    if (diffInHours < 24) {
      return `Last seen ${diffInHours} hour${diffInHours > 1 ? 's' : ''} ago`;
    }
    const diffInDays = Math.floor(diffInHours / 24);
    return `Last seen ${diffInDays} day${diffInDays > 1 ? 's' : ''} ago`;
  }

  const friends = ref([]);
  const requests = ref([]);
  const error = ref(null);
  const emit = defineEmits(['toast-update']);
  const selectedFriend = ref(null);
  const messages = ref([]);
  const newMessage = ref('');
  const friendProfile = ref(null);
  const pusher = ref(null);
  const channel = ref(null);
  const scrollArea = ref(null);

  const pusherConfig = {
    appKey: useRuntimeConfig().public.pusherAppKey,
    cluster: "us2",
  }

  const scrollToBottom = () => {
    nextTick(()=>{
      const viewportEl = scrollArea.value?.scrollAreaViewport?.$el
      if (viewportEl) {
        setTimeout(() => {
          console.log(viewportEl.parentElement);
          viewportEl.parentElement.scrollTop = viewportEl.parentElement.scrollHeight;
        }, 0); 
      }
    });
  };

  const fetchFriends = async () =>{
    try{
      const response = await fetch(`https://www.pairgrid.com/api/getrequests/getrequests?user_id=${user.id}&kind=friend`, {
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
      const response = await fetch(`https://www.pairgrid.com/api/getrequests/getrequests?user_id=${user.id}&kind=request`, {
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
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${user.id}&friend_email=${request.email}&operation=add`, {
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
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${user.id}&friend_email=${request.email}&operation=remove`, {
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
  const removeFriend = async (request) => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${user.id}&friend_email=${request.email}&operation=remove`, {
        method: 'GET',
      });
      if (!response.ok) throw new Error('Failed to remove friend');
      friends.value = friends.value.filter((r) => r.email !== request.email);
      emit('toast-update', `${request.name} removed from friends list`);
    } catch (err) {
      console.error(err);
      emit('toast-update', 'Error removing friend');
    }
  };
  const sendMessage = async () => {
    if(!newMessage.value || !selectedFriend.value) return;
    try{
      const encryptionKey = generateEncryptionKey(user.id); 
      const encryptedMessage = encryptMessage(newMessage.value, encryptionKey);
      const payload = {
        sender_id: user.id,
        receiver_email: selectedFriend.value.email,
        content: encryptedMessage.encryptedData,
        key: encryptedMessage.iv,
      };
      messages.value.push({
        id: new Date().getTime(),
        sender: 'me',
        text: newMessage.value,
      });
      scrollToBottom();
      newMessage.value = '';
      const response = await fetch('https://www.pairgrid.com/api/sendmessage/sendmessage', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });
      if(!response.ok) throw new Error('Failed to send message');
    } catch (err) {
      console.error(err);
      emit('toast-update', 'Error sending message');
    }
  }
  const generateEncryptionKey = (userID) => {
    const serverSideSecret = useRuntimeConfig().public.encryptionKey;
    return CryptoJS.SHA256(userID + serverSideSecret);
  }
  const encryptMessage = (message, key) => {
    const iv = CryptoJS.lib.WordArray.random(16);
    const encrypted = CryptoJS.AES.encrypt(message, key, { iv: iv });
    
    return {
      encryptedData: encrypted.ciphertext.toString(CryptoJS.enc.Hex),
      iv: iv.toString(CryptoJS.enc.Hex)
    };
  }
  const decryptMessage = (encryptedMessage, key, iv) => {
    const ivWordArray = CryptoJS.enc.Hex.parse(iv);
    const encryptedWordArray = CryptoJS.enc.Hex.parse(encryptedMessage);

    const decrypted = CryptoJS.AES.decrypt(
      { ciphertext: encryptedWordArray },
      key,
      { iv: ivWordArray }
    );

    return decrypted.toString(CryptoJS.enc.Utf8);  
  }
  
  const unsubscribeFromChatChannel = () =>{
    if (channel.value) {
      channel.value.unbind_all();
      channel.value.unsubscribe();
    }
    if (pusher.value) {
      pusher.value.disconnect();
    }
  }
  const subscribeToChatChannel = () => {
    if(!selectedFriend.value || !friendProfile.value) return;
    unsubscribeFromChatChannel();
    const friendID = friendProfile.value.id;
    const firstID = user.id < friendID ? user.id : friendID;
    const secondID = user.id > friendID ? user.id : friendID;
    const newChannel = `chat-${firstID}-${secondID}`;
    pusher.value = new Pusher(pusherConfig.appKey, {
      cluster: pusherConfig.cluster,
    });
    channel.value = pusher.value.subscribe(newChannel);
    channel.value.bind('new-message', (data) => {
      const decryptedMessage = decryptMessage(data.encrypted_content, generateEncryptionKey(data.sender_id), data.key);
      
      if(data.sender_id != user.id) {
        scrollToBottom();
        messages.value.push({
          id: data.created_at,
          sender: data.sender_id == user.id ? 'me' : selectedFriend.value.name,
          text: decryptedMessage,
        });
      }
    });
  }

  onBeforeUnmount(() => {
    unsubscribeFromChatChannel();
  });
  

  const fetchFriendProfile = async (friend) => {
    try {
      const emailData = { email: friend.email };
      const response = await fetch('https://www.pairgrid.com/api/getuser/getuser', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(emailData),
      });
      if (!response.ok) throw new Error('Failed to fetch user profile');
      const data = await response.json();
      friendProfile.value = data;
      await getMessages();
      scrollToBottom();
      subscribeToChatChannel();
    } catch (err) {
      console.error(err);
      emit('toast-update', 'Error fetching friend profile');
    }
  };

  const getMessages = async () => {
    try {
      const response = await fetch(`https://www.pairgrid.com/api/getmessages/getmessages?user_id=${user.id}&friend_id=${friendProfile.value.id}`, {
        method: 'GET',
      });
      if (!response.ok) throw new Error('Failed to fetch messages');
      const data = await response.json();
      messages.value = data.map(message => {
        const decryptedMessage = decryptMessage(
          message.encrypted_content, 
          generateEncryptionKey(message.sender_id), 
          message.key
        );

        return {
          id: message.created_at,
          sender: message.sender_id == user.id ? 'me' : selectedFriend.value.name,
          text: decryptedMessage,
        };
      });
    } catch (err) {
      console.error(err);
      emit('toast-update', 'Error loading chat');
    }
  }
  
  const selectFriend = (friend) => {
    selectedFriend.value = friend
    messages.value = [];
    fetchFriendProfile(friend);
  }
  </script>