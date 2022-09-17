<template>
  <el-container>
    <el-header class="header">
      <div class="cap-state">
        <el-tooltip
          effect="dark"
          :content="startCapture ? '结束抓包' : '开始抓包'"
          placement="top-start"
        >
          <el-icon
            v-if="!startCapture"
            color="#00b848"
            :size="25"
            @click="changeCaptureState"
          >
            <VideoPlay />
          </el-icon>
          <el-icon
            v-else
            color="#ff3a3a"
            :size="25"
            @click="changeCaptureState"
          >
            <VideoPause />
          </el-icon>
        </el-tooltip>
      </div>

      <div class="delete-button">
        <el-tooltip effect="dark" content="清空" placement="top-start">
          <el-icon :size="25" color="#7f7f7f" @click="clearCapture()"
            ><Delete
          /></el-icon>
        </el-tooltip>
      </div>
    </el-header>
    <el-main>
      <el-table :data="tableData" style="width: 100%" max-height="500">
        <el-table-column
          fixed
          prop="request.protocol"
          label="协议类型"
          width="120"
        />
        <el-table-column prop="request.method" label="请求类型" width="120" />
        <el-table-column prop="request.host" label="Host" />
        <el-table-column prop="request.path" label="接口路径" />
        <el-table-column fixed="right" label="Operations" width="320">
          <template #default="scope">
            <el-button
              link
              type="success"
              size="small"
              @click="copyInfo(scope.row)"
            >
              复制
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>

<script lang="ts" setup>
import { ref, onBeforeUnmount } from 'vue'
import useClipboard from '@/utils/useClipboard'
import { Mock } from '@/types/InterfaceTask'
import { PushMessage } from '@/types/message'

const { toClipboard } = useClipboard()
const startCapture = ref(false)
const tableData = ref([] as Mock[])
let ws: WebSocket | null

function changeCaptureState() {
  startCapture.value = !startCapture.value
  if (startCapture.value) {
    connectServer()
  } else {
    ws?.close()
  }
}

function clearCapture() {
  tableData.value = []
}

async function copyInfo(scope: Mock) {
  try {
    await toClipboard(JSON.stringify(scope, null, 2))
    console.log('Copied to clipboard')
  } catch (e) {
    console.error(e)
  }
}

function addDataInfo(message: Mock) {
  tableData.value.push(message)
}

function connectServer() {
  ws = new WebSocket('ws://localhost:52162/connect')
  ws.onopen = function (evt) {
    // print('OPEN')
  }

  ws.onclose = function (evt) {
    ws = null
  }

  ws.onmessage = function (evt) {
    const d: PushMessage = JSON.parse(evt.data)
    addDataInfo(JSON.parse(d.content))
  }

  ws.onerror = function (evt) {
    console.error(evt)
  }
}

onBeforeUnmount(() => {
  ws?.close()
})
</script>

<style scoped>
.header {
  padding-top: 10px;
  padding-bottom: 10px;
  display: flex;
  align-items: center;
  border: 1px solid var(--el-border-color);
  border-radius: 0px;
  background-color: rgb(232, 232, 232);
}

.cap-state {
  height: 25px;
  display: inline-block;
  /* border: 1px solid; */
  border-radius: 5px;
}

.cap-state:hover {
  background-color: #d5d5d5;
}

.delete-button {
  height: 25px;
  display: inline-block;
  /* border: 1px solid; */
  border-radius: 5px;
  margin-left: 20px;
}

.delete-button:hover {
  background-color: #d5d5d5;
}
</style>
