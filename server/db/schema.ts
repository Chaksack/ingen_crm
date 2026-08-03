import { relations } from 'drizzle-orm'
import { boolean, decimal, integer, pgTable, text, timestamp, unique } from 'drizzle-orm/pg-core'
import { nanoid } from 'nanoid/non-secure'

function id() {
  return text('id').primaryKey().$defaultFn(() => nanoid(10))
}

function timestamps() {
  return {
    createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
    updatedAt: timestamp('updated_at', { withTimezone: true }).notNull().defaultNow().$onUpdate(() => new Date()),
  }
}

// ---------- Staff ----------

export const staff = pgTable('staff', {
  id: id(),
  firstName: text('first_name').notNull(),
  lastName: text('last_name').notNull(),
  email: text('email').notNull().unique(),
  phoneNumber: text('phone_number'),
  role: text('role', { enum: ['Staff', 'Manager', 'Admin'] }).notNull().default('Staff'),
  department: text('department'),
  status: text('status', { enum: ['active', 'inactive'] }).notNull().default('active'),
  avatar: text('avatar'),
  ...timestamps(),
})

// ---------- Auth ----------

export const users = pgTable('users', {
  id: id(),
  email: text('email').notNull().unique(),
  passwordHash: text('password_hash'),
  name: text('name').notNull(),
  avatar: text('avatar'),
  role: text('role', { enum: ['admin', 'manager', 'staff'] }).notNull().default('staff'),
  status: text('status', { enum: ['active', 'inactive', 'pending'] }).notNull().default('active'),
  staffId: text('staff_id').references(() => staff.id, { onDelete: 'set null' }),
  ...timestamps(),
})

export const authCodes = pgTable('auth_codes', {
  id: id(),
  userId: text('user_id').notNull().references(() => users.id, { onDelete: 'cascade' }),
  code: text('code').notNull(),
  purpose: text('purpose', { enum: ['login_otp', 'password_reset', 'invite'] }).notNull(),
  expiresAt: timestamp('expires_at', { withTimezone: true }).notNull(),
  consumedAt: timestamp('consumed_at', { withTimezone: true }),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
})

// ---------- Clients ----------

export const clients = pgTable('clients', {
  id: id(),
  name: text('name').notNull(),
  contactPerson: text('contact_person'),
  email: text('email').notNull().unique(),
  phone: text('phone'),
  address: text('address'),
  notes: text('notes'),
  status: text('status', { enum: ['active', 'inactive'] }).notNull().default('active'),
  creditScore: integer('credit_score'),
  kyc: boolean('kyc').notNull().default(false),
  dob: text('dob'),
  nationality: text('nationality'),
  idType: text('id_type'),
  idNumber: text('id_number'),
  monthlyIncome: decimal('monthly_income', { precision: 14, scale: 2 }),
  avatar: text('avatar'),
  ...timestamps(),
})

export const clientBankAccounts = pgTable('client_bank_accounts', {
  id: id(),
  clientId: text('client_id').notNull().references(() => clients.id, { onDelete: 'cascade' }),
  bankName: text('bank_name').notNull(),
  accountNumber: text('account_number').notNull(),
  accountHolder: text('account_holder').notNull(),
  ...timestamps(),
})

export const clientMomoAccounts = pgTable('client_momo_accounts', {
  id: id(),
  clientId: text('client_id').notNull().references(() => clients.id, { onDelete: 'cascade' }),
  networkName: text('network_name').notNull(),
  momoNumber: text('momo_number').notNull(),
  accountHolder: text('account_holder').notNull(),
  ...timestamps(),
})

// ---------- Finance: Invoices ----------

export const invoiceStatusValues = ['draft', 'sent', 'partially_paid', 'paid', 'overdue', 'void'] as const

export const invoices = pgTable('invoices', {
  id: id(),
  invoiceNumber: text('invoice_number').notNull().unique(),
  receiptNumber: text('receipt_number').unique(),
  clientId: text('client_id').notNull().references(() => clients.id, { onDelete: 'restrict' }),
  projectId: text('project_id'),
  issueDate: text('issue_date').notNull(),
  dueDate: text('due_date').notNull(),
  currency: text('currency').notNull().default('GHS'),
  subtotal: decimal('subtotal', { precision: 14, scale: 2 }).notNull().default('0'),
  taxExempt: boolean('tax_exempt').notNull().default(false),
  taxAmount: decimal('tax_amount', { precision: 14, scale: 2 }).notNull().default('0'),
  discount: decimal('discount', { precision: 14, scale: 2 }).notNull().default('0'),
  total: decimal('total', { precision: 14, scale: 2 }).notNull().default('0'),
  status: text('status', { enum: invoiceStatusValues }).notNull().default('draft'),
  notes: text('notes'),
  terms: text('terms'),
  createdBy: text('created_by').references(() => users.id, { onDelete: 'set null' }),
  ...timestamps(),
})

export const invoiceLineItems = pgTable('invoice_line_items', {
  id: id(),
  invoiceId: text('invoice_id').notNull().references(() => invoices.id, { onDelete: 'cascade' }),
  description: text('description').notNull(),
  quantity: decimal('quantity', { precision: 12, scale: 2 }).notNull().default('1'),
  unitPrice: decimal('unit_price', { precision: 14, scale: 2 }).notNull().default('0'),
  amount: decimal('amount', { precision: 14, scale: 2 }).notNull().default('0'),
  order: integer('order').notNull().default(0),
})

export const invoiceTaxes = pgTable('invoice_taxes', {
  id: id(),
  invoiceId: text('invoice_id').notNull().references(() => invoices.id, { onDelete: 'cascade' }),
  name: text('name').notNull(),
  rate: decimal('rate', { precision: 6, scale: 3 }).notNull(),
  amount: decimal('amount', { precision: 14, scale: 2 }).notNull(),
  compound: boolean('compound').notNull().default(false),
  order: integer('order').notNull().default(0),
})

// ---------- Finance: Quotations ----------

export const quotationStatusValues = ['draft', 'sent', 'accepted', 'rejected', 'expired'] as const

export const quotations = pgTable('quotations', {
  id: id(),
  quoteNumber: text('quote_number').notNull().unique(),
  clientId: text('client_id').notNull().references(() => clients.id, { onDelete: 'restrict' }),
  issueDate: text('issue_date').notNull(),
  expiryDate: text('expiry_date').notNull(),
  currency: text('currency').notNull().default('GHS'),
  subtotal: decimal('subtotal', { precision: 14, scale: 2 }).notNull().default('0'),
  taxExempt: boolean('tax_exempt').notNull().default(false),
  taxAmount: decimal('tax_amount', { precision: 14, scale: 2 }).notNull().default('0'),
  discount: decimal('discount', { precision: 14, scale: 2 }).notNull().default('0'),
  total: decimal('total', { precision: 14, scale: 2 }).notNull().default('0'),
  status: text('status', { enum: quotationStatusValues }).notNull().default('draft'),
  notes: text('notes'),
  terms: text('terms'),
  convertedInvoiceId: text('converted_invoice_id').references(() => invoices.id, { onDelete: 'set null' }),
  createdBy: text('created_by').references(() => users.id, { onDelete: 'set null' }),
  ...timestamps(),
})

export const quotationLineItems = pgTable('quotation_line_items', {
  id: id(),
  quotationId: text('quotation_id').notNull().references(() => quotations.id, { onDelete: 'cascade' }),
  description: text('description').notNull(),
  quantity: decimal('quantity', { precision: 12, scale: 2 }).notNull().default('1'),
  unitPrice: decimal('unit_price', { precision: 14, scale: 2 }).notNull().default('0'),
  amount: decimal('amount', { precision: 14, scale: 2 }).notNull().default('0'),
  order: integer('order').notNull().default(0),
})

export const quotationTaxes = pgTable('quotation_taxes', {
  id: id(),
  quotationId: text('quotation_id').notNull().references(() => quotations.id, { onDelete: 'cascade' }),
  name: text('name').notNull(),
  rate: decimal('rate', { precision: 6, scale: 3 }).notNull(),
  amount: decimal('amount', { precision: 14, scale: 2 }).notNull(),
  compound: boolean('compound').notNull().default(false),
  order: integer('order').notNull().default(0),
})

// ---------- Finance: Payments ----------

export const payments = pgTable('payments', {
  id: id(),
  invoiceId: text('invoice_id').notNull().references(() => invoices.id, { onDelete: 'cascade' }),
  amount: decimal('amount', { precision: 14, scale: 2 }).notNull(),
  method: text('method', { enum: ['cash', 'bank_transfer', 'mobile_money', 'card', 'cheque'] }).notNull(),
  reference: text('reference'),
  paidAt: text('paid_at').notNull(),
  notes: text('notes'),
  recordedBy: text('recorded_by').references(() => users.id, { onDelete: 'set null' }),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
})

// ---------- Finance: Expenses ----------

export const expenses = pgTable('expenses', {
  id: id(),
  category: text('category').notNull(),
  amount: decimal('amount', { precision: 14, scale: 2 }).notNull(),
  currency: text('currency').notNull().default('GHS'),
  expenseDate: text('expense_date').notNull(),
  description: text('description'),
  paymentMethod: text('payment_method', { enum: ['cash', 'bank_transfer', 'mobile_money', 'card', 'cheque'] }).notNull(),
  receiptUrl: text('receipt_url'),
  status: text('status', { enum: ['recorded', 'approved', 'rejected'] }).notNull().default('recorded'),
  recordedBy: text('recorded_by').references(() => users.id, { onDelete: 'set null' }),
  ...timestamps(),
})

// ---------- Customer Support ----------

export const supportTicketStatusValues = ['open', 'in_progress', 'resolved', 'closed'] as const
export const supportTicketPriorityValues = ['low', 'medium', 'high', 'urgent'] as const
export const supportTicketContactValues = ['chat', 'email'] as const

export const supportTickets = pgTable('support_tickets', {
  id: id(),
  ticketNumber: text('ticket_number').notNull().unique(),
  name: text('name').notNull(),
  email: text('email').notNull(),
  phone: text('phone'),
  company: text('company'),
  subject: text('subject').notNull(),
  description: text('description').notNull(),
  category: text('category'),
  priority: text('priority', { enum: supportTicketPriorityValues }).notNull().default('medium'),
  status: text('status', { enum: supportTicketStatusValues }).notNull().default('open'),
  preferredContact: text('preferred_contact', { enum: supportTicketContactValues }).notNull().default('email'),
  // Nullable + unique: lets an anonymous customer authenticate to their own ticket's chat
  // thread via a link, without requiring a full account. Older (pre-chat) tickets have none.
  accessToken: text('access_token').unique(),
  clientId: text('client_id').references(() => clients.id, { onDelete: 'set null' }),
  assignedTo: text('assigned_to').references(() => users.id, { onDelete: 'set null' }),
  ...timestamps(),
})

export const supportTicketMessages = pgTable('support_ticket_messages', {
  id: id(),
  ticketId: text('ticket_id').notNull().references(() => supportTickets.id, { onDelete: 'cascade' }),
  authorType: text('author_type', { enum: ['staff', 'client'] }).notNull(),
  authorName: text('author_name').notNull(),
  body: text('body').notNull(),
  createdAt: timestamp('created_at', { withTimezone: true }).notNull().defaultNow(),
})

// ---------- Tax Settings ----------

export const taxRates = pgTable('tax_rates', {
  id: id(),
  name: text('name').notNull(),
  rate: decimal('rate', { precision: 6, scale: 3 }).notNull(),
  compound: boolean('compound').notNull().default(false),
  order: integer('order').notNull().default(0),
  active: boolean('active').notNull().default(true),
  ...timestamps(),
})

export const companyTaxSettings = pgTable('company_tax_settings', {
  id: id(),
  companyName: text('company_name').notNull().default('Ingenicx'),
  tin: text('tin'),
  vatNumber: text('vat_number'),
  address: text('address'),
  phone: text('phone'),
  email: text('email'),
  ...timestamps(),
})

// ---------- Document numbering ----------

export const documentSequences = pgTable('document_sequences', {
  key: text('key').primaryKey(),
  value: integer('value').notNull().default(0),
}, table => [
  unique('document_sequences_key_unique').on(table.key),
])

// ---------- Relations ----------

export const invoicesRelations = relations(invoices, ({ many, one }) => ({
  lineItems: many(invoiceLineItems),
  taxes: many(invoiceTaxes),
  payments: many(payments),
  client: one(clients, { fields: [invoices.clientId], references: [clients.id] }),
}))

export const invoiceLineItemsRelations = relations(invoiceLineItems, ({ one }) => ({
  invoice: one(invoices, { fields: [invoiceLineItems.invoiceId], references: [invoices.id] }),
}))

export const invoiceTaxesRelations = relations(invoiceTaxes, ({ one }) => ({
  invoice: one(invoices, { fields: [invoiceTaxes.invoiceId], references: [invoices.id] }),
}))

export const paymentsRelations = relations(payments, ({ one }) => ({
  invoice: one(invoices, { fields: [payments.invoiceId], references: [invoices.id] }),
}))

export const quotationsRelations = relations(quotations, ({ many, one }) => ({
  lineItems: many(quotationLineItems),
  taxes: many(quotationTaxes),
  client: one(clients, { fields: [quotations.clientId], references: [clients.id] }),
}))

export const quotationLineItemsRelations = relations(quotationLineItems, ({ one }) => ({
  quotation: one(quotations, { fields: [quotationLineItems.quotationId], references: [quotations.id] }),
}))

export const quotationTaxesRelations = relations(quotationTaxes, ({ one }) => ({
  quotation: one(quotations, { fields: [quotationTaxes.quotationId], references: [quotations.id] }),
}))

export const clientsRelations = relations(clients, ({ many }) => ({
  bankAccounts: many(clientBankAccounts),
  momoAccounts: many(clientMomoAccounts),
  invoices: many(invoices),
  quotations: many(quotations),
  supportTickets: many(supportTickets),
}))

export const supportTicketsRelations = relations(supportTickets, ({ many, one }) => ({
  messages: many(supportTicketMessages),
  client: one(clients, { fields: [supportTickets.clientId], references: [clients.id] }),
  assignee: one(users, { fields: [supportTickets.assignedTo], references: [users.id] }),
}))

export const supportTicketMessagesRelations = relations(supportTicketMessages, ({ one }) => ({
  ticket: one(supportTickets, { fields: [supportTicketMessages.ticketId], references: [supportTickets.id] }),
}))
