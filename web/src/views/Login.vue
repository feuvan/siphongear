<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const auth = useAuthStore()
const form = reactive({ username: '', password: '' })

async function submit() {
  try {
    await auth.login(form.username, form.password)
    router.push({ name: 'dashboard' })
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'login failed')
  }
}
</script>

<template>
  <div class="login">
    <el-card class="card">
      <h2>SiphonGear</h2>
      <el-form @submit.prevent="submit" label-position="top">
        <el-form-item label="Username">
          <el-input v-model="form.username" autofocus />
        </el-form-item>
        <el-form-item label="Password">
          <el-input v-model="form.password" type="password" show-password @keyup.enter="submit" />
        </el-form-item>
        <el-button type="primary" @click="submit" style="width: 100%">Sign In</el-button>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.login { display: flex; align-items: center; justify-content: center; height: 100vh; background: #f0f2f5; }
.card { width: 360px; padding: 8px 16px; }
h2 { text-align: center; margin: 8px 0 24px; }
</style>
