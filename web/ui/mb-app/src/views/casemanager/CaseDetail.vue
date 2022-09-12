<template>
    <div
        class="radius"
        style="{ borderRadius: var(--el-border-radius-base) }"
      >
      {{ caseData }}
      </div>
</template>

<script lang="ts" setup>
import { ref, onBeforeUnmount } from "vue";
import bus from '@/libs/bus'

const caseData = ref('')

// 启用监听
bus.on('caseDetail', changeCaseData);

// 在组件卸载之前移除监听
onBeforeUnmount( () => {
  bus.off('caseDetail', changeCaseData);
})

function changeCaseData(data: string) {
    caseData.value = data;
}

</script>

<style scoped>
.radius {
  height: 85vh;
  width: 95%;
  border: 1px solid var(--el-border-color);
  border-radius: 0;
  margin-left: 20px;
  /* 保留字符的换行符和空格 */
  white-space: pre-wrap;
}
</style>
