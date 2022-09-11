<template>
  <div class="common-layout">
    <el-container>
      <el-header>Header</el-header>
      <el-container>
        <el-aside width="5%"><AppBar /></el-aside>
        <el-main><router-view /></el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script lang="ts" setup>
import AppBar from '@/views/appbar/AppBar.vue'
</script>

<style scoped>
.el-header {
  background-color: rgb(77, 77, 77);
}

.el-aside {
  height: 92vh;
}
</style>
