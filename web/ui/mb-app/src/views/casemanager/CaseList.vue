<template>
  <el-table :data="filterTableData" style="width: 100%" :border="true">
    <!-- 表的内容 -->
    <el-table-column label="文件夹路径" prop="dir" />
    <!-- 展示全部接口 -->
    <el-table-column type="expand">
      <template #default="props">
        <!-- 这层显示全部的接口-->
        <el-table
          :data="props.row.itfs"
          table-layout="auto"
          style="background-color: rgb(232, 228, 175)"
          :header-cell-style="{
            background: 'rgb(169 203 162)',
            color: '#777266'
          }"
          :row-style="{
            background: 'rgb(225 237 235)'
          }"
        >
          <el-table-column prop="protocol" label="协议类型"/>
          <el-table-column prop="serviceMethod" label="请求类型"/>
          <el-table-column prop="serviceName" label="接口/服务名"/>
          <el-table-column prop="servicePath" label="接口路径"/>
          <el-table-column
            prop="serviceDescription"
            label="接口描述"
          />
          <!-- 显示接口下面的全部用例 -->
          <el-table-column type="expand" class="itf-case-type">
            <template #default="props02">
              <el-table
                :data="props02.row.cases"
                :header-cell-style="{
                  background: 'rgb(232 216 230)',
                  color: '#606266'
                }"
                :cell-style="refreshRunStatus"
                :row-style="{
                  background: 'rgb(253 235 251)'
                }"
              >
                <el-table-column prop="name" label="用例名称" />
                <el-table-column
                  prop="description"
                  label="用例描述"
                />
                <!-- 运行状态 -->
                <el-table-column label="执行状态">
                  <template #default="scope">
                    <span v-text="getRunStatus(scope.row)"></span>
                  </template>
                </el-table-column>
                <!-- 操作 -->
                <el-table-column label="Operations"  width="220">
                  <template #default="scope">
                    <el-button size="small" @click="showDetails(scope.row)"
                      >显示详情</el-button
                    >
                    <el-button
                      size="small"
                      type="danger"
                      @click="
                        runSingleCase(scope.index, props02.row, scope.row)
                      "
                      >执行用例</el-button
                    >
                  </template>
                </el-table-column>
              </el-table>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </el-table-column>
    <!-- 右边的布局 -->
    <el-table-column align="right">
      <template #header>
        <el-input v-model="search" size="small" placeholder="搜索服务名称" />
      </template>
    </el-table-column>
  </el-table>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue'
import { ItfTaskWithFileInfo, Case } from '@/types/InterfaceTask'
import bus from '@/libs/bus'
import ResponseData from '@/types/ResponseData'
import CaseTaskDataService from '@/services/CaseTaskDataService'

type TbType = {
  dir: string
  itfs: ItfTaskWithFileInfo[]
}

type RunTaskReply = {
  status: number
  failedReason: string
}

const receiver = defineProps({
  tableData: {
    default: [] as TbType[]
  }
})

const runStatusMap = ref(new Map<string, RunTaskReply>())
const search = ref('')
// const tableData = ref([] as TbType[])
const filterTableData = computed(() => {
  return receiver.tableData.filter((data) => {
    if (!search.value) {
      return true
    }

    for (const itf of data.itfs) {
      if (itf.serviceName.toLowerCase().includes(search.value.toLowerCase())) {
        return true
      }

      for (const c of itf.cases) {
        if (c.name.toLowerCase().includes(search.value.toLowerCase())) {
          return true
        }
      }
    }
    return false
  })
})

function getRunStatus(row: Case) {
  let r = runStatusMap.value.get(row.name)
  if (!r) {
    return '尚未执行'
  }

  if (r.status === 1) {
    return '用例执行成功'
  } else {
    return r.failedReason
  }
}

function showDetails(row: Case) {
  bus.emit('caseDetail', JSON.stringify(row, null, 2))
}

function refreshRunStatus(c: {
  column: any
  row: Case
  rowIndex: number
  columnIndex: number
}) {
  if (c.columnIndex !== 2) {
    return
  }

  let r = runStatusMap.value.get(c.row.name)
  if (!r) {
    return
  }

  if (r.status === 1) {
    return {
      background: 'rgb(120 230 126)'
    }
  } else {
    return {
      background: 'rgb(246 122 110)'
    }
  }
}

function runSingleCase(index: number, itf: ItfTaskWithFileInfo, c: Case) {
  CaseTaskDataService.runSingleCase(itf.serviceName, c.name).then(
    (response: ResponseData) => {
      console.log(response)
      runStatusMap.value.set(c.name, response.data)
    }
  )
}

// reference https://github.dev/bezkoder/vue-3-typescript-example/tree/master/src
</script>

<style scoped>
.read-the-docs {
  color: #888;
}

.itf-case-type {
  background-color: rgb(120, 237, 239);
}
</style>
