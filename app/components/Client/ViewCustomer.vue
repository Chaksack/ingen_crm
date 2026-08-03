<script setup lang="ts">
import {
  AlertCircle,
  AtSign,
  Banknote,
  ChevronLeft,
  CircleCheck,
  MapPinned,
  Phone,
  PlusCircle,
  TriangleAlert,
} from 'lucide-vue-next'
import { computed, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { toast } from 'vue-sonner'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Avatar, AvatarFallback, AvatarImage } from '~/components/ui/avatar'

const creditScore = ref(0)
const maxCreditScore = 1000
const customers = ref<any>({})
const isLoading = ref(false)
const error = ref<string | null>(null)
const route = useRoute()

// Calculate progress percentage
const progress = computed(() => {
  const score = customers.value?.creditScore || creditScore.value || 0
  return (score / maxCreditScore) * 100
})

// Function to fetch user data from the API
async function fetchUser() {
  isLoading.value = true
  error.value = null
  try {
    const data = await $fetch<any>(`/api/clients/${route.params.id}`)
    customers.value = data

    // Update credit score from API if available
    if (data?.creditScore) {
      creditScore.value = data.creditScore
    }
  }
  catch (err: any) {
    console.error('Failed to fetch user details:', err)
    error.value = err?.message || 'Failed to load user details'
    toast.error('Error', { description: error.value })
  }
  finally {
    isLoading.value = false
  }
}

// Badge variant function based on status
function getBadgeVariant(status: string) {
  switch (status) {
    case 'active':
      return 'default'
    case 'inactive':
      return 'destructive'
    case 'pending':
      return 'secondary'
    default:
      return 'secondary'
  }
}

// KYC badge variant and icon
function getKycVariant(kyc: boolean) {
  return kyc
    ? { variant: 'default', icon: CircleCheck }
    : { variant: 'destructive', icon: TriangleAlert }
}

// Fetch user data when the component mounts
onMounted(() => {
  fetchUser()
})
</script>

<template>
  <div class="flex min-h-screen w-full flex-col bg-muted/40">
    <div class="flex flex-col sm:gap-4 sm:py-4 sm:pl-14">
      <header class="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6" />

      <main class="flex-1 items-start gap-4 p-4 sm:px-6 sm:py-0 md:gap-8">
        <div class="mx-auto grid max-w-full flex-1 auto-rows-max gap-4">
          <!-- Loading/Error State -->
          <div v-if="isLoading" class="text-center text-lg font-semibold">
            Loading customer data...
          </div>

          <Alert v-if="error && !isLoading" variant="destructive" class="mb-4">
            <AlertCircle class="w-5 h-5" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>{{ error }}</AlertDescription>
          </Alert>

          <!-- Header Section -->
          <div v-if="!isLoading && customers && Object.keys(customers).length > 0" class="flex items-center gap-4">
            <NuxtLink to="/clients">
              <Button variant="outline" size="icon" class="h-7 w-7">
                <ChevronLeft class="h-4 w-4" />
                <span class="sr-only">Back</span>
              </Button>
            </NuxtLink>

            <Avatar class="relative overflow-visible">
              <AvatarImage class="rounded-full" :src="customers.avatar || ''" alt="Customer Avatar" />
              <AvatarFallback class="text-white">
                {{ customers.name ? customers.name.substring(0, 2).toUpperCase() : '' }}
              </AvatarFallback>
              <span
                v-if="customers.status"
                :class="(customers.status === 'active') ? 'bg-green-500' : 'bg-red-500'"
                class="absolute bottom-[-4px] right-[-4px] w-3.5 h-3.5 rounded-full border-2 border-white"
              />
            </Avatar>

            <div class="flex shrink-0 whitespace-nowrap text-xl font-semibold tracking-tight sm:grow-0">
              <h3>{{ customers.name }}</h3>
              <p v-if="customers.idNumber" class="mx-2 mt-2 font-semibold text-sm">
                - {{ customers.idNumber }}
              </p>
            </div>

            <Badge :variant="getBadgeVariant(customers.status)">
              {{ customers.status || 'N/A' }}
            </Badge>
            <Badge :variant="getKycVariant(customers.kyc).variant">
              <component :is="getKycVariant(customers.kyc).icon" class="w-4 h-4 inline-block mr-1" />
              Compliance Check: {{ customers.kyc ? 'Verified' : 'Not Verified' }}
            </Badge>

            <div class="items-center gap-2 md:ml-auto md:flex">
              <Dialog>
                <DialogTrigger as-child>
                  <Button size="lg" class="bg-red-700 flex text-white">
                    Update Credit/Loan
                    <PlusCircle class="h-6 w-6 ml-2" />
                  </Button>
                </DialogTrigger>
                <DialogContent class="sm:max-w-2xl">
                  <DialogHeader>
                    <DialogTitle>Update Credit/Loan</DialogTitle>
                    <DialogDescription>
                      Modify your credit or loan details here.
                    </DialogDescription>
                  </DialogHeader>
                  <div class="py-4">
                    Update information
                  </div>
                  <DialogFooter>
                    <Button type="submit">
                      Save changes
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>

              <Dialog>
                <DialogTrigger as-child>
                  <Button size="lg" class="bg-lime-400 flex text-black">
                    Add New Credit/Loan
                    <PlusCircle class="h-6 w-6 ml-2" />
                  </Button>
                </DialogTrigger>
                <DialogContent class="sm:max-w-2xl">
                  <DialogHeader>
                    <DialogTitle>New Credit/Loan Application</DialogTitle>
                    <DialogDescription>
                      Add your credit or loan details here.
                    </DialogDescription>
                  </DialogHeader>
                  <div class="py-4">
                    Update information
                  </div>
                  <DialogFooter>
                    <Button type="submit">
                      Save changes
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </div>
          </div>

          <!-- Customer Info Row -->
          <div v-if="!isLoading && customers && Object.keys(customers).length > 0" class="flex shrink-0 whitespace-nowrap text-md font-medium tracking-tight sm:grow-0 flex-wrap gap-4">
            <h3 class="flex items-center">
              <MapPinned class="w-6 h-6 mr-2" />
              {{ customers.nationality || 'N/A' }}
            </h3>
            <h3 class="flex items-center">
              <Phone class="w-6 h-6 mr-2" />
              {{ customers.phone || 'N/A' }}
            </h3>
            <h3 class="flex items-center">
              <AtSign class="w-6 h-6 mr-2" />
              {{ customers.email || 'N/A' }}
            </h3>
            <h3 class="flex items-center">
              <Banknote class="w-6 h-6 mr-2" />
              £ {{ customers.monthlyIncome || '0' }}
            </h3>
          </div>

          <!-- Stats Cards -->
          <div v-if="!isLoading && customers && Object.keys(customers).length > 0" class="grid mt-6 gap-4 sm:grid-cols-2 md:grid-cols-4 lg:grid-cols-5">
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">
                  Credit Rating
                </CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  Good
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">
                  Credit Score
                </CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  {{ customers.creditScore || 0 }}/1000
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">
                  Credit Limit
                </CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  £ 16,399,000
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">
                  Total Assets
                </CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  £ 254,399,000
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">
                  Turnover
                </CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  £ 254,399,000
                </CardTitle>
              </CardHeader>
            </Card>
          </div>

          <!-- Tabs Section -->
          <Tabs v-if="!isLoading && customers && Object.keys(customers).length > 0" default-value="score">
            <div class="flex flex-col sm:flex-row items-center mt-4">
              <TabsList class="w-full bg-transparent font-thin lg:w-auto">
                <TabsTrigger value="score">
                  Score
                </TabsTrigger>
                <TabsTrigger value="accounts">
                  Accounts
                </TabsTrigger>
                <TabsTrigger value="credit">
                  Credits & loans
                </TabsTrigger>
                <TabsTrigger value="history">
                  History
                </TabsTrigger>
                <TabsTrigger value="compliance">
                  Compliance
                </TabsTrigger>
              </TabsList>
            </div>

            <!-- Score Tab -->
            <TabsContent value="score">
              <div class="grid gap-2 sm:grid-cols-1 md:grid-cols-[1fr_2fr]">
                <Card>
                  <CardHeader class="mt-2 pb-2">
                    <CardTitle>Current Score</CardTitle>
                    <CardDescription class="text-xs text-muted-foreground">
                      Score is in the GOLD Rating (800 - 1000)
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div class="flex flex-col items-center">
                      <div class="text-4xl font-semibold mb-4">
                        {{ customers.creditScore || 0 }} pts
                      </div>
                      <div class="relative w-full bg-gray-300 rounded-full h-4 overflow-hidden">
                        <div
                          class="bg-gradient-to-r from-red-500 via-yellow-400 to-green-500 h-4 rounded-full transition-all duration-300 ease-in-out"
                          :style="{ width: `${progress}%` }"
                        >
                          <div
                            class="absolute indicator top-1/2 transform -translate-y-1/2 w-5 h-5 bg-white border-2 border-black rounded-full flex items-center justify-center transition-all duration-300 ease-in-out"
                            :style="{ left: `calc(${progress}% - 12px)` }"
                          />
                        </div>
                      </div>
                      <div class="mt-2 text-sm">
                        {{ customers.creditScore || 0 }} / 1000 pts
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent class="pt-6">
                    <div class="grid gap-4 sm:grid-cols-1 md:grid-cols-3">
                      <Card>
                        <CardHeader class="pb-2">
                          <CardDescription>Current Rating</CardDescription>
                          <CardTitle class="text-xl">
                            Good
                          </CardTitle>
                        </CardHeader>
                        <CardContent>
                          <div class="text-xs text-muted-foreground">
                            Date changed: 11 November 2024
                          </div>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardHeader class="pb-2">
                          <CardDescription>Previous Rating</CardDescription>
                          <CardTitle class="text-xl">
                            Silver
                          </CardTitle>
                        </CardHeader>
                        <CardContent>
                          <div class="text-xs text-muted-foreground">
                            Date changed: 11 November 2024
                          </div>
                        </CardContent>
                      </Card>
                      <Card>
                        <CardHeader class="pb-2">
                          <CardDescription>Growth Score</CardDescription>
                          <CardTitle class="text-xl">
                            77 - Very Likely
                          </CardTitle>
                        </CardHeader>
                      </Card>
                    </div>
                  </CardContent>
                  <CardFooter>
                    <Card class="w-full">
                      <CardHeader class="pb-2">
                        <CardTitle>Description</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div class="text-xs text-muted-foreground">
                          Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
                        </div>
                      </CardContent>
                    </Card>
                  </CardFooter>
                </Card>
              </div>
            </TabsContent>

            <!-- Accounts Tab -->
            <TabsContent value="accounts">
              <Card class="mb-4">
                <CardHeader>
                  <CardTitle>Bank Accounts</CardTitle>
                  <CardDescription>Bank Account Information</CardDescription>
                </CardHeader>
                <CardContent v-if="customers.bankAccounts && customers.bankAccounts.length > 0">
                  <Accordion type="single" class="w-full" collapsible>
                    <AccordionItem
                      v-for="(account, index) in customers.bankAccounts"
                      :key="account.id"
                      class="border-b"
                      :value="`bank-${index}`"
                    >
                      <AccordionTrigger class="hover:no-underline">
                        <div class="flex items-center gap-4 flex-1">
                          <Avatar>
                            <AvatarFallback>
                              {{ account.bankName ? account.bankName[0] : 'B' }}
                            </AvatarFallback>
                          </Avatar>
                          <div class="text-left">
                            <div class="font-semibold">
                              {{ account.bankName || 'Unknown Bank' }}
                            </div>
                          </div>
                          <div class="ml-auto text-sm">
                            Account #: {{ account.accountNumber }}
                          </div>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <div class="grid gap-4 p-4 md:grid-cols-2">
                          <div>
                            <div class="text-sm text-muted-foreground">
                              Account Holder
                            </div>
                            <div class="font-semibold">
                              {{ account.accountHolder || 'Not provided' }}
                            </div>
                          </div>
                        </div>
                      </AccordionContent>
                    </AccordionItem>
                  </Accordion>
                </CardContent>
                <CardContent v-else>
                  <Alert>
                    <AlertCircle class="w-5 h-5" />
                    <AlertTitle>No Bank Accounts Available</AlertTitle>
                  </Alert>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Mobile Money Accounts</CardTitle>
                  <CardDescription>Mobile Money Account Information</CardDescription>
                </CardHeader>
                <CardContent v-if="customers.momoAccounts && customers.momoAccounts.length > 0">
                  <Accordion type="single" class="w-full" collapsible>
                    <AccordionItem
                      v-for="(account, index) in customers.momoAccounts"
                      :key="account.id"
                      class="border-b"
                      :value="`momo-${index}`"
                    >
                      <AccordionTrigger class="hover:no-underline">
                        <div class="flex items-center gap-4 flex-1">
                          <Avatar>
                            <AvatarFallback>
                              {{ account.networkName ? account.networkName[0].toUpperCase() : 'M' }}
                            </AvatarFallback>
                          </Avatar>
                          <div class="text-left">
                            <div class="font-semibold">
                              {{ account.networkName || 'Unknown Network' }}
                            </div>
                          </div>
                          <div class="ml-auto text-sm">
                            Mobile #: {{ account.momoNumber }}
                          </div>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <div class="grid gap-4 p-4 md:grid-cols-2">
                          <div>
                            <div class="text-sm text-muted-foreground">
                              Account Holder
                            </div>
                            <div class="font-semibold">
                              {{ account.accountHolder || 'Not provided' }}
                            </div>
                          </div>
                        </div>
                      </AccordionContent>
                    </AccordionItem>
                  </Accordion>
                </CardContent>
                <CardContent v-else>
                  <Alert>
                    <AlertCircle class="w-5 h-5" />
                    <AlertTitle>No Mobile Money Accounts Available</AlertTitle>
                  </Alert>
                </CardContent>
              </Card>
            </TabsContent>

            <!-- Credit Tab -->
            <TabsContent value="credit">
              <Card class="mb-4">
                <CardHeader>
                  <CardTitle>Loans</CardTitle>
                  <CardDescription>Loan Account Information</CardDescription>
                </CardHeader>
                <CardContent>
                  <Alert>
                    <AlertCircle class="w-5 h-5" />
                    <AlertTitle>No Loans Available</AlertTitle>
                  </Alert>
                </CardContent>
              </Card>
            </TabsContent>

            <!-- History Tab -->
            <TabsContent value="history">
              <Card>
                <CardHeader>
                  <CardTitle>Transaction History</CardTitle>
                  <CardDescription>Coming soon...</CardDescription>
                </CardHeader>
                <CardContent>
                  <div class="text-center text-muted-foreground py-8">
                    No history data available
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <!-- Compliance Tab -->
            <TabsContent value="compliance">
              <Card>
                <CardHeader>
                  <CardTitle>Compliance Information</CardTitle>
                  <CardDescription>Coming soon...</CardDescription>
                </CardHeader>
                <CardContent>
                  <div class="text-center text-muted-foreground py-8">
                    No compliance data available
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.indicator {
  box-shadow: 0 0 5px rgba(0, 0, 0, 0.2);
}
</style>
