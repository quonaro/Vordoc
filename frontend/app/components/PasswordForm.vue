<script setup lang="ts">
const props = defineProps<{
  doc: string
  pagePath: string
}>()

const emit = defineEmits<{
  success: []
  close: []
}>()

function close() {
  emit('close')
}

const password = ref('')
const submitting = ref(false)
const error = ref<string | null>(null)

const config = useRuntimeConfig()

async function submit() {
  if (!password.value) return

  submitting.value = true
  error.value = null

  try {
    await $fetch(`${config.public.apiBase}/v1/${props.doc}/${props.pagePath}`, {
      method: 'POST',
      credentials: 'include',
      body: { password: password.value },
    })
    emit('success')
  } catch (e: unknown) {
    const message =
      e && typeof e === 'object' && 'data' in e
        ? (e as { data?: { error?: string } }).data?.error
        : undefined
    error.value = message || 'Failed to verify password'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Transition
    appear
    enter-active-class="transition duration-300 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition duration-200 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 p-4 backdrop-blur-sm"
      @click.self="close"
    >
      <div
        class="password-card w-full max-w-md rounded-lg border bg-card p-8 shadow-xl"
      >
        <h2 class="mb-2 text-xl font-semibold">Password required</h2>
        <p class="mb-6 text-sm text-muted-foreground">
          This page is protected. Enter the password to continue.
        </p>

        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <input
              v-model="password"
              type="password"
              placeholder="Password"
              class="w-full rounded-md border border-input bg-background px-3 py-3 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>

          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

          <UiButton
            type="submit"
            class="w-full"
            :disabled="submitting || !password"
          >
            {{ submitting ? 'Verifying...' : 'Unlock' }}
          </UiButton>
        </form>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.password-card {
  animation: card-in 0.35s ease-out;
}

@keyframes card-in {
  from {
    opacity: 0;
    transform: scale(0.96) translateY(12px);
  }

  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
</style>
