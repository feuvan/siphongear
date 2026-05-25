<script setup lang="ts">
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { api } from '@/api'

const form = reactive({ old_password: '', new_password: '', confirm: '' })

async function changePassword() {
  if (form.new_password !== form.confirm) {
    ElMessage.error('passwords do not match')
    return
  }
  if (form.new_password.length < 6) {
    ElMessage.error('password must be at least 6 chars')
    return
  }
  try {
    await api.changePassword(form.old_password, form.new_password)
    ElMessage.success('password changed')
    form.old_password = form.new_password = form.confirm = ''
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'change failed')
  }
}
</script>

<template>
  <div>
    <div class="page-bar">
      <h2>Settings</h2>
    </div>
    <el-card>
      <h3 class="card-title">Change Password</h3>
      <el-form label-width="160px" style="max-width: 480px">
        <el-form-item label="Current Password">
          <el-input v-model="form.old_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="New Password">
          <el-input v-model="form.new_password" type="password" show-password />
        </el-form-item>
        <el-form-item label="Confirm New Password">
          <el-input v-model="form.confirm" type="password" show-password />
        </el-form-item>
        <el-button type="primary" @click="changePassword">Update</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.card-title { margin: 0 0 16px; font-size: 16px; font-weight: 600; color: var(--sg-text-primary); }
</style>
