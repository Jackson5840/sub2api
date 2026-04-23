<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="rounded-2xl bg-white p-6 shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 class="text-xl font-black text-gray-900 dark:text-white">
              {{ t('admin.ops.requestTranscripts.title') }}
            </h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
              {{ t('admin.ops.requestTranscripts.description') }}
            </p>
          </div>

          <div class="flex items-center gap-2">
            <button type="button" class="btn btn-secondary" @click="resetFilters">
              {{ t('common.reset') }}
            </button>
            <button type="button" class="btn btn-secondary" :disabled="loading" @click="fetchLogs">
              {{ t('common.refresh') }}
            </button>
            <button type="button" class="btn btn-primary" :disabled="loading" @click="searchLogs">
              {{ t('common.search') }}
            </button>
          </div>
        </div>

        <div class="mt-6 grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
          <div>
            <label class="input-label">{{ t('admin.ops.requestTranscripts.filters.timeRange') }}</label>
            <Select v-model="filters.time_range" :options="timeRangeOptions" />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.requestTranscripts.filters.requestId') }}</label>
            <input
              v-model.trim="filters.request_id"
              type="text"
              class="input"
              :placeholder="t('admin.ops.requestTranscripts.filters.requestIdPlaceholder')"
              @keydown.enter.prevent="searchLogs"
            />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.requestTranscripts.filters.clientRequestId') }}</label>
            <input
              v-model.trim="filters.client_request_id"
              type="text"
              class="input"
              :placeholder="t('admin.ops.requestTranscripts.filters.clientRequestIdPlaceholder')"
              @keydown.enter.prevent="searchLogs"
            />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.requestTranscripts.filters.model') }}</label>
            <input
              v-model.trim="filters.model"
              type="text"
              class="input"
              :placeholder="t('admin.ops.requestTranscripts.filters.modelPlaceholder')"
              @keydown.enter.prevent="searchLogs"
            />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.requestTranscripts.filters.userId') }}</label>
            <Select
              v-model="filters.user_id"
              :options="userOptions"
              searchable
              :placeholder="t('admin.ops.requestTranscripts.filters.userIdPlaceholder')"
              :search-placeholder="t('admin.ops.requestTranscripts.filters.userSearchPlaceholder')"
              :empty-text="t('common.noOptionsFound')"
            />
          </div>

          <div>
            <label class="input-label">{{ t('admin.ops.requestTranscripts.filters.accountId') }}</label>
            <input
              v-model.trim="filters.account_id"
              type="text"
              inputmode="numeric"
              class="input"
              :placeholder="t('admin.ops.requestTranscripts.filters.accountIdPlaceholder')"
              @keydown.enter.prevent="searchLogs"
            />
          </div>

          <div class="md:col-span-2 xl:col-span-2">
            <label class="input-label">{{ t('admin.ops.requestTranscripts.filters.query') }}</label>
            <input
              v-model.trim="filters.q"
              type="text"
              class="input"
              :placeholder="t('admin.ops.requestTranscripts.filters.queryPlaceholder')"
              @keydown.enter.prevent="searchLogs"
            />
          </div>
        </div>
      </div>

      <div class="overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="text-sm font-bold text-gray-900 dark:text-white">
              {{ t('admin.ops.requestTranscripts.listTitle') }}
            </div>
            <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.ops.requestTranscripts.total', { total }) }}
            </div>
          </div>

          <div v-if="loading" class="flex items-center justify-center px-6 py-16 text-sm text-gray-500 dark:text-gray-400">
            {{ t('common.loading') }}
          </div>

          <div v-else-if="logs.length === 0" class="px-6 py-16 text-center">
            <div class="text-sm font-medium text-gray-600 dark:text-gray-300">
              {{ t('admin.ops.requestTranscripts.empty') }}
            </div>
            <div class="mt-1 text-xs text-gray-400">
              {{ t('admin.ops.requestTranscripts.emptyHint') }}
            </div>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
              <thead class="bg-gray-50 dark:bg-dark-900">
                <tr>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                    {{ t('admin.ops.requestTranscripts.table.user') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                    {{ t('admin.ops.requestTranscripts.table.time') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                    {{ t('admin.ops.requestTranscripts.table.model') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                    {{ t('admin.ops.requestTranscripts.table.requestText') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                    {{ t('admin.ops.requestTranscripts.table.responseText') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                    {{ t('admin.ops.requestTranscripts.table.usageCost') }}
                  </th>
                  <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                    {{ t('admin.ops.requestTranscripts.table.requestId') }}
                  </th>
                </tr>
              </thead>

              <tbody class="divide-y divide-gray-200 dark:divide-dark-700">
                <tr
                  v-for="log in logs"
                  :key="log.id"
                  class="cursor-pointer bg-white transition-colors hover:bg-gray-50 dark:bg-dark-800 dark:hover:bg-dark-700/60"
                  @click="openDetail(log)"
                >
                  <td class="px-4 py-3 text-xs text-gray-700 dark:text-gray-200">
                    <div class="max-w-[180px] truncate font-medium" :title="userLabel(log)">
                      {{ userLabel(log) }}
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-xs text-gray-600 dark:text-gray-300">
                    {{ formatDateTime(log.created_at) }}
                  </td>
                  <td class="px-4 py-3 text-xs text-gray-700 dark:text-gray-200">
                    <div class="font-medium">{{ log.model || '-' }}</div>
                    <div class="mt-1 text-[11px] text-gray-400">{{ log.platform || '-' }}</div>
                  </td>
                  <td class="max-w-[280px] px-4 py-3 text-xs text-gray-600 dark:text-gray-300">
                    <div class="line-clamp-3 whitespace-pre-wrap break-words">
                      {{ previewText(requestText(log)) || '-' }}
                    </div>
                  </td>
                  <td class="max-w-[280px] px-4 py-3 text-xs text-gray-600 dark:text-gray-300">
                    <div class="line-clamp-3 whitespace-pre-wrap break-words">
                      {{ previewText(responseText(log)) || '-' }}
                    </div>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 text-xs text-gray-600 dark:text-gray-300">
                    <div class="font-medium text-gray-800 dark:text-gray-100">
                      {{ formatTokenCount(totalTokens(log)) }}
                    </div>
                    <div class="mt-1 text-[11px] text-gray-400">
                      ${{ formatCost(totalCost(log)) }}
                    </div>
                  </td>
                  <td class="px-4 py-3 text-xs text-gray-600 dark:text-gray-300">
                    <div class="max-w-[180px] truncate font-mono" :title="log.request_id || ''">
                      {{ log.request_id || '-' }}
                    </div>
                    <div v-if="log.client_request_id" class="mt-1 max-w-[180px] truncate font-mono text-[11px] text-gray-400" :title="log.client_request_id">
                      {{ log.client_request_id }}
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="logs.length > 0" class="border-t border-gray-100 px-4 py-3 dark:border-dark-700">
            <Pagination
              :page="page"
              :page-size="pageSize"
              :total="total"
              @update:page="handlePageChange"
              @update:pageSize="handlePageSizeChange"
            />
          </div>
      </div>

      <BaseDialog
        :show="detailOpen && !!selectedLog"
        :title="t('admin.ops.requestTranscripts.detailTitleShort')"
        width="extra-wide"
        :close-on-click-outside="true"
        @close="closeDetail"
      >
        <template v-if="selectedLog">
            <div class="flex items-start justify-between gap-3">
              <div>
                <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{ formatDateTime(selectedLog.created_at) }}
                </div>
              </div>

              <button
                v-if="selectedLog.request_id"
                type="button"
                class="btn btn-secondary btn-sm"
                @click="copyToClipboard(selectedLog.request_id, t('admin.ops.requestTranscripts.requestIdCopied'))"
              >
                {{ t('admin.ops.requestTranscripts.copyRequestId') }}
              </button>
            </div>

            <div class="mt-6 grid grid-cols-1 gap-3">
              <div>
                <div class="mb-2 flex items-center justify-between gap-2">
                  <div class="text-sm font-bold text-gray-900 dark:text-white">
                    {{ t('admin.ops.requestTranscripts.detail.requestText') }}
                  </div>
                  <button
                    type="button"
                    class="btn btn-secondary btn-sm"
                    :disabled="!requestText(selectedLog)"
                    @click="copyToClipboard(requestText(selectedLog), t('admin.ops.requestTranscripts.requestTextCopied'))"
                  >
                    {{ t('admin.ops.requestTranscripts.copyRequestText') }}
                  </button>
                </div>
                <div v-if="extraBool(selectedLog, 'request_text_truncated')" class="mb-2 text-[11px] text-yellow-600 dark:text-yellow-400">
                  {{ t('admin.ops.requestTranscripts.detail.requestTextTruncated') }}
                </div>
                <pre class="max-h-[260px] overflow-auto rounded-xl border border-gray-200 bg-gray-50 p-4 text-xs text-gray-800 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-100"><code>{{ requestText(selectedLog) || '—' }}</code></pre>
              </div>

              <div>
                <div class="mb-2 flex items-center justify-between gap-2">
                  <div class="text-sm font-bold text-gray-900 dark:text-white">
                    {{ t('admin.ops.requestTranscripts.detail.responseText') }}
                  </div>
                  <button
                    type="button"
                    class="btn btn-secondary btn-sm"
                    :disabled="!responseText(selectedLog)"
                    @click="copyToClipboard(responseText(selectedLog), t('admin.ops.requestTranscripts.responseTextCopied'))"
                  >
                    {{ t('admin.ops.requestTranscripts.copyResponseText') }}
                  </button>
                </div>
                <div v-if="extraBool(selectedLog, 'response_text_truncated')" class="mb-2 text-[11px] text-yellow-600 dark:text-yellow-400">
                  {{ t('admin.ops.requestTranscripts.detail.responseTextTruncated') }}
                </div>
                <pre class="max-h-[320px] overflow-auto rounded-xl border border-gray-200 bg-gray-50 p-4 text-xs text-gray-800 dark:border-dark-700 dark:bg-dark-900 dark:text-gray-100"><code>{{ responseText(selectedLog) || '—' }}</code></pre>
              </div>

              <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.requestId') }}</div>
                <div class="mt-1 break-all font-mono text-sm text-gray-900 dark:text-white">{{ selectedLog.request_id || '—' }}</div>
              </div>

              <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.clientRequestId') }}</div>
                <div class="mt-1 break-all font-mono text-sm text-gray-900 dark:text-white">{{ selectedLog.client_request_id || '—' }}</div>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                  <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.userId') }}</div>
                  <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedLog.user_id ?? extraNumber(selectedLog, 'user_id') ?? '—' }}</div>
                </div>
                <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                  <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.accountId') }}</div>
                  <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedLog.account_id ?? extraNumber(selectedLog, 'account_id') ?? '—' }}</div>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                  <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.platform') }}</div>
                  <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedLog.platform || '—' }}</div>
                </div>
                <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                  <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.model') }}</div>
                  <div class="mt-1 break-all text-sm text-gray-900 dark:text-white">{{ selectedLog.model || '—' }}</div>
                </div>
              </div>

              <div class="grid grid-cols-2 gap-3">
                <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                  <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.apiKeyId') }}</div>
                  <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ extraNumber(selectedLog, 'api_key_id') ?? '—' }}</div>
                </div>
                <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                  <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.groupId') }}</div>
                  <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ extraNumber(selectedLog, 'group_id') ?? '—' }}</div>
                </div>
              </div>

              <template v-if="hasUsageMetrics(selectedLog)">
                <div class="grid grid-cols-2 gap-3">
                  <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                    <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.totalTokens') }}</div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatTokenCount(totalTokens(selectedLog)) }}</div>
                  </div>
                  <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                    <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.totalCost') }}</div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-white">${{ formatCost(totalCost(selectedLog)) }}</div>
                  </div>
                </div>

                <div class="grid grid-cols-2 gap-3">
                  <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                    <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.inputTokens') }}</div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatTokenCount(extraNumber(selectedLog, 'input_tokens')) }}</div>
                  </div>
                  <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                    <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.outputTokens') }}</div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatTokenCount(extraNumber(selectedLog, 'output_tokens')) }}</div>
                  </div>
                </div>

                <div class="grid grid-cols-2 gap-3">
                  <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                    <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.cacheCreationTokens') }}</div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatTokenCount(extraNumber(selectedLog, 'cache_creation_tokens')) }}</div>
                  </div>
                  <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                    <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.cacheReadTokens') }}</div>
                    <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatTokenCount(extraNumber(selectedLog, 'cache_read_tokens')) }}</div>
                  </div>
                </div>
              </template>

              <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-900">
                <div class="text-[11px] font-bold uppercase tracking-wider text-gray-400">{{ t('admin.ops.requestTranscripts.detail.endpoints') }}</div>
                <div class="mt-2 space-y-1 text-xs text-gray-700 dark:text-gray-200">
                  <div><span class="font-bold text-gray-400">{{ t('admin.ops.requestTranscripts.detail.requestPath') }}:</span> {{ extraString(selectedLog, 'request_path') || '—' }}</div>
                  <div><span class="font-bold text-gray-400">{{ t('admin.ops.requestTranscripts.detail.inboundEndpoint') }}:</span> {{ extraString(selectedLog, 'inbound_endpoint') || '—' }}</div>
                  <div><span class="font-bold text-gray-400">{{ t('admin.ops.requestTranscripts.detail.upstreamEndpoint') }}:</span> {{ extraString(selectedLog, 'upstream_endpoint') || '—' }}</div>
                  <div><span class="font-bold text-gray-400">{{ t('admin.ops.requestTranscripts.detail.upstreamModel') }}:</span> {{ extraString(selectedLog, 'upstream_model') || '—' }}</div>
                </div>
              </div>
            </div>

        </template>
      </BaseDialog>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import { opsAPI, type OpsSystemLog, type OpsSystemLogQuery } from '@/api/admin/ops'
import { adminAPI } from '@/api/admin'
import { useClipboard } from '@/composables/useClipboard'
import { useAppStore } from '@/stores'
import { formatDateTime } from './ops/utils/opsFormatters'

const REQUEST_TRANSCRIPT_COMPONENT = 'audit.request_transcript'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const logs = ref<OpsSystemLog[]>([])
const selectedLog = ref<OpsSystemLog | null>(null)
const detailOpen = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

type RequestTranscriptFilters = {
  time_range: NonNullable<OpsSystemLogQuery['time_range']>
  request_id: string
  client_request_id: string
  user_id: number | null
  account_id: string
  model: string
  q: string
}

const filters = ref<RequestTranscriptFilters>({
  time_range: '24h',
  request_id: '',
  client_request_id: '',
  user_id: null,
  account_id: '',
  model: '',
  q: ''
})

const userOptions = ref<Array<{ value: number | null; label: string }>>([
  { value: null, label: t('common.all') }
])
const userInfoMap = ref<Map<number, { username: string; email: string }>>(new Map())

const timeRangeOptions = [
  { value: '1h', label: '1h' },
  { value: '24h', label: '24h' },
  { value: '7d', label: '7d' },
  { value: '30d', label: '30d' }
]

async function loadUserOptions() {
  const pageSize = 200
  let page = 1
  let totalUsers = 0
  const items: Array<{ id: number; username: string; email: string }> = []

  try {
    do {
      const res = await adminAPI.users.list(page, pageSize)
      totalUsers = res.total || 0
      for (const user of res.items || []) {
        if (typeof user.id === 'number') {
          items.push({
            id: user.id,
            username: typeof user.username === 'string' ? user.username.trim() : '',
            email: typeof user.email === 'string' ? user.email.trim() : ''
          })
        }
      }
      page += 1
      if (page > 50) break
    } while (items.length < totalUsers)

    items.sort((a, b) => {
      const left = a.username || a.email || String(a.id)
      const right = b.username || b.email || String(b.id)
      return left.localeCompare(right)
    })
    userOptions.value = [
      { value: null, label: t('common.all') },
      ...items.map((user) => ({
        value: user.id,
        label: user.username
          ? `${user.username} (${user.email || '#' + user.id})`
          : `${user.email || '#' + user.id} (#${user.id})`
      }))
    ]
    userInfoMap.value = new Map(items.map((user) => [user.id, {
      username: user.username,
      email: user.email
    }]))
  } catch (error) {
    console.error('[RequestTranscriptsView] Failed to load users', error)
    userOptions.value = [{ value: null, label: t('common.all') }]
    userInfoMap.value = new Map()
  }
}

function previewText(text: string): string {
  const value = (text || '').trim()
  if (!value) return ''
  return value.length > 120 ? value.slice(0, 120) + '...' : value
}

function extraString(log: OpsSystemLog | null, key: string): string {
  const value = log?.extra?.[key]
  if (typeof value === 'string') return value
  if (value == null) return ''
  return String(value)
}

function extraBool(log: OpsSystemLog | null, key: string): boolean {
  return log?.extra?.[key] === true
}

function extraNumber(log: OpsSystemLog | null, key: string): number | null {
  const value = log?.extra?.[key]
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : null
  }
  return null
}

function totalTokens(log: OpsSystemLog | null): number | null {
  const values = [
    extraNumber(log, 'input_tokens'),
    extraNumber(log, 'output_tokens'),
    extraNumber(log, 'cache_creation_tokens'),
    extraNumber(log, 'cache_read_tokens')
  ]
  if (values.every((value) => value == null)) {
    return null
  }
  return values.reduce<number>((sum, value) => sum + (value ?? 0), 0)
}

function totalCost(log: OpsSystemLog | null): number | null {
  const actual = extraNumber(log, 'actual_cost')
  if (actual != null) return actual
  return extraNumber(log, 'total_cost')
}

function hasUsageMetrics(log: OpsSystemLog | null): boolean {
  return totalTokens(log) != null || totalCost(log) != null
}

function formatTokenCount(value: number | null): string {
  if (value == null || !Number.isFinite(value)) return '—'
  return Math.round(value).toLocaleString()
}

function formatCost(value: number | null): string {
  if (value == null || !Number.isFinite(value)) return '—'
  return value.toFixed(6)
}

function requestText(log: OpsSystemLog | null): string {
  return extraString(log, 'request_text')
}

function responseText(log: OpsSystemLog | null): string {
  return extraString(log, 'response_text')
}

function userLabel(log: OpsSystemLog | null): string {
  const userID = log?.user_id ?? extraNumber(log, 'user_id')
  if (typeof userID === 'number' && userInfoMap.value.has(userID)) {
    const info = userInfoMap.value.get(userID)
    if (info?.username) return info.username
    if (info?.email) return info.email
  }
  if (typeof userID === 'number') {
    return `#${userID}`
  }
  return '—'
}

function openDetail(log: OpsSystemLog) {
  selectedLog.value = log
  detailOpen.value = true
}

function closeDetail() {
  detailOpen.value = false
}

function buildQuery(): OpsSystemLogQuery {
  const query: OpsSystemLogQuery = {
    page: page.value,
    page_size: pageSize.value,
    component: REQUEST_TRANSCRIPT_COMPONENT,
    time_range: filters.value.time_range
  }

  if (filters.value.request_id) query.request_id = filters.value.request_id
  if (filters.value.client_request_id) query.client_request_id = filters.value.client_request_id
  if (filters.value.model) query.model = filters.value.model
  if (filters.value.q) query.q = filters.value.q

  if (typeof filters.value.user_id === 'number' && filters.value.user_id > 0) {
    query.user_id = filters.value.user_id
  }
  if (filters.value.account_id) {
    const accountId = Number.parseInt(filters.value.account_id, 10)
    if (Number.isFinite(accountId) && accountId > 0) query.account_id = accountId
  }

  return query
}

async function fetchLogs() {
  loading.value = true
  try {
    const res = await opsAPI.listSystemLogs(buildQuery())
    logs.value = res.items || []
    total.value = res.total || 0

    if (!selectedLog.value || !logs.value.some((item) => item.id === selectedLog.value?.id)) {
      selectedLog.value = logs.value[0] || null
      if (detailOpen.value && !selectedLog.value) {
        detailOpen.value = false
      }
    }
  } catch (error: any) {
    console.error('[RequestTranscriptsView] Failed to load transcripts', error)
    appStore.showError(error?.message || t('admin.ops.requestTranscripts.failedToLoad'))
    logs.value = []
    total.value = 0
    selectedLog.value = null
  } finally {
    loading.value = false
  }
}

function searchLogs() {
  page.value = 1
  fetchLogs()
}

function resetFilters() {
  filters.value = {
    time_range: '24h',
    request_id: '',
    client_request_id: '',
    user_id: null,
    account_id: '',
    model: '',
    q: ''
  }
  page.value = 1
  fetchLogs()
}

function handlePageChange(next: number) {
  page.value = next
  fetchLogs()
}

function handlePageSizeChange(next: number) {
  pageSize.value = next
  page.value = 1
  fetchLogs()
}

onMounted(() => {
  loadUserOptions()
  fetchLogs()
})
</script>
