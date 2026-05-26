<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { api } from '@/api'

const AUTO_KEY = 'dashboard.autoReload'
const INT_KEY = 'dashboard.autoReloadInterval'
const TAG_KEY = 'dashboard.tagFilter'
const INTERVAL_OPTIONS = [10, 30, 60, 120, 300]

interface Card {
  collector_id: number
  collector_name: string
  site_id: number
  site_name: string
  site_base_url: string
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
  prev_value_num: number | null
  prev_value_str: string | null
  prev_value_json: string | null
  prev_ts: string | null
  site_tags: string[] | null
  last_status: string
}

const router = useRouter()
const cards = ref<Card[]>([])
const loading = ref(false)
const refreshing = ref<Record<number, boolean>>({})
const autoReload = ref(localStorage.getItem(AUTO_KEY) !== '0')
const intervalSec = ref(Number(localStorage.getItem(INT_KEY)) || 30)
const selectedTags = ref<Set<string>>(loadTagFilter())
let timer: number | null = null

function loadTagFilter(): Set<string> {
  try {
    const raw = localStorage.getItem(TAG_KEY)
    if (!raw) return new Set()
    const arr = JSON.parse(raw)
    if (Array.isArray(arr)) return new Set(arr.filter(x => typeof x === 'string'))
  } catch {}
  return new Set()
}

function persistTagFilter() {
  if (selectedTags.value.size === 0) {
    localStorage.removeItem(TAG_KEY)
  } else {
    localStorage.setItem(TAG_KEY, JSON.stringify([...selectedTags.value]))
  }
}

function stopTimer() {
  if (timer !== null) {
    clearInterval(timer)
    timer = null
  }
}

function startTimer() {
  stopTimer()
  if (!autoReload.value) return
  const sec = Math.max(5, Number(intervalSec.value) || 30)
  timer = window.setInterval(() => {
    if (loading.value) return
    if (typeof document !== 'undefined' && document.visibilityState === 'hidden') return
    reload()
  }, sec * 1000)
}

async function reload() {
  loading.value = true
  try {
    cards.value = await api.dashboard()
  } finally {
    loading.value = false
  }
}

async function refreshAll() {
  if (loading.value) return
  loading.value = true
  try {
    const ids = Array.from(new Set(filteredCards.value.map(c => c.collector_id))).filter(Boolean)
    if (!ids.length) {
      cards.value = await api.dashboard()
      return
    }
    const results = await Promise.allSettled(ids.map(id => api.collectors.run(id, {})))
    const failed = results.filter(r => r.status === 'rejected').length
    cards.value = await api.dashboard()
    if (failed) {
      ElMessage.warning(`${ids.length - failed}/${ids.length} collectors refreshed; ${failed} failed`)
    } else {
      ElMessage.success(`refreshed ${ids.length} collector(s)`)
    }
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'refresh failed')
  } finally {
    loading.value = false
  }
}

function formatValue(c: Card): string {
  if (c.value_num !== null && c.value_num !== undefined) {
    return formatNum(c.value_num)
  }
  if (c.value_str) return c.value_str
  if (c.value_json) return c.value_json
  return '—'
}

function formatPrev(c: Card): string {
  if (c.prev_value_num !== null && c.prev_value_num !== undefined) {
    return formatNum(c.prev_value_num)
  }
  if (c.prev_value_str) return c.prev_value_str
  if (c.prev_value_json) return c.prev_value_json
  return '—'
}

function formatNum(n: number): string {
  if (Number.isInteger(n)) return String(n)
  return n.toFixed(Math.min(4, Math.max(0, 6 - String(Math.trunc(n)).length)))
}

function hasPrev(c: Card): boolean {
  return c.prev_value_num !== null && c.prev_value_num !== undefined
    || c.prev_value_str !== null && c.prev_value_str !== undefined
    || c.prev_value_json !== null && c.prev_value_json !== undefined
}

interface Delta {
  abs: number
  pct: number | null
  dir: 'up' | 'down' | 'flat'
}

function delta(c: Card): Delta | null {
  const cur = c.value_num
  const prev = c.prev_value_num
  if (cur === null || cur === undefined) return null
  if (prev === null || prev === undefined) return null
  const abs = cur - prev
  const dir: 'up' | 'down' | 'flat' = abs > 0 ? 'up' : abs < 0 ? 'down' : 'flat'
  const pct = prev !== 0 ? (abs / Math.abs(prev)) * 100 : null
  return { abs, pct, dir }
}

function formatDelta(d: Delta): string {
  if (d.dir === 'flat') return '0'
  const sign = d.abs > 0 ? '+' : ''
  const absStr = `${sign}${formatNum(d.abs)}`
  if (d.pct === null) return absStr
  const pctSign = d.pct > 0 ? '+' : ''
  return `${absStr} (${pctSign}${d.pct.toFixed(2)}%)`
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

const allTags = computed<string[]>(() => {
  const set = new Set<string>()
  for (const c of cards.value) {
    if (!c.site_tags) continue
    for (const t of c.site_tags) set.add(t)
  }
  return [...set].sort((a, b) => a.localeCompare(b))
})

function cardMatchesTags(c: Card): boolean {
  if (selectedTags.value.size === 0) return true
  const tags = c.site_tags
  if (!tags || !tags.length) return false
  for (const t of tags) {
    if (selectedTags.value.has(t)) return true
  }
  return false
}

function toggleTag(t: string) {
  const next = new Set(selectedTags.value)
  if (next.has(t)) next.delete(t)
  else next.add(t)
  selectedTags.value = next
  persistTagFilter()
}

function clearTags() {
  if (selectedTags.value.size === 0) return
  selectedTags.value = new Set()
  persistTagFilter()
}

const filteredCards = computed(() => cards.value.filter(cardMatchesTags))

const groupedBySite = computed(() => {
  const groups: Record<string, { siteId: number; siteName: string; baseUrl: string; cards: Card[] }> = {}
  for (const c of filteredCards.value) {
    const key = c.site_name || `Site #${c.site_id || 0}`
    if (!groups[key]) groups[key] = { siteId: c.site_id, siteName: key, baseUrl: c.site_base_url || '', cards: [] }
    else if (!groups[key].baseUrl && c.site_base_url) groups[key].baseUrl = c.site_base_url
    groups[key].cards.push(c)
  }
  return Object.values(groups).sort((a, b) => a.siteName.localeCompare(b.siteName))
})

const totalSiteCount = computed(() => {
  const ids = new Set<number | string>()
  for (const c of cards.value) {
    ids.add(c.site_name || `Site #${c.site_id || 0}`)
  }
  return ids.size
})

const filterActive = computed(() => selectedTags.value.size > 0)

onMounted(async () => {
  await reload()
  pruneSelectedTags()
  startTimer()
})

function pruneSelectedTags() {
  if (selectedTags.value.size === 0) return
  const known = new Set(allTags.value)
  const next = new Set<string>()
  let changed = false
  for (const t of selectedTags.value) {
    if (known.has(t)) next.add(t)
    else changed = true
  }
  if (changed) {
    selectedTags.value = next
    persistTagFilter()
  }
}

watch(autoReload, v => {
  localStorage.setItem(AUTO_KEY, v ? '1' : '0')
  startTimer()
})

watch(intervalSec, v => {
  localStorage.setItem(INT_KEY, String(v))
  startTimer()
})

watch(cards, () => pruneSelectedTags(), { deep: false })

onBeforeUnmount(stopTimer)
</script>

<template>
  <div>
    <div class="page-bar">
      <div>
        <h2>Dashboard</h2>
        <div class="subtitle">
          <template v-if="filterActive">
            {{ filteredCards.length }} of {{ cards.length }} indicator(s) across {{ groupedBySite.length }} of {{ totalSiteCount }} site(s)
          </template>
          <template v-else>
            {{ cards.length }} indicator(s) across {{ groupedBySite.length }} site(s)
          </template>
        </div>
      </div>
      <div class="page-bar-actions">
        <div class="auto-reload">
          <span class="lbl">Auto</span>
          <el-switch v-model="autoReload" />
          <el-select
            v-model="intervalSec"
            size="small"
            :disabled="!autoReload"
            style="width: 88px"
          >
            <el-option
              v-for="s in INTERVAL_OPTIONS"
              :key="s"
              :value="s"
              :label="`${s}s`"
            />
          </el-select>
        </div>
        <el-button :loading="loading" @click="reload">Reload</el-button>
        <el-button type="primary" :loading="loading" @click="refreshAll">Refresh All</el-button>
      </div>
    </div>

    <div v-if="allTags.length" class="tag-filter">
      <el-tag
        :effect="filterActive ? 'plain' : 'dark'"
        class="tag-filter-chip"
        :class="{ active: !filterActive }"
        @click="clearTags"
      >All</el-tag>
      <el-tag
        v-for="t in allTags"
        :key="t"
        :effect="selectedTags.has(t) ? 'dark' : 'plain'"
        class="tag-filter-chip"
        :class="{ active: selectedTags.has(t) }"
        @click="toggleTag(t)"
      >{{ t }}</el-tag>
    </div>

    <el-empty v-if="!cards.length && !loading" description="No indicators yet. Add some collectors first." />
    <el-empty
      v-else-if="filterActive && !filteredCards.length && !loading"
      :description="`No sites match the selected tag(s).`"
    />

    <el-row :gutter="16">
      <el-col
        v-for="group in groupedBySite"
        :key="group.siteName"
        :xs="24" :sm="24" :md="12" :lg="8" :xl="6"
      >
        <div class="site-card" :style="{ '--accent': siteAccent(group.siteId) } as any">
          <div class="site-header">
            <span class="site-dot"></span>
            <a
              v-if="group.baseUrl"
              class="site-name site-link"
              :href="group.baseUrl"
              target="_blank"
              rel="noopener noreferrer"
              :title="group.baseUrl"
              @click.stop
            >{{ group.siteName }}</a>
            <span v-else class="site-name">{{ group.siteName }}</span>
            <span class="site-count">{{ group.cards.length }}</span>
          </div>

          <div v-if="group.cards[0]?.site_tags?.length" class="site-tags">
            <el-tag
              v-for="t in group.cards[0].site_tags || []"
              :key="t"
              size="small"
              :effect="selectedTags.has(t) ? 'dark' : 'plain'"
              class="site-tag-chip"
              @click.stop="toggleTag(t)"
            >{{ t }}</el-tag>
          </div>

          <div class="metrics">
            <div v-for="c in group.cards" :key="c.indicator_id" class="metric">
              <div class="metric-head">
                <span class="metric-name">{{ c.name }}</span>
                <el-tag size="small" :type="statusType(c.last_status)" effect="plain">
                  {{ c.last_status || '—' }}
                </el-tag>
              </div>

              <div class="metric-value">
                <div class="metric-value-main">
                  <span class="num">{{ formatValue(c) }}</span>
                  <span v-if="c.unit" class="unit">{{ c.unit }}</span>
                </div>
                <div v-if="hasPrev(c)" class="metric-prev">
                  <span class="prev-val">prev {{ formatPrev(c) }}<span v-if="c.unit" class="prev-unit"> {{ c.unit }}</span></span>
                  <span
                    v-if="delta(c)"
                    class="delta"
                    :class="delta(c)!.dir"
                  >{{ formatDelta(delta(c)!) }}</span>
                </div>
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
          </div>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.subtitle { color: var(--sg-text-secondary); font-size: 13px; margin-top: 4px; }

.auto-reload {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-right: 8px;
}
.auto-reload .lbl {
  font-size: 12px;
  color: var(--sg-text-secondary);
}

.tag-filter {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 8px 0 14px;
  align-items: center;
}
.tag-filter-chip {
  cursor: pointer;
  user-select: none;
  transition: transform .12s;
}
.tag-filter-chip:hover {
  transform: translateY(-1px);
}

.site-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin: -4px 0 8px;
}
.site-tag-chip {
  cursor: pointer;
  user-select: none;
}
.site-card {
  position: relative;
  background: var(--sg-bg-card);
  border: 1px solid var(--sg-border-soft);
  border-radius: var(--sg-radius);
  padding: 14px 16px 12px;
  margin-bottom: 16px;
  overflow: hidden;
  transition: box-shadow .2s, transform .2s;
  box-shadow: var(--ep-box-shadow-light);
}
.site-card::before {
  content: '';
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 3px;
  background: var(--accent, #6366f1);
}
.site-card:hover {
  box-shadow: 0 8px 22px rgba(15, 23, 42, .08);
  transform: translateY(-1px);
}

.site-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 10px;
  margin-bottom: 10px;
  border-bottom: 1px solid var(--sg-border-soft);
}
.site-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent, #6366f1);
}
.site-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--sg-text-primary);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.site-link {
  color: inherit;
  text-decoration: none;
  cursor: pointer;
}
.site-link:hover {
  color: var(--accent, #6366f1);
  text-decoration: underline;
}
.site-count {
  font-size: 12px;
  color: var(--sg-text-secondary);
  background: var(--sg-aside-hover-bg);
  padding: 1px 8px;
  border-radius: 10px;
}

.metrics {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.metric {
  padding: 10px 0;
  border-bottom: 1px dashed var(--sg-border-soft);
}
.metric:last-child {
  border-bottom: none;
  padding-bottom: 2px;
}

.metric-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  gap: 6px;
}
.metric-name {
  font-size: 13px;
  color: var(--sg-text-secondary);
  font-weight: 500;
}

.metric-value {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 6px;
  min-height: 32px;
}
.metric-value-main {
  display: flex;
  align-items: baseline;
  gap: 6px;
  min-width: 0;
}
.metric-value .num {
  font-size: 24px;
  font-weight: 600;
  color: var(--sg-text-primary);
  line-height: 1;
  letter-spacing: -0.5px;
}
.metric-value .unit {
  font-size: 12px;
  color: var(--sg-text-secondary);
}
.metric-prev {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  font-size: 12px;
  line-height: 1.3;
  text-align: right;
  color: var(--sg-text-secondary);
  white-space: nowrap;
}
.metric-prev .prev-val {
  color: var(--sg-text-secondary);
}
.metric-prev .prev-unit {
  color: var(--sg-text-muted);
}
.metric-prev .delta {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}
.metric-prev .delta.up {
  color: var(--el-color-success);
}
.metric-prev .delta.down {
  color: var(--el-color-danger);
}
.metric-prev .delta.flat {
  color: var(--sg-text-muted);
}

.meta {
  font-size: 12px;
  color: var(--sg-text-secondary);
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  flex-wrap: wrap;
}
.meta .key {
  font-family: ui-monospace, "SF Mono", Menlo, monospace;
  background: var(--sg-aside-hover-bg);
  padding: 1px 6px;
  border-radius: 4px;
  color: var(--sg-text-primary);
}
.meta .dot { color: var(--sg-text-muted); }
.meta .collector {
  color: var(--sg-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 6px;
}
.ts {
  font-size: 11px;
  color: var(--sg-text-muted);
}
.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
}
.actions :deep(.el-button + .el-button) {
  margin-left: 4px;
}
.actions :deep(.el-button) {
  padding: 4px 6px;
  height: auto;
  font-size: 12px;
}
</style>
