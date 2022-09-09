<template>
  <el-table :data="tableData" style="width: 100%">
    <el-table-column prop="protocol" label="Protocol" width="120" />
    <el-table-column prop="serviceMethod" label="请求类型" width="120" />
    <el-table-column prop="serviceName" label="ServiceName" width="320" />
    <el-table-column prop="servicePath" label="servicePath" width="320" />
    <el-table-column prop="serviceDescription" label="serviceDescription" width="500" />
  </el-table>
</template>

<script lang="ts">
import { ref , defineComponent} from 'vue'
import {InterfaceTask} from '@/types/InterfaceTask'
import ResponseData from '@/types/ResponseData'

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
        fetch('http://127.0.0.1:6060/getCaseList')
    .then((data) => {
      return data.json()
    })
    .then((data: ResponseData) => {
      console.log(data.data)
      // 这里才能拿到数据，原因上面讲了
      this.tableData.push(...data.data)
    })
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
