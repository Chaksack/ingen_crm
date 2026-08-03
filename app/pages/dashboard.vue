<script setup lang="ts">
import NumberFlow from '@number-flow/vue'
import { TrendingDown, TrendingUp, TrendingUpIcon } from 'lucide-vue-next'

const dataCard = ref({
  totalRevenue: 0,
  newCustomers: 0,
  activeAccount: 0,
  growthRate: 0,
})

onMounted(() => {
  dataCard.value = {
    totalRevenue: 1250.44,
    newCustomers: 1234,
    activeAccount: 45678,
    growthRate: 4.5,
  }
})

const timeRange = ref('30d')

const isDesktop = useMediaQuery('(min-width: 768px)')
watch(isDesktop, () => {
  if (isDesktop.value) {
    timeRange.value = '30d'
  }
  else {
    timeRange.value = '7d'
  }
}, { immediate: true })
</script>

<template>
  <div class="w-full flex flex-col gap-4">
    <div class="flex flex-wrap items-center justify-between gap-2">
      <h2 class="text-2xl font-bold tracking-tight">
        Dashboard
      </h2>
      <div class="flex items-center space-x-2">
        <BaseDateRangePicker />
        <Button>Download</Button>
      </div>
    </div>
    <main class="@container/main flex flex-1 flex-col gap-4 md:gap-8">
      <div class="grid grid-cols-1 gap-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:shadow-xs @xl/main:grid-cols-2 @5xl/main:grid-cols-4">
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>Total Revenue</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <NumberFlow
                :value="dataCard.totalRevenue"
                :format="{ style: 'currency', currency: 'USD', trailingZeroDisplay: 'stripIfInteger' }"
              />
            </CardTitle>
            <CardAction>
              <Badge variant="outline">
                <TrendingUpIcon />
                +12.5%
              </Badge>
            </CardAction>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              Trending up this month <TrendingUp class="size-4" />
            </div>
            <div class="text-muted-foreground">
              Visitors for the last 6 months
            </div>
          </CardFooter>
        </Card>
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>New Customers</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <NumberFlow
                :value="dataCard.newCustomers"
              />
            </CardTitle>
            <CardAction>
              <Badge variant="outline">
                <TrendingDown />
                -20%
              </Badge>
            </CardAction>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              Down 20% this period <TrendingDown class="size-4" />
            </div>
            <div class="text-muted-foreground">
              Acquisition needs attention
            </div>
          </CardFooter>
        </Card>
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>Active Accounts</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <NumberFlow
                :value="dataCard.activeAccount"
              />
            </CardTitle>
            <CardAction>
              <Badge variant="outline">
                <TrendingUp />
                +12.5%
              </Badge>
            </CardAction>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              Strong user retention <TrendingUp class="size-4" />
            </div>
            <div class="text-muted-foreground">
              Engagement exceed targets
            </div>
          </CardFooter>
        </Card>
        <Card class="@container/card">
          <CardHeader>
            <CardDescription>Growth Rate</CardDescription>
            <CardTitle class="text-2xl font-semibold tabular-nums @[250px]/card:text-3xl">
              <NumberFlow
                :value="dataCard.growthRate"
                suffix="%"
              />
            </CardTitle>
            <CardAction>
              <Badge variant="outline">
                <TrendingUp />
                +4.5%
              </Badge>
            </CardAction>
          </CardHeader>
          <CardFooter class="flex-col items-start gap-1.5 text-sm">
            <div class="line-clamp-1 flex gap-2 font-medium">
              Steady performance increase <TrendingUp class="size-4" />
            </div>
            <div class="text-muted-foreground">
              Meets growth projections
            </div>
          </CardFooter>
        </Card>
      </div>
            <div class="grid grid-cols-1 gap-4 *:data-[slot=card]:bg-gradient-to-t *:data-[slot=card]:shadow-xs @xl/main:grid-cols-2">
              <Card class="@container/card">
        <CardHeader>
          <CardTitle>Total Credit Checks</CardTitle>
          <CardDescription>
            <span className="hidden @[540px]/card:block">
              Total for the last 3 months
            </span>
            <span className="@[540px]/card:hidden">Last 3 months</span>
          </CardDescription>
          <CardAction>
            <ToggleGroup
              v-model="timeRange"
              type="single"
              variant="outline"
              class="hidden *:data-[slot=toggle-group-item]:!px-4 @[767px]/card:flex"
            >
              <ToggleGroupItem value="90d">
                Last 3 months
              </ToggleGroupItem>
              <ToggleGroupItem value="30d">
                Last 30 days
              </ToggleGroupItem>
              <ToggleGroupItem value="7d">
                Last 7 days
              </ToggleGroupItem>
            </ToggleGroup>
            <Select v-model="timeRange">
              <SelectTrigger
                class="flex w-40 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate @[767px]/card:hidden"
                size="sm"
                aria-label="Select a value"
              >
                <SelectValue placeholder="Last 3 months" />
              </SelectTrigger>
              <SelectContent class="rounded-xl">
                <SelectItem value="90d" class="rounded-lg">
                  Last 3 months
                </SelectItem>
                <SelectItem value="30d" class="rounded-lg">
                  Last 30 days
                </SelectItem>
                <SelectItem value="7d" class="rounded-lg">
                  Last 7 days
                </SelectItem>
              </SelectContent>
            </Select>
          </CardAction>
        </CardHeader>
        <CardContent>
          <DashboardTotalVisitors :time-range="timeRange" />
        </CardContent>
      </Card>
      <Card class="@container/card">
        <CardHeader>
          <CardTitle>Total Loan Approvals</CardTitle>
          <CardDescription>
            <span className="hidden @[540px]/card:block">
              Total for the last 3 months
            </span>
            <span className="@[540px]/card:hidden">Last 3 months</span>
          </CardDescription>
          <CardAction>
            <ToggleGroup
              v-model="timeRange"
              type="single"
              variant="outline"
              class="hidden *:data-[slot=toggle-group-item]:!px-4 @[767px]/card:flex"
            >
              <ToggleGroupItem value="90d">
                Last 3 months
              </ToggleGroupItem>
              <ToggleGroupItem value="30d">
                Last 30 days
              </ToggleGroupItem>
              <ToggleGroupItem value="7d">
                Last 7 days
              </ToggleGroupItem>
            </ToggleGroup>
            <Select v-model="timeRange">
              <SelectTrigger
                class="flex w-40 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate @[767px]/card:hidden"
                size="sm"
                aria-label="Select a value"
              >
                <SelectValue placeholder="Last 3 months" />
              </SelectTrigger>
              <SelectContent class="rounded-xl">
                <SelectItem value="90d" class="rounded-lg">
                  Last 3 months
                </SelectItem>
                <SelectItem value="30d" class="rounded-lg">
                  Last 30 days
                </SelectItem>
                <SelectItem value="7d" class="rounded-lg">
                  Last 7 days
                </SelectItem>
              </SelectContent>
            </Select>
          </CardAction>
        </CardHeader>
        <CardContent>
          <DashboardTotalVisitors :time-range="timeRange" />
        </CardContent>
      </Card>
</div>
      <Card class="@container/card">
        <CardHeader>
          <CardTitle>Total Visitors</CardTitle>
          <CardDescription>
            <span className="hidden @[540px]/card:block">
              Total for the last 3 months
            </span>
            <span className="@[540px]/card:hidden">Last 3 months</span>
          </CardDescription>
          <CardAction>
            <ToggleGroup
              v-model="timeRange"
              type="single"
              variant="outline"
              class="hidden *:data-[slot=toggle-group-item]:!px-4 @[767px]/card:flex"
            >
              <ToggleGroupItem value="90d">
                Last 3 months
              </ToggleGroupItem>
              <ToggleGroupItem value="30d">
                Last 30 days
              </ToggleGroupItem>
              <ToggleGroupItem value="7d">
                Last 7 days
              </ToggleGroupItem>
            </ToggleGroup>
            <Select v-model="timeRange">
              <SelectTrigger
                class="flex w-40 **:data-[slot=select-value]:block **:data-[slot=select-value]:truncate @[767px]/card:hidden"
                size="sm"
                aria-label="Select a value"
              >
                <SelectValue placeholder="Last 3 months" />
              </SelectTrigger>
              <SelectContent class="rounded-xl">
                <SelectItem value="90d" class="rounded-lg">
                  Last 3 months
                </SelectItem>
                <SelectItem value="30d" class="rounded-lg">
                  Last 30 days
                </SelectItem>
                <SelectItem value="7d" class="rounded-lg">
                  Last 7 days
                </SelectItem>
              </SelectContent>
            </Select>
          </CardAction>
        </CardHeader>
        <CardContent>
          <DashboardTotalVisitors :time-range="timeRange" />
        </CardContent>
      </Card>
      <Tabs default-value="users">
            <div class="flex items-center">
              <TabsList>
                <TabsTrigger value="users">
                  Users
                </TabsTrigger>
                <TabsTrigger value="business">
                  Business
                </TabsTrigger>
              </TabsList>
              <div class="ml-auto flex items-center gap-2">
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="outline" size="sm" class="h-7 gap-1 rounded-md px-3">
                      <ListFilter class="h-3.5 w-3.5" />
                      <span class="sr-only sm:not-sr-only">Filter</span>
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuLabel>Filter by</DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem>
                      <div class="items-top flex space-x-2">
                        <Checkbox id="terms1" />
                        <label
                          for="terms2"
                          class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                          Fulfilled
                        </label>
                      </div>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <div class="items-top flex space-x-2">
                        <Checkbox id="terms1" />
                        <label
                          for="terms2"
                          class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                          Declined
                        </label>
                      </div>
                    </DropdownMenuItem>
                    <DropdownMenuItem>
                      <div class="items-top flex space-x-2">
                        <Checkbox id="terms1" />
                        <label
                          for="terms2"
                          class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                        >
                          Refunded
                        </label>
                      </div>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
                <Button variant="outline" size="sm" class="h-7 gap-1 rounded-md px-3">
                  <File class="h-3.5 w-3.5" />
                  <span class="sr-only sm:not-sr-only">Export</span>
                </Button>
              </div>
            </div>
            <TabsContent value="users">
              <Card>
                <CardHeader class="px-7">
                  <CardTitle>Users</CardTitle>
                  <CardDescription>
                    Recent users registered.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Customer</TableHead>
                        <TableHead class="hidden sm:table-cell">
                          Type
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
                    <TableBody>
                      <TableRow class="bg-accent">
                        <TableCell>
                          <div class="font-medium">
                            Liam Johnson
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            liam@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-23
                        </TableCell>
                        <TableCell class="text-right">
                          $250.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Olivia Smith
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            olivia@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Refund
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="outline">
                            Declined
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-24
                        </TableCell>
                        <TableCell class="text-right">
                          $150.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Noah Williams
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            noah@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Subscription
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-25
                        </TableCell>
                        <TableCell class="text-right">
                          $350.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Emma Brown
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            emma@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-26
                        </TableCell>
                        <TableCell class="text-right">
                          $450.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Liam Johnson
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            liam@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-23
                        </TableCell>
                        <TableCell class="text-right">
                          $250.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Liam Johnson
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            liam@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-23
                        </TableCell>
                        <TableCell class="text-right">
                          $250.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Olivia Smith
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            olivia@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Refund
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="outline">
                            Declined
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-24
                        </TableCell>
                        <TableCell class="text-right">
                          $150.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Emma Brown
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            emma@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-26
                        </TableCell>
                        <TableCell class="text-right">
                          $450.00
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="business">
              <Card>
                <CardHeader class="px-7">
                  <CardTitle>Business</CardTitle>
                  <CardDescription>
                    Recent business registered.
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Customer</TableHead>
                        <TableHead class="hidden sm:table-cell">
                          Type
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
                    <TableBody>
                      <TableRow class="bg-accent">
                        <TableCell>
                          <div class="font-medium">
                            Liam Johnson
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            liam@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-23
                        </TableCell>
                        <TableCell class="text-right">
                          $250.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Olivia Smith
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            olivia@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Refund
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="outline">
                            Declined
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-24
                        </TableCell>
                        <TableCell class="text-right">
                          $150.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Noah Williams
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            noah@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Subscription
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-25
                        </TableCell>
                        <TableCell class="text-right">
                          $350.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Emma Brown
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            emma@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-26
                        </TableCell>
                        <TableCell class="text-right">
                          $450.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Liam Johnson
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            liam@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-23
                        </TableCell>
                        <TableCell class="text-right">
                          $250.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Liam Johnson
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            liam@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-23
                        </TableCell>
                        <TableCell class="text-right">
                          $250.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Olivia Smith
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            olivia@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Refund
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="outline">
                            Declined
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-24
                        </TableCell>
                        <TableCell class="text-right">
                          $150.00
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableCell>
                          <div class="font-medium">
                            Emma Brown
                          </div>
                          <div class="hidden text-sm text-muted-foreground md:inline">
                            emma@example.com
                          </div>
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          Sale
                        </TableCell>
                        <TableCell class="hidden sm:table-cell">
                          <Badge class="text-xs" variant="secondary">
                            Fulfilled
                          </Badge>
                        </TableCell>
                        <TableCell class="hidden md:table-cell">
                          2023-06-26
                        </TableCell>
                        <TableCell class="text-right">
                          $450.00
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
