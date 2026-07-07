<script setup lang="ts">
import { LockKeyhole } from '@lucide/vue'

interface PageNode {
  path: string
  title: string
  access?: string
  has_index?: boolean
  children?: PageNode[]
}

const props = defineProps<{
  nodes: PageNode[]
  docName: string
  currentPath: string
}>()

function isActive(path: string): boolean {
  return (
    props.currentPath === path || (path === 'index' && props.currentPath === '')
  )
}

function isProtected(node: PageNode): boolean {
  return node.access === 'password'
}
</script>

<template>
  <ul class="space-y-1">
    <li v-for="node in nodes" :key="node.path">
      <NuxtLink
        v-if="!node.children?.length"
        :to="`/${docName}/${node.path}`"
        class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors"
        :class="{
          'bg-primary text-primary-foreground font-medium': isActive(node.path),
          'hover:text-primary': !isActive(node.path),
        }"
      >
        <LockKeyhole
          v-if="isProtected(node)"
          class="h-3.5 w-3.5 text-muted-foreground"
        />
        <span>{{ node.title }}</span>
      </NuxtLink>
      <div v-else class="space-y-1">
        <NuxtLink
          v-if="node.has_index"
          :to="`/${docName}/${node.path}`"
          class="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors"
          :class="{
            'bg-primary text-primary-foreground font-medium': isActive(
              node.path,
            ),
            'hover:text-primary': !isActive(node.path),
          }"
        >
          <LockKeyhole v-if="isProtected(node)" class="h-3.5 w-3.5" />
          <span>{{ node.title }}</span>
        </NuxtLink>
        <span
          v-else
          class="flex items-center gap-2 px-3 py-2 text-sm font-medium text-muted-foreground"
        >
          <LockKeyhole v-if="isProtected(node)" class="h-3.5 w-3.5" />
          <span>{{ node.title }}</span>
        </span>
        <div class="ml-3 border-l-2 border-border pl-3">
          <SidebarTree
            :nodes="node.children"
            :doc-name="docName"
            :current-path="currentPath"
          />
        </div>
      </div>
    </li>
  </ul>
</template>
