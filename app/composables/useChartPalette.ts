// Validated categorical + status colors (light/dark), per the project's data-viz reference palette.
const CATEGORICAL_LIGHT = ['#2a78d6', '#eb6834', '#1baf7a', '#eda100', '#e87ba4', '#008300', '#4a3aa7', '#e34948']
const CATEGORICAL_DARK = ['#3987e5', '#d95926', '#199e70', '#c98500', '#d55181', '#008300', '#9085e9', '#e66767']

const STATUS_LIGHT = { good: '#0ca30c', warning: '#fab219', serious: '#ec835a', critical: '#d03b3b', muted: '#898781' }
const STATUS_DARK = { good: '#0ca30c', warning: '#fab219', serious: '#ec835a', critical: '#d03b3b', muted: '#898781' }

export function useChartPalette() {
  const colorMode = useColorMode()
  const isDark = computed(() => colorMode.value === 'dark')

  const categorical = computed(() => isDark.value ? CATEGORICAL_DARK : CATEGORICAL_LIGHT)
  const status = computed(() => isDark.value ? STATUS_DARK : STATUS_LIGHT)

  return { categorical, status, isDark }
}
