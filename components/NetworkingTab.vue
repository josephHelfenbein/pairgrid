<template>
    <div>
      <h2 class="text-2xl font-bold mb-4">Recommended Connections</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card v-for="person in recommendedPeople">
          <CardHeader>
            <CardTitle>{{ person.name }}</CardTitle>
            <CardDescription>{{ person.specialty + ', ' + person.occupation }}</CardDescription>
          </CardHeader>
          <CardContent>
            <p class="mb-2">{{ person.bio }}</p>
            <div class="flex space-x-2 text-sm mb-2">
              <p class="dark:bg-slate-800 bg-slate-200 rounded-lg pl-2 pr-2" v-for="language in person.language">{{ language }}</p>
            </div>
            <div class="flex space-x-2 text-sm mb-4">
              <p class="dark:bg-blue-950 bg-blue-100 rounded-lg pl-2 pr-2" v-for="interest in person.interests">{{ interest }}</p>
            </div>
            <Button class="bg-gradient-to-t from-primary to-violet-800 hover:from-primary hover:to-violet-500" @click="connect(person)">Connect</Button>
          </CardContent>
        </Card>
      </div>
    </div>
  </template>
  
  <script setup>
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '@/components/ui/card'

  const recommendedPeople = [
    { 
      name: 'David Lee', 
      specialty: 'Full Stack Developer', 
      occupation: 'Hobbyist',
      bio: 'Passionate about Vue.js and Node.js',
      language: ['JavaScript', 'TypeScript'],
      interests: ['Web Development', 'Cloud Computing', 'Blockchain']
    },
    { 
      name: 'Emma Wilson', 
      specialty: 'Designer',
      occupation: 'Professional', 
      bio: 'Creating intuitive and beautiful interfaces',
      language: [],
      interests: ['UI/UX Design', 'Web Development'],
    },
    { 
      name: 'Sarah Johnson', 
      specialty: 'DevOps Engineer',
      occupation: 'Educator', 
      bio: 'Automating everything with Ansible and Docker',
      language: ['C/C++', 'Ruby', 'Python', 'PHP', 'JavaScript'],
      interests: ['DevOps', 'Cloud Computing'],
    },
    { 
      name: 'Alex Rodriguez', 
      specialty: 'Mobile Developer', 
      occupation: 'Undergraduate Student',
      bio: 'Building cross-platform apps with Flutter',
      language: ['Dart', 'JavaScript'],
      interests: ['Mobile Development', 'Game Development'],
    },

  ]
  
  const connect = (person) => {
    // In a real app, you'd implement the connection logic here
    alert(`Connecting with ${person.name}`)
  }
  </script>