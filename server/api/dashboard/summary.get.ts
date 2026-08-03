import { desc } from 'drizzle-orm'
import { db } from '../../db/client'
import { clients, payments, supportTickets } from '../../db/schema'

function startOfDay(d: Date) {
  const copy = new Date(d)
  copy.setHours(0, 0, 0, 0)
  return copy
}

function isThisMonth(dateStr: string, now: Date) {
  const d = new Date(dateStr)
  return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth()
}

function isLastMonth(dateStr: string, now: Date) {
  const last = new Date(now.getFullYear(), now.getMonth() - 1, 1)
  const d = new Date(dateStr)
  return d.getFullYear() === last.getFullYear() && d.getMonth() === last.getMonth()
}

interface Bucket {
  label: string
  start: Date
  end: Date
}

function buildBuckets(range: '7d' | '30d' | '90d', now: Date): Bucket[] {
  if (range === '90d') {
    // 13 weekly buckets
    const buckets: Bucket[] = []
    const end = startOfDay(now)
    for (let i = 12; i >= 0; i--) {
      const bucketEnd = new Date(end)
      bucketEnd.setDate(bucketEnd.getDate() - (i * 7))
      const bucketStart = new Date(bucketEnd)
      bucketStart.setDate(bucketStart.getDate() - 6)
      buckets.push({
        label: `${bucketStart.getMonth() + 1}/${bucketStart.getDate()}`,
        start: bucketStart,
        end: new Date(bucketEnd.getTime() + 24 * 60 * 60 * 1000 - 1),
      })
    }
    return buckets
  }

  const days = range === '7d' ? 7 : 30
  const buckets: Bucket[] = []
  const end = startOfDay(now)
  for (let i = days - 1; i >= 0; i--) {
    const day = new Date(end)
    day.setDate(day.getDate() - i)
    buckets.push({
      label: `${day.getMonth() + 1}/${day.getDate()}`,
      start: day,
      end: new Date(day.getTime() + 24 * 60 * 60 * 1000 - 1),
    })
  }
  return buckets
}

export default defineEventHandler(async (event) => {
  await requireUser(event)
  const query = getQuery(event)
  const range = (query.range === '7d' || query.range === '90d') ? query.range : '30d'

  const [allInvoices, allPayments, allClients, allTickets] = await Promise.all([
    db.query.invoices.findMany({ with: { client: true } }),
    db.select().from(payments),
    db.select().from(clients),
    db.select().from(supportTickets),
  ])

  const now = new Date()

  const totalRevenue = allPayments.reduce((sum, p) => sum + Number(p.amount), 0)
  const revenueThisMonth = allPayments.filter(p => isThisMonth(p.paidAt, now)).reduce((sum, p) => sum + Number(p.amount), 0)
  const revenueLastMonth = allPayments.filter(p => isLastMonth(p.paidAt, now)).reduce((sum, p) => sum + Number(p.amount), 0)

  const paymentsByInvoice = new Map<string, number>()
  for (const p of allPayments)
    paymentsByInvoice.set(p.invoiceId, (paymentsByInvoice.get(p.invoiceId) ?? 0) + Number(p.amount))

  const outstandingBalance = allInvoices
    .filter(i => i.status !== 'void' && i.status !== 'draft')
    .reduce((sum, i) => sum + Math.max(Number(i.total) - (paymentsByInvoice.get(i.id) ?? 0), 0), 0)

  const totalClients = allClients.length
  const newClientsThisMonth = allClients.filter(c => isThisMonth(c.createdAt.toISOString(), now)).length
  const newClientsLastMonth = allClients.filter(c => isLastMonth(c.createdAt.toISOString(), now)).length

  const openTickets = allTickets.filter(t => t.status === 'open' || t.status === 'in_progress').length
  const openTicketsLastMonth = allTickets.filter(t => (t.status === 'open' || t.status === 'in_progress') && isLastMonth(t.createdAt.toISOString(), now)).length

  const buckets = buildBuckets(range, now)
  const revenueTrend = buckets.map(({ label, start, end }) => {
    const invoiced = allInvoices
      .filter(i => i.status !== 'void')
      .filter((i) => {
        const d = new Date(i.issueDate)
        return d >= start && d <= end
      })
      .reduce((sum, i) => sum + Number(i.total), 0)
    const collected = allPayments
      .filter((p) => {
        const d = new Date(p.paidAt)
        return d >= start && d <= end
      })
      .reduce((sum, p) => sum + Number(p.amount), 0)
    return { label, invoiced, collected }
  })

  const activeInvoices = allInvoices.filter(i => i.status !== 'void')
  const pipeline = {
    draft: activeInvoices.filter(i => i.status === 'draft').length,
    outstanding: activeInvoices.filter(i => i.status === 'sent' || i.status === 'partially_paid').length,
    paid: activeInvoices.filter(i => i.status === 'paid').length,
    overdue: activeInvoices.filter(i => i.status === 'overdue').length,
  }

  const recentInvoices = [...allInvoices]
    .sort((a, b) => new Date(b.issueDate).getTime() - new Date(a.issueDate).getTime())
    .slice(0, 8)
    .map(i => ({
      id: i.id,
      invoiceNumber: i.invoiceNumber,
      clientName: i.client?.name ?? 'N/A',
      status: i.status,
      total: i.total,
      issueDate: i.issueDate,
    }))

  const recentTickets = await db.query.supportTickets.findMany({
    orderBy: desc(supportTickets.createdAt),
    limit: 8,
  })

  return {
    totalRevenue,
    revenueThisMonth,
    revenueLastMonth,
    outstandingBalance,
    totalClients,
    newClientsThisMonth,
    newClientsLastMonth,
    openTickets,
    openTicketsLastMonth,
    revenueTrend,
    pipeline,
    recentInvoices,
    recentTickets: recentTickets.map(t => ({
      id: t.id,
      ticketNumber: t.ticketNumber,
      subject: t.subject,
      name: t.name,
      priority: t.priority,
      status: t.status,
      createdAt: t.createdAt,
    })),
  }
})
