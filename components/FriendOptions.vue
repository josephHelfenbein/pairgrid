<template>
  <Dialog>
    <DropdownMenu v-if="selectedFriend">
      <DropdownMenuTrigger asChild>
        <button class="p-2 w-6 flex-shrink-0">
          <MoreVertical class="h-6 w-6" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
          <DialogTrigger asChild>
            <DropdownMenuItem>View Profile</DropdownMenuItem>
          </DialogTrigger>
          <DropdownMenuItem @click="$emit('removeFriend', selectedFriend)">Remove Friend</DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{{ selectedFriend?.name }}'s Profile</DialogTitle>
        </DialogHeader>
        
        <div class="space-y-2">
          <p><strong>Email:</strong> {{ selectedFriend?.email }}</p>
          <p><strong>Specialty:</strong> {{ selectedFriend?.specialty }}</p>
          <p><strong>Occupation:</strong> {{ selectedFriend?.occupation }}</p>
          <p><strong>Bio:</strong> {{ selectedFriend?.bio }}</p>
          <div>
            <strong>Languages:</strong>
            <div class="flex flex-wrap space-x-2 text-sm">
              <p v-for="language in selectedFriend?.language" :key="language" class="dark:bg-slate-800 bg-slate-200 rounded-lg pl-2 mb-1 pr-2">
                {{ language }}
              </p>
            </div>
          </div>
          <div>
            <strong>Interests:</strong>
            <div class="flex flex-wrap space-x-2 text-sm">
              <p v-for="interest in selectedFriend?.interests" :key="interest" class="dark:bg-blue-950 bg-blue-100 rounded-lg pl-2 mb-1 pr-2">
                {{ interest }}
              </p>
            </div>
          </div>
        </div>

        <DialogFooter>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </template>
  
<script setup>
  import { MoreVertical } from 'lucide-vue-next'
  import { 
    DropdownMenu, 
    DropdownMenuContent, 
    DropdownMenuItem, 
    DropdownMenuTrigger 
  } from '@/components/ui/dropdown-menu'
  import { 
    Dialog, 
    DialogContent, 
    DialogHeader, 
    DialogTitle, 
    DialogTrigger, 
    DialogFooter 
  } from '@/components/ui/dialog'
  
  defineProps({
    selectedFriend: Object,
  })
  
  defineEmits(['removeFriend'])
</script>