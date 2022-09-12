<template>
  <div>
    <el-row>
      <el-col :span="12">
        <CaseList :tableData="tableData" />
      </el-col>
      <el-col :span="12">
        <CaseDetail />
      </el-col>
    </el-row>
  </div>
</template>

<script lang="ts" setup>
// references 
// * https://blog.csdn.net/qq_27517377/article/details/123163381
// * https://vue3.chengpeiquan.com/communication.html#%E5%85%84%E5%BC%9F%E7%BB%84%E4%BB%B6%E9%80%9A%E4%BF%A1
import { onMounted, ref } from 'vue'
import ArrayUtils from '@/utils/array'
import CaseDetail from './CaseDetail.vue'
import CaseList from './CaseList.vue'
import ResponseData from '@/types/ResponseData'
import { ItfTaskWithFileInfo } from '@/types/InterfaceTask'
import CaseTaskDataService from '@/services/CaseTaskDataService'

type TbType = {
  dir: string
  itfs: ItfTaskWithFileInfo[]
}

onMounted(() => {
  refersList()
})

const tableData = ref([] as TbType[])

function refersList() {
  CaseTaskDataService.getAll()
    .then((response: ResponseData) => {
      let arr: ItfTaskWithFileInfo[] = response.data
      let tbData = [] as TbType[]

      ArrayUtils.groupMapBy(arr, (item) => item.dirpath)?.forEach((v, k) => {
        tbData.push({
          // dir: k.substring(k.lastIndexOf('/') + 1),
          dir: k,
          itfs: v
        })
      })

      tableData.value.push(...tbData)
    })
    .catch((e: Error) => {
      console.error(e)
    })
}
</script>

<style lang="scss" scoped></style>
