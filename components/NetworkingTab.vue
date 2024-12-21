<template>
    <div>
      <h2 class="text-2xl font-bold mb-4">Recommended Connections</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card v-for="person in recommendedPeople" :key="person.id">
          <CardHeader>
            <CardTitle>{{ person.name }}</CardTitle>
            <CardDescription>{{ person.title }}</CardDescription>
          </CardHeader>
          <CardContent>
            <p class="mb-4">{{ person.bio }}</p>
            <Button @click="connect(person)">Connect</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  </template>
  
  <script setup>
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'
  
  const recommendedPeople = [
    { id: 1, name: 'David Lee', title: 'Full Stack Developer', bio: 'Passionate about Vue.js and Node.js' },
    { id: 2, name: 'Emma Wilson', title: 'UX Designer', bio: 'Creating intuitive and beautiful interfaces' },
    { id: 3, name: 'Michael Chen', title: 'Data Scientist', bio: 'Machine learning enthusiast' },
    { id: 4, name: 'Sarah Johnson', title: 'DevOps Engineer', bio: 'Automating everything with Ansible and Docker' },
    { id: 5, name: 'Alex Rodriguez', title: 'Mobile Developer', bio: 'Building cross-platform apps with Flutter' },
  ]
  
  const connect = (person) => {
    // In a real app, you'd implement the connection logic here
    alert(`Connecting with ${person.name}`)
  }
  </script>