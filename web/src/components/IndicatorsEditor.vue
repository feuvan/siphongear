<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'
import ResponsiveTable, { type RTColumn } from '@/components/ResponsiveTable.vue'

const props = defineProps<{ collectorId: number, indicators: any[] }>()
const emit = defineEmits<{ (e: 'update:indicators', v: any[]): void }>()

const rows = ref<any[]>([])
const dialog = ref(false)
const form = reactive<any>({ id: 0, key: '', name: '', type: 'number', unit: '', display: 'gauge', hidden: false })

const columns: RTColumn[] = [
  { key: 'id', label: 'ID', width: 70, hideOnMobile: true },
  { key: 'key', label: 'Key (var name)', primary: true, width: 160 },
  { key: 'name', label: 'Display Name' },
  { key: 'type', label: 'Type', width: 120 },
  { key: 'unit', label: 'Unit', width: 100 },
  { key: 'display', label: 'Display', width: 120 },
  { key: 'dashboard', label: 'Dashboard', width: 110, slot: 'dashboard' },
  { key: 'actions', label: 'Actions', width: 180, slot: 'actions' }
]

async function reload() {
  rows.value = await api.indicators.list(props.collectorId)
  emit('update:indicators', rows.value)
}

function openCreate() {
  Object.assign(form, { id: 0, key: '', name: '', type: 'number', unit: '', display: 'gauge', hidden: false })
  dialog.value = true
}

function openEdit(row: any) {
  Object.assign(form, { hidden: false }, row)
  dialog.value = true
}

async function save() {
  if (form.id) await api.indicators.update(form.id, form)
  else await api.indicators.create(props.collectorId, form)
  dialog.value = false
  await reload()
  ElMessage.success('saved')
}

async function remove(row: any) {
  await ElMessageBox.confirm(`Delete indicator ${row.key}?`, 'Confirm')
  await api.indicators.remove(row.id)
  await reload()
}

async function toggleVisible(row: any, visible: boolean) {
  try {
    await api.indicators.update(row.id, { ...row, hidden: !visible })
    await reload()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'update failed')
  }
}

watch(
  () => props.indicators,
  v => { rows.value = Array.isArray(v) ? [...v] : [] },
  { immediate: true }
)
</script>

<template>
  <div>
    <div class="indicators-toolbar">
      <el-button type="primary" @click="openCreate">New Indicator</el-button>
      <el-text type="info" class="indicators-hint">
        Indicators reference variables produced by your pipeline (extract step). Bind them by var name (key).
      </el-text>
    </div>
    <ResponsiveTable :rows="rows" :columns="columns" row-key="id">
      <template #dashboard="{ row }">
        <el-switch
          :model-value="!row.hidden"
          @change="(v: boolean) => toggleVisible(row, v)"
        />
      </template>
      <template #actions="{ row }">
        <el-button link @click="openEdit(row)">Edit</el-button>
        <el-button link type="danger" @click="remove(row)">Delete</el-button>
      </template>
    </ResponsiveTable>

    <el-dialog v-model="dialog" :title="form.id ? 'Edit Indicator' : 'New Indicator'" width="520px">
      <el-form label-width="120px">
        <el-form-item label="Key"><el-input v-model="form.key" placeholder="var name from extract step" /></el-form-item>
        <el-form-item label="Name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="Type">
          <el-select v-model="form.type">
            <el-option label="number" value="number" />
            <el-option label="string" value="string" />
            <el-option label="bool" value="bool" />
            <el-option label="json" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="Unit"><el-input v-model="form.unit" /></el-form-item>
        <el-form-item label="Display">
          <el-select v-model="form.display">
            <el-option label="gauge" value="gauge" />
            <el-option label="line" value="line" />
            <el-option label="table" value="table" />
          </el-select>
        </el-form-item>
        <el-form-item label="Show on Dashboard">
          <el-switch :model-value="!form.hidden" @change="(v: boolean) => form.hidden = !v" />
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
.indicators-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px 12px;
  margin-bottom: 12px;
}
.indicators-hint {
  flex: 1;
  min-width: 0;
}
@media (max-width: 768px) {
  .indicators-hint {
    flex: 1 1 100%;
  }
}
</style>
