<script setup lang="ts">
const { t } = useText()

const props = defineProps<{
  doc: string
  pagePath: string
  scope?: string
}>()

const emit = defineEmits<{
  success: []
  close: []
}>()

function close() {
  emit('close')
}

function onBackdropClick() {
  if (!autoVerify.value) close()
}

const password = ref('')
const submitting = ref(false)
const error = ref<string | null>(null)
const remember = ref(false)
const autoVerify = ref(false)

const config = useRuntimeConfig()

function effectiveScope(responseScope?: string): string {
  return responseScope || props.scope || props.pagePath || '_'
}

function storageKey(scope?: string): string {
  return `vordoc_pwd_${props.doc}_${effectiveScope(scope)}`
}

function savePassword(pwd: string, scope?: string) {
  try {
    localStorage.setItem(storageKey(scope), btoa(pwd))
  } catch {
    // ignore storage errors
  }
}

function clearSavedPassword(scope?: string) {
  try {
    localStorage.removeItem(storageKey(scope))
  } catch {
    // ignore storage errors
  }
}

function loadPassword(): string | null {
  try {
    const raw = localStorage.getItem(storageKey())
    return raw ? atob(raw) : null
  } catch {
    return null
  }
}

async function verify(pwd: string): Promise<string | undefined> {
  if (!pwd) return undefined

  submitting.value = true
  error.value = null

  try {
    const response = await $fetch<{
      success?: boolean
      scope?: string
    }>(`${config.public.apiBase}/v1/${props.doc}/${props.pagePath}`, {
      method: 'POST',
      credentials: 'include',
      body: { password: pwd },
    })
    emit('success')
    return response.scope
  } catch (e: unknown) {
    const code =
      e && typeof e === 'object' && 'data' in e
        ? (e as { data?: { error?: string } }).data?.error
        : undefined
    error.value = code ? t(`errors.${code}`) : t('password.failed')
    if (autoVerify.value) {
      clearSavedPassword()
    }
    return undefined
  } finally {
    submitting.value = false
  }
}

async function submit() {
  if (!password.value) return
  const responseScope = await verify(password.value)
  if (!error.value && remember.value) {
    savePassword(password.value, responseScope)
  }
}

onMounted(() => {
  const saved = loadPassword()
  if (saved) {
    autoVerify.value = true
    verify(saved).finally(() => {
      autoVerify.value = false
    })
  }
})
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
      @click.self="onBackdropClick"
    >
      <div
        v-if="autoVerify"
        class="password-card flex flex-col items-center gap-4 rounded-lg border bg-card p-8 shadow-xl"
      >
        <div
          class="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"
        />
        <p class="text-sm text-muted-foreground">
          {{ t('password.authenticating') }}
        </p>
      </div>

      <div
        v-else
        class="password-card w-full max-w-md rounded-lg border bg-card p-8 shadow-xl"
      >
        <h2 class="mb-2 text-xl font-semibold">{{ t('password.title') }}</h2>
        <p class="mb-6 text-sm text-muted-foreground">
          {{ t('password.description') }}
        </p>

        <form class="space-y-4" @submit.prevent="submit">
          <div>
            <input
              v-model="password"
              type="password"
              :placeholder="t('password.placeholder')"
              class="w-full rounded-md border border-input bg-background px-3 py-3 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            />
          </div>

          <label
            class="flex cursor-pointer items-center gap-2 text-sm text-muted-foreground"
          >
            <input
              v-model="remember"
              type="checkbox"
              class="h-4 w-4 rounded border-input accent-primary"
            />
            {{ t('password.remember') }}
          </label>

          <p v-if="error" class="text-sm text-destructive">{{ error }}</p>

          <UiButton
            type="submit"
            class="w-full"
            :disabled="submitting || !password"
          >
            {{ submitting ? t('password.verifying') : t('password.unlock') }}
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
