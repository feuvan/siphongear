<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
auth.loadFromStorage()

const isLogin = computed(() => route.name === 'login')

const activeMenu = computed(() => String(route.name || ''))

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <el-container v-if="!isLogin" class="app-shell">
    <el-aside width="220px" class="aside">
      <div class="brand">SiphonGear</div>
      <el-menu :default-active="activeMenu" router>
        <el-menu-item index="dashboard" :route="{ name: 'dashboard' }">
          <el-icon><Odometer /></el-icon>
          <span>Dashboard</span>
        </el-menu-item>
        <el-menu-item index="collectors" :route="{ name: 'collectors' }">
          <el-icon><Connection /></el-icon>
          <span>Collectors</span>
        </el-menu-item>
        <el-menu-item index="sites" :route="{ name: 'sites' }">
          <el-icon><Globe /></el-icon>
          <span>Sites</span>
        </el-menu-item>
        <el-menu-item index="credentials" :route="{ name: 'credentials' }">
          <el-icon><Lock /></el-icon>
          <span>Credentials</span>
        </el-menu-item>
        <el-menu-item index="settings" :route="{ name: 'settings' }">
          <el-icon><Setting /></el-icon>
          <span>Settings</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="spacer"></div>
        <el-dropdown>
          <span class="user">
            {{ auth.user?.username || 'guest' }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="router.push({ name: 'settings' })">Settings</el-dropdown-item>
              <el-dropdown-item divided @click="logout">Logout</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
  <router-view v-else />
</template>

<style>
html, body, #app { height: 100%; margin: 0; }
.app-shell { height: 100vh; }
.aside { background: #001529; color: #fff; }
.brand { padding: 16px; font-size: 18px; font-weight: 600; color: #fff; }
.aside .el-menu { background: #001529; border: 0; }
.aside .el-menu-item { color: #cdd6e3; }
.aside .el-menu-item.is-active { background: #1f3d6b; color: #fff; }
.header { display: flex; align-items: center; border-bottom: 1px solid #eaecef; background: #fff; }
.spacer { flex: 1; }
.user { cursor: pointer; }
</style>
