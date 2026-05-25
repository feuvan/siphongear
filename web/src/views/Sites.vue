<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'
import ResponsiveTable, { type RTColumn } from '@/components/ResponsiveTable.vue'

const rows = ref<any[]>([])
const dialog = ref(false)
const editing = ref<any>(null)
const form = reactive({ id: 0, name: '', base_url: '', tags: '', notes: '' })

const columns: RTColumn[] = [
  { key: 'id', label: 'ID', width: 80, hideOnMobile: true },
  { key: 'name', label: 'Name', primary: true },
  { key: 'base_url', label: 'Base URL' },
  { key: 'tags', label: 'Tags', width: 200 },
  { key: 'actions', label: 'Actions', slot: 'actions', width: 160 }
]

async function reload() { rows.value = await api.sites.list() }

function openCreate() {
  editing.value = null
  Object.assign(form, { id: 0, name: '', base_url: '', tags: '', notes: '' })
  dialog.value = true
}

function openEdit(row: any) {
  editing.value = row
  Object.assign(form, row)
  dialog.value = true
}

async function save() {
  if (typeof form.base_url === 'string') {
    form.base_url = form.base_url.trim().replace(/\/+$/, '')
  }
  if (form.id) await api.sites.update(form.id, form)
  else await api.sites.create(form)
  dialog.value = false
  await reload()
  ElMessage.success('saved')
}

async function remove(row: any) {
  await ElMessageBox.confirm(`Delete ${row.name}?`, 'Confirm')
  await api.sites.remove(row.id)
  await reload()
}

onMounted(reload)
</script>

<template>
  <div>
    <div class="page-bar">
      <h2>Sites</h2>
      <div class="page-bar-actions">
        <el-button type="primary" @click="openCreate">New Site</el-button>
      </div>
    </div>
    <ResponsiveTable :rows="rows" :columns="columns" row-key="id">
      <template #actions="{ row }">
        <el-button link @click="openEdit(row)">Edit</el-button>
        <el-button link type="danger" @click="remove(row)">Delete</el-button>
      </template>
    </ResponsiveTable>

    <el-dialog v-model="dialog" :title="form.id ? 'Edit Site' : 'New Site'" width="520px">
      <el-form label-width="100px">
        <el-form-item label="Name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="Base URL"><el-input v-model="form.base_url" /></el-form-item>
        <el-form-item label="Tags"><el-input v-model="form.tags" /></el-form-item>
        <el-form-item label="Notes"><el-input v-model="form.notes" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">Cancel</el-button>
        <el-button type="primary" @click="save">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>
