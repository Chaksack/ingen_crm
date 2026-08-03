<script lang="ts" setup>
export interface LinkProp {
  title: string
  label?: string
  icon: string
  variant: 'default' | 'ghost'
}

interface NavProps {
  isCollapsed: boolean
  links: LinkProp[]
}

defineProps<NavProps>()
</script>

<template>
  <div
    :data-collapsed="isCollapsed"
    class="group flex flex-col gap-0.5 py-1 data-[collapsed=true]:py-2"
  >
    <nav class="grid gap-0.5 px-2 group-[[data-collapsed=true]]:justify-center group-[[data-collapsed=true]]:px-2">
      <template v-for="(link, index) of links">
        <Tooltip v-if="isCollapsed" :key="`1-${index}`" :delay-duration="0">
          <TooltipTrigger as-child>
            <a
              href="#"
              :class="cn(
                'flex h-9 w-9 items-center justify-center rounded-md transition-colors',
                link.variant === 'default'
                  ? 'bg-[#EFF6FC] text-[#0078D4] dark:bg-[#0078D4]/15 dark:text-[#4CC2FF]'
                  : 'text-foreground/70 hover:bg-accent hover:text-foreground',
              )"
            >
              <Icon :name="link.icon" class="size-4" />
              <span class="sr-only">{{ link.title }}</span>
            </a>
          </TooltipTrigger>
          <TooltipContent side="right" class="flex items-center gap-4">
            {{ link.title }}
            <span v-if="link.label" class="ml-auto text-muted-foreground">
              {{ link.label }}
            </span>
          </TooltipContent>
        </Tooltip>

        <a
          v-else
          :key="`2-${index}`"
          href="#"
          :class="cn(
            'relative flex items-center gap-2.5 rounded-md px-2.5 py-[7px] text-[13px] font-medium transition-colors',
            link.variant === 'default'
              ? 'bg-[#EFF6FC] text-[#0078D4] dark:bg-[#0078D4]/15 dark:text-[#4CC2FF]'
              : 'text-foreground/80 hover:bg-accent',
          )"
        >
          <span
            v-if="link.variant === 'default'"
            class="absolute left-0 top-1/2 h-4 w-[3px] -translate-y-1/2 rounded-r-full bg-[#0078D4]"
          />
          <Icon :name="link.icon" class="size-4 shrink-0" />
          <span class="truncate">{{ link.title }}</span>
          <span
            v-if="link.label"
            :class="cn(
              'ml-auto shrink-0 rounded-full px-1.5 py-px text-[11px] tabular-nums',
              link.variant === 'default'
                ? 'font-semibold text-[#0078D4] dark:text-[#4CC2FF]'
                : 'text-muted-foreground',
            )"
          >
            {{ link.label }}
          </span>
        </a>
      </template>
    </nav>
  </div>
</template>
