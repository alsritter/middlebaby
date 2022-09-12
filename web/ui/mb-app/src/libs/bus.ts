import mitt from 'mitt';
type Events = {
  caseDetail: string; // 改变用例详情时发送事件
};

const emitter = mitt<Events>();

export default emitter;