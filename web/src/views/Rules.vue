<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'
import ResponsiveTable, { type RTColumn } from '@/components/ResponsiveTable.vue'

interface Condition {
  type: 'compare'
  op: 'lt' | 'lte' | 'gt' | 'gte' | 'eq' | 'ne'
  value: number
}

interface Action {
  type: 'indicator_color'
}

interface RuleRow {
  id: number
  name: string
  enabled: boolean
  priority: number
  indicator_key: string
  target_type: 'all' | 'tags'
  target_tags: string
  target_tags_arr: string[]
  conditions: Condition[]
  actions: Action[]
  created_at?: string
  updated_at?: string
}

const OP_OPTIONS: { value: Condition['op']; label: string }[] = [
  { value: 'lt', label: '<' },
  { value: 'lte', label: '≤' },
  { value: 'gt', label: '>' },
  { value: 'gte', label: '≥' },
  { value: 'eq', label: '=' },
  { value: 'ne', label: '≠' }
]

const ACTION_OPTIONS = [
  { value: 'indicator_color', label: 'Indicator color (warning)' }
]

const rows = ref<RuleRow[]>([])
const dialog = ref(false)
const indicatorKeys = ref<string[]>([])
const allSiteTags = ref<string[]>([])
const tagInput = ref('')
const tagInputVisible = ref(false)
const tagInputRef = ref<any>(null)

const form = reactive({
  id: 0,
  name: '',
  enabled: true,
  priority: 100,
  indicator_key: '',
  target_type: 'all' as 'all' | 'tags',
  target_tags: [] as string[],
  cond_op: 'lt' as Condition['op'],
  cond_value: 0,
  action_type: 'indicator_color' as Action['type']
})

const columns: RTColumn[] = [
  { key: 'id', label: 'ID', width: 80, hideOnMobile: true },
  { key: 'name', label: 'Name', primary: true, slot: 'name' },
  { key: 'indicator_key', label: 'Indicator', width: 160 },
  { key: 'condition', label: 'Condition', width: 160, slot: 'condition' },
  { key: 'action', label: 'Action', width: 200, slot: 'action' },
  { key: 'target', label: 'Target', width: 200, slot: 'target' },
  { key: 'priority', label: 'Prio', width: 80 },
  { key: 'enabled', label: 'On', width: 80, slot: 'enabled' },
  { key: 'actions', label: 'Actions', slot: 'actions', width: 160 }
]

function opLabel(op: string): string {
  const o = OP_OPTIONS.find(x => x.value === op)
  return o ? o.label : op
}

function conditionSummary(r: RuleRow): string {
  const c = r.conditions?.[0]
  if (!c) return '—'
  return `${opLabel(c.op)} ${c.value}`
}

function actionSummary(r: RuleRow): string {
  const a = r.actions?.[0]
  if (!a) return '—'
  const o = ACTION_OPTIONS.find(x => x.value === a.type)
  return o ? o.label : a.type
}

function targetSummary(r: RuleRow): string {
  if (r.target_type === 'all') return 'All sites'
  const tags = r.target_tags_arr || []
  if (!tags.length) return 'Tags: (none)'
  return `Tags: ${tags.join(', ')}`
}

async function reload() {
  rows.value = await api.rules.list()
}

async function loadHints() {
  try {
    const cards: any[] = await api.dashboard()
    const keys = new Set<string>()
    for (const c of cards) if (c.key) keys.add(c.key)
    indicatorKeys.value = [...keys].sort()
  } catch {}
  try {
    const sites: any[] = await api.sites.list()
    const tags = new Set<string>()
    for (const s of sites) {
      if (!s.tags) continue
      for (const part of String(s.tags).split(',')) {
        const t = part.trim()
        if (t) tags.add(t)
      }
    }
    allSiteTags.value = [...tags].sort()
  } catch {}
}

function openCreate() {
  Object.assign(form, {
    id: 0,
    name: '',
    enabled: true,
    priority: 100,
    indicator_key: '',
    target_type: 'all',
    target_tags: [],
    cond_op: 'lt',
    cond_value: 0,
    action_type: 'indicator_color'
  })
  tagInput.value = ''
  tagInputVisible.value = false
  dialog.value = true
}

function openEdit(row: RuleRow) {
  const c = row.conditions?.[0]
  const a = row.actions?.[0]
  Object.assign(form, {
    id: row.id,
    name: row.name,
    enabled: row.enabled,
    priority: row.priority,
    indicator_key: row.indicator_key,
    target_type: row.target_type,
    target_tags: [...(row.target_tags_arr || [])],
    cond_op: c?.op || 'lt',
    cond_value: c?.value ?? 0,
    action_type: a?.type || 'indicator_color'
  })
  tagInput.value = ''
  tagInputVisible.value = false
  dialog.value = true
}

function showTagInput() {
  tagInputVisible.value = true
  nextTick(() => tagInputRef.value?.focus?.())
}

function commitTagInput() {
  const v = tagInput.value.trim()
  if (v && !form.target_tags.includes(v)) {
    form.target_tags.push(v)
  }
  tagInput.value = ''
  tagInputVisible.value = false
}

function onTagInputKey(e: KeyboardEvent) {
  if (e.key === ',') {
    e.preventDefault()
    commitTagInput()
    tagInputVisible.value = true
    nextTick(() => tagInputRef.value?.focus?.())
  } else if (e.key === 'Backspace' && tagInput.value === '' && form.target_tags.length) {
    form.target_tags.pop()
  }
}

function removeTag(t: string) {
  form.target_tags = form.target_tags.filter(x => x !== t)
}

function pickSuggestedTag(t: string) {
  if (!form.target_tags.includes(t)) form.target_tags.push(t)
}

const suggestedTagsRemaining = computed(() =>
  allSiteTags.value.filter(t => !form.target_tags.includes(t))
)

async function save() {
  if (!form.name.trim()) {
    ElMessage.error('name is required')
    return
  }
  if (!form.indicator_key.trim()) {
    ElMessage.error('indicator key is required')
    return
  }
  if (tagInput.value.trim()) commitTagInput()
  if (form.target_type === 'tags' && form.target_tags.length === 0) {
    ElMessage.error('select at least one tag, or switch target to "All sites"')
    return
  }
  const body = {
    name: form.name.trim(),
    enabled: form.enabled,
    priority: Number(form.priority) || 0,
    indicator_key: form.indicator_key.trim(),
    target_type: form.target_type,
    target_tags: form.target_type === 'tags' ? form.target_tags : [],
    conditions: [{ type: 'compare', op: form.cond_op, value: Number(form.cond_value) }],
    actions: [{ type: form.action_type }]
  }
  try {
    if (form.id) await api.rules.update(form.id, body)
    else await api.rules.create(body)
    dialog.value = false
    await reload()
    ElMessage.success('saved')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'save failed')
  }
}

async function remove(row: RuleRow) {
  await ElMessageBox.confirm(`Delete rule "${row.name}"?`, 'Confirm')
  await api.rules.remove(row.id)
  await reload()
}

async function toggleEnabled(row: RuleRow) {
  try {
    const fresh = await api.rules.get(row.id)
    const body = {
      name: fresh.name,
      enabled: !fresh.enabled,
      priority: fresh.priority,
      indicator_key: fresh.indicator_key,
      target_type: fresh.target_type,
      target_tags: fresh.target_tags_arr || [],
      conditions: fresh.conditions || [],
      actions: fresh.actions || []
    }
    await api.rules.update(row.id, body)
    await reload()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'toggle failed')
  }
}

onMounted(async () => {
  await Promise.all([reload(), loadHints()])
})
</script>

<template>
  <div>
    <div class="page-bar">
      <div>
        <h2>Rules</h2>
        <div class="subtitle">Threshold rules visually flag indicators on the Dashboard.</div>
      </div>
      <div class="page-bar-actions">
        <el-button type="primary" @click="openCreate">New Rule</el-button>
      </div>
    </div>

    <ResponsiveTable :rows="rows" :columns="columns" row-key="id">
      <template #name="{ row }">
        <span :class="{ 'rule-disabled': !row.enabled }">{{ row.name }}</span>
      </template>
      <template #condition="{ row }">{{ conditionSummary(row) }}</template>
      <template #action="{ row }">{{ actionSummary(row) }}</template>
      <template #target="{ row }">
        <span v-if="row.target_type === 'all'">All sites</span>
        <span v-else>
          <el-tag
            v-for="t in row.target_tags_arr || []"
            :key="t"
            size="small"
            effect="plain"
            class="tag-cell-item"
          >{{ t }}</el-tag>
          <span v-if="!(row.target_tags_arr || []).length" class="tag-cell-empty">(none)</span>
        </span>
      </template>
      <template #enabled="{ row }">
        <el-switch :model-value="row.enabled" @click.stop="toggleEnabled(row)" />
      </template>
      <template #actions="{ row }">
        <el-button link @click="openEdit(row)">Edit</el-button>
        <el-button link type="danger" @click="remove(row)">Delete</el-button>
      </template>
    </ResponsiveTable>

    <el-dialog v-model="dialog" :title="form.id ? 'Edit Rule' : 'New Rule'" width="640px">
      <el-form label-width="120px">
        <el-form-item label="Name">
          <el-input v-model="form.name" placeholder="e.g., balance low" />
        </el-form-item>

        <el-form-item label="Enabled">
          <el-switch v-model="form.enabled" />
        </el-form-item>

        <el-form-item label="Priority">
          <el-input-number v-model="form.priority" :min="0" :max="9999" :step="10" />
          <span class="form-hint">Smaller priority is evaluated first.</span>
        </el-form-item>

        <el-form-item label="Indicator key">
          <el-autocomplete
            v-model="form.indicator_key"
            :fetch-suggestions="(q: string, cb: any) => cb(
              indicatorKeys
                .filter(k => !q || k.toLowerCase().includes(q.toLowerCase()))
                .map(k => ({ value: k }))
            )"
            placeholder="e.g., balance"
            style="width: 240px"
          />
        </el-form-item>

        <el-form-item label="Condition">
          <div class="cond-row">
            <span class="lhs">value</span>
            <el-select v-model="form.cond_op" style="width: 100px">
              <el-option
                v-for="op in OP_OPTIONS"
                :key="op.value"
                :value="op.value"
                :label="op.label"
              />
            </el-select>
            <el-input-number v-model="form.cond_value" :step="1" controls-position="right" />
          </div>
        </el-form-item>

        <el-form-item label="Action">
          <el-select v-model="form.action_type" style="width: 280px">
            <el-option
              v-for="a in ACTION_OPTIONS"
              :key="a.value"
              :value="a.value"
              :label="a.label"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="Target">
          <el-radio-group v-model="form.target_type">
            <el-radio value="all">All sites</el-radio>
            <el-radio value="tags">Sites with tag(s)</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="form.target_type === 'tags'" label="Tags">
          <div class="tag-editor">
            <el-tag
              v-for="t in form.target_tags"
              :key="t"
              closable
              size="small"
              effect="plain"
              @close="removeTag(t)"
            >{{ t }}</el-tag>
            <el-input
              v-if="tagInputVisible"
              ref="tagInputRef"
              v-model="tagInput"
              size="small"
              class="tag-editor-input"
              @keydown="onTagInputKey"
              @keydown.enter.prevent="commitTagInput"
              @blur="commitTagInput"
            />
            <el-button v-else size="small" plain @click="showTagInput">+ Tag</el-button>
          </div>
          <div v-if="suggestedTagsRemaining.length" class="tag-hints">
            <span class="form-hint">Existing site tags:</span>
            <el-tag
              v-for="t in suggestedTagsRemaining"
              :key="t"
              size="small"
              effect="plain"
              class="tag-hint-item"
              @click="pickSuggestedTag(t)"
            >+ {{ t }}</el-tag>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">Cancel</el-button>
        <el-button type="primary" @click="save">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.subtitle {
  color: var(--sg-text-secondary);
  font-size: 13px;
  margin-top: 4px;
}
.rule-disabled {
  color: var(--sg-text-muted);
  text-decoration: line-through;
}
.tag-cell-item {
  margin-right: 4px;
}
.tag-cell-empty {
  color: var(--sg-text-muted);
}
.cond-row {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.cond-row .lhs {
  color: var(--sg-text-secondary);
}
.tag-editor {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.tag-editor-input {
  width: 120px;
}
.tag-hints {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
  align-items: center;
}
.tag-hint-item {
  cursor: pointer;
}
.form-hint {
  font-size: 12px;
  color: var(--sg-text-secondary);
  margin-left: 8px;
}
</style>
