<template>
    <Card>
      <CardHeader>
        <CardTitle>Programming Preferences</CardTitle>
      </CardHeader>
      <CardContent>
        <form @submit.prevent="savePreferences" class="space-y-6">
          <div class="space-y-2">
            <Label for="language">Favorite Programming Language</Label>
            <Select v-model="preferences.language">
              <SelectTrigger id="language">
                <SelectValue placeholder="Select a language" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="javascript">JavaScript</SelectItem>
                <SelectItem value="python">Python</SelectItem>
                <SelectItem value="java">Java</SelectItem>
                <SelectItem value="csharp">C#</SelectItem>
                <SelectItem value="ruby">Ruby</SelectItem>
              </SelectContent>
            </Select>
          </div>
  
          <div class="space-y-2">
            <Label for="environment">Preferred Development Environment</Label>
            <Select v-model="preferences.environment">
              <SelectTrigger id="environment">
                <SelectValue placeholder="Select an environment" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="vscode">Visual Studio Code</SelectItem>
                <SelectItem value="intellij">IntelliJ IDEA</SelectItem>
                <SelectItem value="sublime">Sublime Text</SelectItem>
                <SelectItem value="vim">Vim</SelectItem>
              </SelectContent>
            </Select>
          </div>
  
          <div class="space-y-2">
            <Label>Coding Style</Label>
            <RadioGroup v-model="preferences.style">
              <div class="flex items-center space-x-2">
                <RadioGroupItem value="tabs" id="tabs" />
                <Label for="tabs">Tabs</Label>
              </div>
              <div class="flex items-center space-x-2">
                <RadioGroupItem value="spaces" id="spaces" />
                <Label for="spaces">Spaces</Label>
              </div>
            </RadioGroup>
          </div>
  
          <div class="space-y-2">
            <Label>Interests (select multiple)</Label>
            <div class="space-y-2">
              <div v-for="interest in interests" :key="interest" class="flex items-center space-x-2">
                <Checkbox
                  :id="interest"
                  :checked="preferences.interests.includes(interest)"
                  @update:checked="toggleInterest(interest)"
                />
                <Label :for="interest">{{ interest }}</Label>
              </div>
            </div>
          </div>
  
          <Button type="submit">Save Preferences</Button>
        </form>
      </CardContent>
    </Card>
  </template>
  
  <script setup>
  import { ref } from 'vue'
  import { Button } from '@/components/ui/button'
  import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card'
  import { Label } from '@/components/ui/label'
  import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
  import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
  import { Checkbox } from '@/components/ui/checkbox'
  
  const interests = [
    'Web Development',
    'Mobile Development',
    'Data Science',
    'Machine Learning',
    'DevOps',
    'Cybersecurity',
    'Blockchain',
  ]
  
  const preferences = ref({
    language: 'javascript',
    environment: 'vscode',
    style: 'spaces',
    interests: [],
  })
  
  const toggleInterest = (interest) => {
    const index = preferences.value.interests.indexOf(interest)
    if (index === -1) {
      preferences.value.interests.push(interest)
    } else {
      preferences.value.interests.splice(index, 1)
    }
  }
  
  const savePreferences = () => {
    // In a real app, you'd save these preferences to a backend or local storage
    console.log('Saving preferences:', preferences.value)
    alert('Preferences saved!')
  }
  </script>