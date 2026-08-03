export interface TaxRatePreview {
  id: string
  name: string
  rate: string
  compound: boolean
  active: boolean
}

export interface TaxLinePreview {
  name: string
  rate: number
  amount: number
  compound: boolean
}

function round2(n: number) {
  return Math.round(n * 100) / 100
}

export function computeTaxPreview(taxableBase: number, rates: TaxRatePreview[]) {
  const active = rates.filter(r => r.active)
  const nonCompound = active.filter(r => !r.compound)
  const compound = active.filter(r => r.compound)

  const lines: TaxLinePreview[] = []
  let leviesTotal = 0

  for (const r of nonCompound) {
    const amount = round2(taxableBase * (Number(r.rate) / 100))
    leviesTotal += amount
    lines.push({ name: r.name, rate: Number(r.rate), amount, compound: false })
  }

  const compoundBase = taxableBase + leviesTotal
  for (const r of compound) {
    const amount = round2(compoundBase * (Number(r.rate) / 100))
    lines.push({ name: r.name, rate: Number(r.rate), amount, compound: true })
  }

  const taxAmount = round2(lines.reduce((sum, l) => sum + l.amount, 0))
  return { lines, taxAmount }
}

export function useTaxPreview() {
  const { data: taxSettings } = useFetch<{ rates: TaxRatePreview[] }>('/api/tax-settings')
  const activeRates = computed(() => (taxSettings.value?.rates ?? []).filter(r => r.active))
  return { activeRates }
}
