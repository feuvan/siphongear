<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'

const props = defineProps<{ collectorId: number, indicators: any[] }>()
const emit = defineEmits<{ (e: 'update:indicators', v: any[]): void }>()

const rows = ref<any[]>([])
const dialog = ref(false)
const form = reactive<any>({ id: 0, key: '', name: '', type: 'number', unit: '', display: 'gauge' })

async function reload() {
  rows.value = await api.indicators.list(props.collectorId)
  emit('update:indicators', rows.value)
}

function openCreate() {
  Object.assign(form, { id: 0, key: '', name: '', type: 'number', unit: '', display: 'gauge' })
  dialog.value = true
}

function openEdit(row: any) {
  Object.assign(form, row)
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

onMounted(() => { rows.value = props.indicators || [] })
</script>

<template>
  <div>
    <div style="margin-bottom: 12px">
      <el-button type="primary" @click="openCreate">New Indicator</el-button>
      <el-text type="info" style="margin-left: 12px;">
        Indicators reference variables produced by your pipeline (extract step). Bind them by var name (key).
      </el-text>
    </div>
    <el-table :data="rows" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="key" label="Key (var name)" width="160" />
      <el-table-column prop="name" label="Display Name" />
      <el-table-column prop="type" label="Type" width="120" />
      <el-table-column prop="unit" label="Unit" width="100" />
      <el-table-column prop="display" label="Display" width="120" />
      <el-table-column label="Actions" width="180">
        <template #default="{ row }">
          <el-button link @click="openEdit(row)">Edit</el-button>
          <el-button link type="danger" @click="remove(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

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
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">Cancel</el-button>
        <el-button type="primary" @click="save">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>
