export interface InvoiceLike {
  total: string
  dueDate: string
  status: string
}

export function computeInvoiceView<T extends InvoiceLike>(invoice: T, payments: { amount: string }[]) {
  const total = Number(invoice.total)
  const amountPaid = payments.reduce((sum, p) => sum + Number(p.amount), 0)
  const balanceDue = Math.max(total - amountPaid, 0)

  let effectiveStatus = invoice.status
  if (invoice.status !== 'void') {
    if (balanceDue <= 0 && total > 0) {
      effectiveStatus = 'paid'
    }
    else if (new Date(invoice.dueDate) < new Date() && balanceDue > 0) {
      effectiveStatus = 'overdue'
    }
    else if (amountPaid > 0) {
      effectiveStatus = 'partially_paid'
    }
  }

  return { ...invoice, amountPaid, balanceDue, effectiveStatus }
}

export function computeLineItems(items: Array<{ description: string, quantity: number, unitPrice: number }>) {
  return items.map((item, index) => ({
    description: item.description,
    quantity: String(item.quantity),
    unitPrice: String(item.unitPrice),
    amount: String(Math.round(item.quantity * item.unitPrice * 100) / 100),
    order: index,
  }))
}

export function computeTotals(lineItems: Array<{ amount: string }>, discount: number) {
  const subtotal = Math.round(lineItems.reduce((sum, i) => sum + Number(i.amount), 0) * 100) / 100
  const taxableBase = Math.max(Math.round((subtotal - discount) * 100) / 100, 0)
  return { subtotal, taxableBase }
}
