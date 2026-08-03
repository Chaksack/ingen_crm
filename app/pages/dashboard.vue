<script setup lang="ts">
import NumberFlow from '@number-flow/vue'
import { TrendingDown, TrendingUp } from 'lucide-vue-next'

interface RecentInvoice {
  id: string
  invoiceNumber: string
  clientName: string
  status: string
  total: string
  issueDate: string
}

interface RecentTicket {
  id: string
  ticketNumber: string
  subject: string
  name: string
  priority: string
  status: string
  createdAt: string
}

interface DashboardSummary {
  totalRevenue: number
  revenueThisMonth: number
  revenueLastMonth: number
  outstandingBalance: number
  totalClients: number
  newClientsThisMonth: number
  newClientsLastMonth: number
  openTickets: number
  openTicketsLastMonth: number
  revenueTrend: { label: string, invoiced: number, collected: number }[]
  pipeline: { draft: number, outstanding: number, paid: number, overdue: number }
  recentInvoices: RecentInvoice[]
  recentTickets: RecentTicket[]
}

const { categorical, status: statusColors } = useChartPalette()

const timeRange = ref('30d')
const isDesktop = useMediaQuery('(min-width: 768px)')
watch(isDesktop, () => {
  timeRange.value = isDesktop.value ? '30d' : '7d'
}, { immediate: true })

const { data: summary, pending: isLoading } = await useFetch<DashboardSummary>('/api/dashboard/summary', {
  query: { range: timeRange },
})

function formatMoney(value: number | undefined) {
  return new Intl.NumberFormat('en-GH', { style: 'currency', currency: 'GHS', maximumFractionDigits: 0 }).format(value ?? 0)
}

function percentChange(current: number, previous: number) {
  if (previous === 0)
    return current > 0 ? 100 : 0
  return Math.round(((current - previous) / previous) * 100)
}

const revenueChange = computed(() => percentChange(summary.value?.revenueThisMonth ?? 0, summary.value?.revenueLastMonth ?? 0))
const clientsChange = computed(() => percentChange(summary.value?.newClientsThisMonth ?? 0, summary.value?.newClientsLastMonth ?? 0))
const ticketsChange = computed(() => percentChange(summary.value?.openTickets ?? 0, summary.value?.openTicketsLastMonth ?? 0))

const pipelineData = computed(() => {
  const p = summary.value?.pipeline
  if (!p)
    return []
  return [
    { label: 'Draft', value: p.draft, color: statusColors.value.muted },
    { label: 'Outstanding', value: p.outstanding, color: statusColors.value.warning },
    { label: 'Paid', value: p.paid, color: statusColors.value.good },
    { label: 'Overdue', value: p.overdue, color: statusColors.value.critical },
  ].filter(d => d.value > 0)
})

function invoiceBadgeVariant(status: string) {
  switch (status) {
    case 'paid': return 'default'
    case 'overdue': return 'destructive'
    case 'void': return 'secondary'
    default: return 'secondary'
  }
}

function ticketBadgeVariant(status: string) {
  switch (status) {
    case 'open': return 'default'
    case 'resolved': return 'outline'
    case 'closed': return 'secondary'
    default: return 'secondary'
  }
}
</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-2xl font-bold tracking-tight">
        Dashboard
      </h2>
      <div class="flex items-center space-x-2">
        <ToggleGroup
          v-model="timeRange"
          type="single"
          variant="outline"
          class="hidden *:data-[slot=toggle-group-item]:!px-4 sm:flex"
        >
          <ToggleGroupItem value="90d">
            Last 90 days
          </ToggleGroupItem>
          <ToggleGroupItem value="30d">
            Last 30 days
          </ToggleGroupItem>
          <ToggleGroupItem value="7d">
            Last 7 days
          </ToggleGroupItem>
        </ToggleGroup>
        <Select v-model="timeRange">
          <SelectTrigger class="w-40 sm:hidden" size="sm" aria-label="Select a time range">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="90d">
              Last 90 days
            </SelectItem>
            <SelectItem value="30d">
              Last 30 days
            </SelectItem>
            <SelectItem value="7d">
              Last 7 days
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <main class="@container/main flex flex-1 flex-col gap-4 md:gap-8">
      <div class="grid grid-cols-1 gap-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:shadow-xs @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>Total Revenue</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <Skeleton v-if="isLoading" class="h-8 w-32" />
              <span v-else>{{ formatMoney(summary?.totalRevenue) }}</span>
            </CardTitle>
            <CardAction>
              <Badge v-if="!isLoading" variant="outline">
                <component :is="revenueChange >= 0 ? TrendingUp : TrendingDown" />
                {{ revenueChange >= 0 ? '+' : '' }}{{ revenueChange }}%
              </Badge>
            </CardAction>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              This month: {{ formatMoney(summary?.revenueThisMonth) }}
            </div>
            <div class="text-muted-foreground">
              Total payments collected
            </div>
          </CardFooter>
        </Card>
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>Outstanding Balance</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <Skeleton v-if="isLoading" class="h-8 w-32" />
              <span v-else>{{ formatMoney(summary?.outstandingBalance) }}</span>
            </CardTitle>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              Across sent &amp; overdue invoices
            </div>
            <div class="text-muted-foreground">
              Unpaid balance owed by clients
            </div>
          </CardFooter>
        </Card>
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>Total Clients</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <Skeleton v-if="isLoading" class="h-8 w-20" />
              <NumberFlow v-else :value="summary?.totalClients ?? 0" />
            </CardTitle>
            <CardAction>
              <Badge v-if="!isLoading" variant="outline">
                <component :is="clientsChange >= 0 ? TrendingUp : TrendingDown" />
                {{ clientsChange >= 0 ? '+' : '' }}{{ clientsChange }}%
              </Badge>
            </CardAction>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              {{ summary?.newClientsThisMonth ?? 0 }} new this month
            </div>
            <div class="text-muted-foreground">
              Onboarded clients
            </div>
          </CardFooter>
        </Card>
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>Open Support Tickets</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <Skeleton v-if="isLoading" class="h-8 w-16" />
              <NumberFlow v-else :value="summary?.openTickets ?? 0" />
            </CardTitle>
            <CardAction>
              <Badge v-if="!isLoading" variant="outline">
                <component :is="ticketsChange <= 0 ? TrendingDown : TrendingUp" />
                {{ ticketsChange >= 0 ? '+' : '' }}{{ ticketsChange }}%
              </Badge>
            </CardAction>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              Awaiting a response
            </div>
            <div class="text-muted-foreground">
              <NuxtLink to="/customer-support" class="underline underline-offset-4">
                View support queue
              </NuxtLink>
            </div>
          </CardFooter>
        </Card>
      </div>

      <div class="grid grid-cols-1 gap-4 @4xl/main:grid-cols-3">
        <Card class="@container/card @4xl/main:col-span-2">
          <CardHeader>
            <CardTitle>Revenue Trend</CardTitle>
            <CardDescription>
              Invoiced vs. collected
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-[300px] w-full" />
            <AreaChart
              v-else
              :data="summary?.revenueTrend ?? []"
              :categories="['invoiced', 'collected']"
              index="label"
              :colors="[categorical[0], categorical[1]]"
              :y-formatter="(v: number) => formatMoney(v)"
            />
          </CardContent>
        </Card>

        <Card class="@container/card">
          <CardHeader>
            <CardTitle>Invoice Pipeline</CardTitle>
            <CardDescription>
              By status
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Skeleton v-if="isLoading" class="h-48 w-full" />
            <div v-else-if="pipelineData.length === 0" class="flex h-48 items-center justify-center text-sm text-muted-foreground">
              No invoices yet
            </div>
            <DonutChart
              v-else
              :data="pipelineData"
              category="value"
              index="label"
              :colors="pipelineData.map(d => d.color)"
            />
            <div v-if="!isLoading && pipelineData.length" class="mt-2 grid grid-cols-2 gap-2 text-sm">
              <div v-for="d in pipelineData" :key="d.label" class="flex items-center gap-2">
                <span class="size-2.5 rounded-full" :style="{ backgroundColor: d.color }" />
                <span class="text-muted-foreground">{{ d.label }}</span>
                <span class="ml-auto font-medium">{{ d.value }}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs default-value="invoices">
        <TabsList>
          <TabsTrigger value="invoices">
            Recent Invoices
          </TabsTrigger>
          <TabsTrigger value="tickets">
            Recent Support Tickets
          </TabsTrigger>
        </TabsList>

        <TabsContent value="invoices">
          <Card>
            <CardHeader class="px-7">
              <CardTitle>Recent Invoices</CardTitle>
              <CardDescription>
                The latest invoices issued to clients.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Client</TableHead>
                    <TableHead class="hidden sm:table-cell">
                      Invoice #
                    </TableHead>
                    <TableHead class="hidden sm:table-cell">
                      Status
                    </TableHead>
                    <TableHead class="hidden md:table-cell">
                      Date
                    </TableHead>
                    <TableHead class="text-right">
                      Amount
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody v-if="isLoading">
                  <TableRow v-for="i in 5" :key="`skeleton-${i}`">
                    <TableCell colspan="5">
                      <Skeleton class="h-6 w-full" />
                    </TableCell>
                  </TableRow>
                </TableBody>
                <TableBody v-else>
                  <TableRow v-for="invoice in summary?.recentInvoices" :key="invoice.id">
                    <TableCell>
                      <NuxtLink :to="`/finance/invoices/${invoice.id}`" class="font-medium hover:underline">
                        {{ invoice.clientName }}
                      </NuxtLink>
                    </TableCell>
                    <TableCell class="hidden sm:table-cell">
                      {{ invoice.invoiceNumber }}
                    </TableCell>
                    <TableCell class="hidden sm:table-cell">
                      <Badge class="text-xs" :variant="invoiceBadgeVariant(invoice.status)">
                        {{ invoice.status.replace('_', ' ') }}
                      </Badge>
                    </TableCell>
                    <TableCell class="hidden md:table-cell">
                      {{ invoice.issueDate }}
                    </TableCell>
                    <TableCell class="text-right">
                      {{ formatMoney(Number(invoice.total)) }}
                    </TableCell>
                  </TableRow>
                  <TableRow v-if="!summary?.recentInvoices?.length">
                    <TableCell colspan="5" class="text-center text-muted-foreground">
                      No invoices yet.
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="tickets">
          <Card>
            <CardHeader class="px-7">
              <CardTitle>Recent Support Tickets</CardTitle>
              <CardDescription>
                The latest tickets submitted by clients.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Requester</TableHead>
                    <TableHead class="hidden sm:table-cell">
                      Subject
                    </TableHead>
                    <TableHead class="hidden sm:table-cell">
                      Status
                    </TableHead>
                    <TableHead class="text-right">
                      Priority
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody v-if="isLoading">
                  <TableRow v-for="i in 5" :key="`skeleton-${i}`">
                    <TableCell colspan="4">
                      <Skeleton class="h-6 w-full" />
                    </TableCell>
                  </TableRow>
                </TableBody>
                <TableBody v-else>
                  <TableRow v-for="ticket in summary?.recentTickets" :key="ticket.id">
                    <TableCell>
                      <div class="font-medium">
                        {{ ticket.name }}
                      </div>
                      <div class="hidden text-sm text-muted-foreground md:inline">
                        {{ ticket.ticketNumber }}
                      </div>
                    </TableCell>
                    <TableCell class="hidden sm:table-cell max-w-xs truncate">
                      {{ ticket.subject }}
                    </TableCell>
                    <TableCell class="hidden sm:table-cell">
                      <Badge class="text-xs" :variant="ticketBadgeVariant(ticket.status)">
                        {{ ticket.status.replace('_', ' ') }}
                      </Badge>
                    </TableCell>
                    <TableCell class="text-right">
                      {{ ticket.priority }}
                    </TableCell>
                  </TableRow>
                  <TableRow v-if="!summary?.recentTickets?.length">
                    <TableCell colspan="4" class="text-center text-muted-foreground">
                      No support tickets yet.
                    </TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </main>
  </div>
</template>
