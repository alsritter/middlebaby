<template>
  <el-menu router :collapse="true">
    <el-menu-item index="/case">
      <el-icon><Grid /></el-icon>
      <template #title>测试用例</template>
    </el-menu-item>
    <el-menu-item index="/datamanager">
      <el-icon><document /></el-icon>
      <template #title>Mock 服务</template>
    </el-menu-item>
    <el-menu-item index="/setting">
      <el-icon><setting /></el-icon>
      <template #title>初始化配置</template>
    </el-menu-item>
  </el-menu>
</template>

<script lang="ts" setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
const router = useRouter()

onMounted(() => {
  router.push({
    path: '/case'
  })
})
</script>

<style></style>
