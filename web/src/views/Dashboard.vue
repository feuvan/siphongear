<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api } from '@/api'

interface Card {
  collector_id: number
  collector_name: string
  site_id: number
  site_name: string
  indicator_id: number
  key: string
  name: string
  type: string
  unit: string
  display: string
  value_num: number | null
  value_str: string | null
  value_json: string | null
  ts: string | null
  last_status: string
}

const router = useRouter()
const cards = ref<Card[]>([])
const loading = ref(false)
const refreshing = ref<Record<number, boolean>>({})

async function reload() {
  loading.value = true
  try {
    cards.value = await api.dashboard()
  } finally {
    loading.value = false
  }
}

function formatValue(c: Card): string {
  if (c.value_num !== null && c.value_num !== undefined) {
    const n = c.value_num
    if (Number.isInteger(n)) return String(n)
    return n.toFixed(Math.min(4, Math.max(0, 6 - String(Math.trunc(n)).length)))
  }
  if (c.value_str) return c.value_str
  if (c.value_json) return c.value_json
  return '—'
}

function relativeTime(iso: string | null): string {
  if (!iso) return 'no data'
  const t = new Date(iso).getTime()
  const diff = (Date.now() - t) / 1000
  if (diff < 60) return `${Math.floor(diff)}s ago`
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

function statusType(status: string): 'success' | 'danger' | 'info' | 'warning' {
  if (status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'running') return 'warning'
  return 'info'
}

// hash-derived accent color for cards (consistent per site)
function siteAccent(siteId: number): string {
  const palette = [
    '#5b8def', '#36b37e', '#f5a524', '#e25c84',
    '#9b59b6', '#26a69a', '#ef6c00', '#3f51b5'
  ]
  return palette[siteId % palette.length] || '#909399'
}

async function refresh(card: Card) {
  refreshing.value[card.collector_id] = true
  try {
    await api.collectors.run(card.collector_id, {})
    await reload()
    ElMessage.success(`refreshed ${card.collector_name}`)
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'run failed')
  } finally {
    refreshing.value[card.collector_id] = false
  }
}

function gotoRuns(card: Card) {
  router.push({ name: 'collector-runs', params: { id: card.collector_id } })
}

function gotoEdit(card: Card) {
  router.push({ name: 'collector-edit', params: { id: card.collector_id } })
}

const groupedBySite = computed(() => {
  const groups: Record<string, { siteId: number; siteName: string; cards: Card[] }> = {}
  for (const c of cards.value) {
    const key = c.site_name || `Site #${c.site_id || 0}`
    if (!groups[key]) groups[key] = { siteId: c.site_id, siteName: key, cards: [] }
    groups[key].cards.push(c)
  }
  return Object.values(groups).sort((a, b) => a.siteName.localeCompare(b.siteName))
})

onMounted(reload)
</script>

<template>
  <div>
    <div class="bar">
      <div>
        <h2>Dashboard</h2>
        <div class="subtitle">{{ cards.length }} indicator(s) across {{ groupedBySite.length }} site(s)</div>
      </div>
      <el-button type="primary" :loading="loading" @click="reload">Refresh All</el-button>
    </div>

    <el-empty v-if="!cards.length && !loading" description="No indicators yet. Add some collectors first." />

    <div v-for="group in groupedBySite" :key="group.siteName" class="site-group">
      <div class="site-header" :style="{ borderColor: siteAccent(group.siteId) }">
        <span class="site-dot" :style="{ background: siteAccent(group.siteId) }"></span>
        <span class="site-name">{{ group.siteName }}</span>
        <span class="site-count">{{ group.cards.length }}</span>
      </div>

      <el-row :gutter="16">
        <el-col v-for="c in group.cards" :key="c.indicator_id" :xs="24" :sm="12" :md="8" :lg="6">
          <div class="metric" :style="{ '--accent': siteAccent(group.siteId) } as any">
            <div class="metric-head">
              <span class="metric-name">{{ c.name }}</span>
              <el-tag size="small" :type="statusType(c.last_status)" effect="plain">
                {{ c.last_status || '—' }}
              </el-tag>
            </div>

            <div class="metric-value">
              <span class="num">{{ formatValue(c) }}</span>
              <span v-if="c.unit" class="unit">{{ c.unit }}</span>
            </div>

            <div class="meta">
              <span class="key">{{ c.key }}</span>
              <span class="dot">·</span>
              <span class="collector">{{ c.collector_name }}</span>
            </div>

            <div class="footer">
              <span class="ts">{{ relativeTime(c.ts) }}</span>
              <div class="actions">
                <el-button
                  link size="small" :loading="refreshing[c.collector_id]"
                  @click="refresh(c)"
                >Run</el-button>
                <el-button link size="small" @click="gotoRuns(c)">History</el-button>
                <el-button link size="small" @click="gotoEdit(c)">Edit</el-button>
              </div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<style scoped>
.bar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}
.bar h2 { margin: 0 0 4px; font-size: 24px; }
.subtitle { color: #909399; font-size: 13px; }

.site-group { margin-bottom: 28px; }

.site-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0 10px;
  margin-bottom: 12px;
  border-bottom: 1px solid #ebeef5;
  border-bottom-width: 2px;
}
.site-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.site-name {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}
.site-count {
  font-size: 12px;
  color: #909399;
  background: #f4f4f5;
  padding: 1px 8px;
  border-radius: 10px;
}

.metric {
  position: relative;
  background: #fff;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  transition: box-shadow .2s, transform .2s;
  overflow: hidden;
}
.metric::before {
  content: '';
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 3px;
  background: var(--accent, #5b8def);
}
.metric:hover {
  box-shadow: 0 6px 18px rgba(0, 0, 0, .08);
  transform: translateY(-1px);
}

.metric-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.metric-name {
  font-size: 13px;
  color: #606266;
  font-weight: 500;
}

.metric-value {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 10px;
  min-height: 38px;
}
.metric-value .num {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
  line-height: 1;
  letter-spacing: -0.5px;
}
.metric-value .unit {
  font-size: 13px;
  color: #909399;
}

.meta {
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 12px;
}
.meta .key {
  font-family: ui-monospace, "SF Mono", Menlo, monospace;
  background: #f4f4f5;
  padding: 1px 6px;
  border-radius: 3px;
  color: #606266;
}
.meta .dot { color: #c0c4cc; }
.meta .collector { color: #606266; }

.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 10px;
  border-top: 1px dashed #ebeef5;
}
.ts {
  font-size: 11px;
  color: #c0c4cc;
}
.actions :deep(.el-button + .el-button) {
  margin-left: 4px;
}
</style>
