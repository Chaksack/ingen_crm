export function useBreadcrumbOverride() {
  return useState<string | null>('breadcrumb-last-label', () => null)
}
