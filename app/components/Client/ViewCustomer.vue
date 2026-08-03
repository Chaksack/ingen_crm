<script setup lang="ts">
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { toast } from 'vue-sonner'
import {
  ChevronLeft,
  PlusCircle,
  CircleCheck,
  TriangleAlert,
  MapPinned,
  Phone,
  AtSign,
  Banknote,
  AlertCircle,
} from 'lucide-vue-next'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'

import { ref, computed, onMounted } from 'vue'
import { Avatar, AvatarFallback, AvatarImage } from '~/components/ui/avatar'
import { useRoute } from 'vue-router'

const errorMessage = ref('')
const creditScore = ref(0)
const maxCreditScore = 1000
const customers = ref<any>({})
const isLoading = ref(false)
const error = ref<string | null>(null)
const route = useRoute()

// Calculate progress percentage
const progress = computed(() => {
  const score = customers.value?.credit_score || creditScore.value || 0
  return (score / maxCreditScore) * 100
})

// Calculate stroke dash offset based on progress (for circular progress bar)
const offset = computed(() => {
  const circumference = 251.2 // 2 * π * radius
  return circumference - (circumference * progress.value) / 100
})

// Function to fetch user data from the API
const fetchUser = async () => {
  isLoading.value = true
  error.value = null
  try {
    const data = await $fetch(`http://localhost:8080/api/users/${route.params.id}`, {
      method: 'GET',
      credentials: 'include',
    })
    console.log('Fetched user data:', data)
    customers.value = data

    // Update credit score from API if available
    if (data?.credit_score) {
      creditScore.value = data.credit_score
    }
  } catch (err: any) {
    console.error('Failed to fetch user details:', err)
    error.value = err?.message || 'Failed to load user details'
    toast.error('Error', { description: error.value })
  } finally {
    isLoading.value = false
  }
}

// Badge variant function based on status
const getBadgeVariant = (status: string) => {
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
const getKycVariant = (kyc: boolean) => {
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
      <header class="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6">
      </header>
      
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
            <NuxtLink to="/customer">
              <Button variant="outline" size="icon" class="h-7 w-7">
                <ChevronLeft class="h-4 w-4" />
                <span class="sr-only">Back</span>
              </Button>
            </NuxtLink>
            
            <Avatar class="relative overflow-visible">
              <AvatarImage class="rounded-full" :src="customers.avatar || ''" alt="Customer Avatar" />
              <AvatarFallback class="text-white">
                {{ customers.first_name && customers.last_name ? `${customers.first_name[0]}${customers.last_name[0]}` : '' }}
              </AvatarFallback>
              <span
                  v-if="customers.status"
                  :class="(customers.status === 'active') ? 'bg-green-500' : 'bg-red-500'"
                  class="absolute bottom-[-4px] right-[-4px] w-3.5 h-3.5 rounded-full border-2 border-white"
              ></span>
            </Avatar>
            
            <div class="flex shrink-0 whitespace-nowrap text-xl font-semibold tracking-tight sm:grow-0">
              <h3>{{ customers.first_name }} {{ customers.last_name }}</h3>
              <p v-if="customers.ghcard" class="mx-2 mt-2 font-semibold text-sm"> - {{ customers.ghcard }}</p>
            </div>

            <Badge :variant="getBadgeVariant(customers.status)">{{ customers.status || 'N/A' }}</Badge>
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
                    <Button type="submit">Save changes</Button>
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
                    <Button type="submit">Save changes</Button>
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
              {{ customers.phone_number || 'N/A' }}
            </h3>
            <h3 class="flex items-center">
              <AtSign class="w-6 h-6 mr-2" />
              {{ customers.email || 'N/A' }}
            </h3>
            <h3 class="flex items-center">
              <Banknote class="w-6 h-6 mr-2" />
              £ {{ customers.monthly_income || '0' }}
            </h3>
          </div>

          <!-- Stats Cards -->
          <div v-if="!isLoading && customers && Object.keys(customers).length > 0" class="grid mt-6 gap-4 sm:grid-cols-2 md:grid-cols-4 lg:grid-cols-5">
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">Credit Rating</CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  Good
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">Credit Score</CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  {{ customers.credit_score || 0 }}/1000
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">Credit Limit</CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  £ 16,399,000
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">Total Assets</CardDescription>
                <CardTitle class="text-2xl font-bold text-green-600">
                  £ 254,399,000
                </CardTitle>
              </CardHeader>
            </Card>
            <Card>
              <CardHeader class="pb-2">
                <CardDescription class="font-bold">Turnover</CardDescription>
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
                <TabsTrigger value="score">Score</TabsTrigger>
                <TabsTrigger value="accounts">Accounts</TabsTrigger>
                <TabsTrigger value="credit">Credits & loans</TabsTrigger>
                <TabsTrigger value="history">History</TabsTrigger>
                <TabsTrigger value="compliance">Compliance</TabsTrigger>
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
                      <div class="text-4xl font-semibold mb-4">{{ customers.credit_score || 0 }} pts</div>
                      <div class="relative w-full bg-gray-300 rounded-full h-4 overflow-hidden">
                        <div
                            class="bg-gradient-to-r from-red-500 via-yellow-400 to-green-500 h-4 rounded-full transition-all duration-300 ease-in-out"
                            :style="{ width: progress + '%' }"
                        >
                          <div
                              class="absolute indicator top-1/2 transform -translate-y-1/2 w-5 h-5 bg-white border-2 border-black rounded-full flex items-center justify-center transition-all duration-300 ease-in-out"
                              :style="{ left: `calc(${progress}% - 12px)` }"
                          ></div>
                        </div>
                      </div>
                      <div class="mt-2 text-sm">{{ customers.credit_score || 0 }} / 1000 pts</div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardContent class="pt-6">
                    <div class="grid gap-4 sm:grid-cols-1 md:grid-cols-3">
                      <Card>
                        <CardHeader class="pb-2">
                          <CardDescription>Current Rating</CardDescription>
                          <CardTitle class="text-xl">Good</CardTitle>
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
                          <CardTitle class="text-xl">Silver</CardTitle>
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
                          <CardTitle class="text-xl">77 - Very Likely</CardTitle>
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
                <CardContent v-if="customers.bank_details && customers.bank_details.length > 0">
                  <Accordion type="single" class="w-full" collapsible>
                    <AccordionItem
                        v-for="(bankDetail, index) in customers.bank_details"
                        :key="bankDetail.ID"
                        class="border-b"
                        :value="'bank-' + index"
                    >
                      <AccordionTrigger class="hover:no-underline">
                        <div class="flex items-center gap-4 flex-1">
                          <Avatar>
                            <AvatarImage :src="bankDetail.bank?.bank_logo || ''" alt="Bank Logo" />
                            <AvatarFallback>
                              {{ bankDetail.bank?.bank_name ? bankDetail.bank.bank_name[0] : 'B' }}
                            </AvatarFallback>
                          </Avatar>
                          <div class="text-left">
                            <div class="font-semibold">{{ bankDetail.bank?.bank_name || 'Unknown Bank' }}</div>
                            <div class="text-sm text-muted-foreground">{{ bankDetail.account_type || 'No account type' }}</div>
                          </div>
                          <div class="ml-auto text-sm">
                            Account #: {{ bankDetail.account_number }}
                          </div>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <div class="grid gap-4 p-4 md:grid-cols-2">
                          <div>
                            <div class="text-sm text-muted-foreground">Account Holder</div>
                            <div class="font-semibold">{{ bankDetail.account_holder || 'Not provided' }}</div>
                          </div>
                          <div>
                            <div class="text-sm text-muted-foreground">Bank Prefix</div>
                            <div class="font-semibold">{{ bankDetail.bank?.bank_prefix || 'N/A' }}</div>
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
                <CardContent v-if="customers.momo_details && customers.momo_details.length > 0">
                  <Accordion type="single" class="w-full" collapsible>
                    <AccordionItem
                        v-for="(momoDetail, index) in customers.momo_details"
                        :key="momoDetail.ID"
                        class="border-b"
                        :value="'momo-' + index"
                    >
                      <AccordionTrigger class="hover:no-underline">
                        <div class="flex items-center gap-4 flex-1">
                          <Avatar>
                            <AvatarImage :src="momoDetail.momo?.network_logo || ''" alt="Network Logo" />
                            <AvatarFallback>
                              {{ momoDetail.momo?.network_prefix || 'M' }}
                            </AvatarFallback>
                          </Avatar>
                          <div class="text-left">
                            <div class="font-semibold">{{ momoDetail.momo?.network || 'Unknown Network' }}</div>
                            <div class="text-sm text-muted-foreground">{{ momoDetail.momo_type || 'No type' }}</div>
                          </div>
                          <div class="ml-auto text-sm">
                            Mobile #: {{ momoDetail.mobile_number }}
                          </div>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <div class="grid gap-4 p-4 md:grid-cols-2">
                          <div>
                            <div class="text-sm text-muted-foreground">Account Holder</div>
                            <div class="font-semibold">{{ momoDetail.account_holder || 'Not provided' }}</div>
                          </div>
                          <div>
                            <div class="text-sm text-muted-foreground">Network Prefix</div>
                            <div class="font-semibold">{{ momoDetail.momo?.network_prefix || 'N/A' }}</div>
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
                <CardContent v-if="customers.loans && customers.loans.length > 0">
                  <Accordion type="single" class="w-full" collapsible>
                    <AccordionItem
                        v-for="(loan, index) in customers.loans"
                        :key="loan.ID"
                        class="border-b"
                        :value="'loan-' + index"
                    >
                      <AccordionTrigger class="hover:no-underline">
                        <div class="flex items-center gap-4 flex-1">
                          <Avatar>
                            <AvatarImage :src="loan.business?.avatar || ''" alt="Loan Avatar" />
                            <AvatarFallback>
                              {{ loan.created_by ? loan.created_by.substring(0, 2).toUpperCase() : 'L' }}
                            </AvatarFallback>
                          </Avatar>
                          <div class="text-left">
                            <div class="font-semibold">{{ loan.created_by || 'Unknown' }}</div>
                            <div class="text-sm text-muted-foreground">{{ loan.loan_type || 'No type' }}</div>
                          </div>
                          <div class="ml-auto font-semibold">
                            £ {{ loan.loan_amount }}
                          </div>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <div class="p-4 space-y-4">
                          <div class="flex items-center gap-2 text-sm text-muted-foreground">
                            <CircleCheck :class="loan.loan_status ? 'text-green-500' : 'text-yellow-500'" />
                            {{ loan.loan_status ? 'Paid' : 'Ongoing' }}
                          </div>
                          <div class="grid gap-4 md:grid-cols-3">
                            <div>
                              <div class="text-sm text-muted-foreground">Loan Term</div>
                              <div class="font-semibold">{{ loan.loan_term }} months</div>
                            </div>
                            <div>
                              <div class="text-sm text-muted-foreground">Purpose</div>
                              <div class="font-semibold">{{ loan.loan_purpose || 'N/A' }}</div>
                            </div>
                            <div>
                              <div class="text-sm text-muted-foreground">Start Date</div>
                              <div class="font-semibold">{{ loan.start_date || 'N/A' }}</div>
                            </div>
                            <div>
                              <div class="text-sm text-muted-foreground">End Date</div>
                              <div class="font-semibold">{{ loan.end_date || 'N/A' }}</div>
                            </div>
                            <div>
                              <div class="text-sm text-muted-foreground">Average Amount</div>
                              <div class="font-semibold">£ {{ loan.average_loan_amount || '0' }}</div>
                            </div>
                            <div>
                              <div class="text-sm text-muted-foreground">Amount Owed</div>
                              <div class="font-semibold text-red-600">£ {{ loan.total_amount_owed || '0' }}</div>
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