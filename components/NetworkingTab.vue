<template>
    <div>
      <h2 class="text-2xl font-bold mb-4">Recommended Connections</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card v-for="person in recommendedPeople">
          <div class="ml-4 flex items-center">
            <img :src="person.profile_picture" class="w-16 h-16 rounded-full object-cover" />
            <CardHeader class="pl-4">
              <CardTitle>{{ person.name }}</CardTitle>
              <CardDescription>{{ person.specialty + ', ' + person.occupation }}</CardDescription>
            </CardHeader>
          </div>
          <CardContent>
            <p class="mb-2">{{ person.bio }}</p>
            <div class="flex flex-wrap space-x-2 text-sm mb-1">
              <p class="dark:bg-slate-800 bg-slate-200 rounded-lg pl-2 mb-1 pr-2" v-for="language in person.language">{{ language }}</p>
            </div>
            <div class="flex flex-wrap space-x-2 text-sm mb-3">
              <p class="dark:bg-blue-950 bg-blue-100 rounded-lg pl-2 mb-1 pr-2" v-for="interest in person.interests">{{ interest }}</p>
            </div>
            <Button v-if="!sentTo.includes(person)" class="bg-gradient-to-t from-primary to-violet-800 hover:from-primary hover:to-violet-500" @click="connect(person)">Connect</Button>
            <Button v-else disabled class="bg-gray-500 cursor-not-allowed">Request Sent</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  </template>
  
  <script setup>
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
  import { defineProps, defineEmits, onMounted } from 'vue';

  const props = defineProps({
    user: {
      type: Object,
      required: true,
    }
  })
  const user = props.user;
  const emit = defineEmits(['toast-update']);

  const recommendedPeople = ref([]);
  const sentTo = ref([]);
  const error = ref(null);
  const fetchRecommendedPeople = async () =>{
    try{
      const response = await fetch(`https://www.pairgrid.com/api/getusers/getusers?user_id=${user.id}`, {
        method: 'GET',
      });
      if(!response.ok) throw new Error('Failed to fetch recommended people');
      const data = await response.json();
      recommendedPeople.value = data;
    } catch (err) {
      console.error(err);
      error.value = err.message;
      emit('toast-update', 'Error fetching recommended connections');
    }
  };
  onMounted(() => {
    fetchRecommendedPeople();
  });
  
  const connect = async (person) => {
    try{
      const response = await fetch(`https://www.pairgrid.com/api/addfriend/addfriend?user_id=${user.id}&friend_email=${person.email}`, {
        method: 'GET',
      })
      if(!response.ok) throw new Error('Failed to connect with the user');
      const data = await response.json();
      sentTo.value.push(person);
      emit('toast-update', `Sent friend request to ${person.name}`);
    } catch(err) {
      console.error(err);
      emit('toast-update', 'Error connecting with the user');
    }
  }
  </script>