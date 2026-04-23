<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <UserDashboardStats :stats="stats" :balance="effectiveBalance" :is-simple="authStore.isSimpleMode" />
        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" @refresh="refreshAll" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div v-if="!adminViewUserID" class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'; import { useRoute } from 'vue-router'; import { useAuthStore } from '@/stores/auth'; import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import AppLayout from '@/components/layout/AppLayout.vue'; import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'; import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'; import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'
import { adminAPI } from '@/api/admin'

const authStore = useAuthStore(); const user = computed(() => authStore.user)
const route = useRoute()
const stats = ref<UserStatsType | null>(null); const loading = ref(false); const loadingUsage = ref(false); const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([]); const modelStats = ref<ModelStat[]>([]); const recentUsage = ref<UsageLog[]>([])
const adminViewUser = ref<{ id: number; balance: number } | null>(null)

const formatLD = (d: Date) => d.toISOString().split('T')[0]
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000))); const endDate = ref(formatLD(new Date())); const granularity = ref('day')

const adminViewUserID = computed(() => {
  const raw = route.query.admin_user_id
  if (!authStore.isAdmin || raw == null) return null
  const value = Array.isArray(raw) ? raw[0] : raw
  const parsed = Number.parseInt(String(value), 10)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
})
const effectiveBalance = computed(() => adminViewUser.value?.balance ?? user.value?.balance ?? 0)

const loadStats = async () => {
  loading.value = true
  try {
    if (adminViewUserID.value) {
      const [dashboardStats, adminUser] = await Promise.all([
        adminAPI.users.getUserDashboardStats(adminViewUserID.value),
        adminAPI.users.getById(adminViewUserID.value)
      ])
      stats.value = dashboardStats
      adminViewUser.value = { id: adminUser.id, balance: adminUser.balance }
      return
    }
    adminViewUser.value = null
    await authStore.refreshUser()
    stats.value = await usageAPI.getDashboardStats()
  } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false }
}
const loadCharts = async () => {
  loadingCharts.value = true
  try {
    if (adminViewUserID.value) {
      const res = await Promise.all([
        adminAPI.dashboard.getUsageTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any, user_id: adminViewUserID.value }),
        adminAPI.dashboard.getModelStats({ start_date: startDate.value, end_date: endDate.value, user_id: adminViewUserID.value })
      ])
      trendData.value = res[0].trend || []
      modelStats.value = res[1].models || []
      return
    }
    const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || []
  } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false }
}
const loadRecent = async () => {
  loadingUsage.value = true
  try {
    if (adminViewUserID.value) {
      const res = await adminAPI.usage.list({ user_id: adminViewUserID.value, start_date: startDate.value, end_date: endDate.value, page: 1, page_size: 5 })
      recentUsage.value = res.items || []
      return
    }
    const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5)
  } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false }
}
const refreshAll = () => { loadStats(); loadCharts(); loadRecent() }

onMounted(() => { refreshAll() })
</script>
