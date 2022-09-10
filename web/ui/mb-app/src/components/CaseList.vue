<template>
  <el-table :data="tableData" style="width: 100%">
    <el-table-column prop="protocol" label="协议类型" width="120" />
    <el-table-column prop="serviceMethod" label="请求类型" width="120" />
    <el-table-column prop="serviceName" label="接口/服务名" width="320" />
    <el-table-column prop="servicePath" label="接口路径" width="320" />
    <el-table-column prop="serviceDescription" label="serviceDescription" width="500" />
  </el-table>
</template>

<script lang="ts">
import { ref , defineComponent} from 'vue'
import {InterfaceTask} from '@/types/InterfaceTask'
import ResponseData from '@/types/ResponseData'
import CaseTaskDataService from '@/services/CaseTaskDataService';

// reference https://github.dev/bezkoder/vue-3-typescript-example/tree/master/src
// const tableData = ref([] as InterfaceTask[])
export default defineComponent({
  data() {
    return {
      tableData: ref([] as InterfaceTask[])
    }
  },
  methods: {
    refersList() {
      CaseTaskDataService.getAll()
      .then((response: ResponseData) => {
          this.tableData.push(...response.data)
        })
        .catch((e: Error) => {
          console.log(e);
        });
    }
  },
  mounted() {
    this.refersList();
  },
})
</script>

<style scoped>
.read-the-docs {
  color: #888;
}
</style>
