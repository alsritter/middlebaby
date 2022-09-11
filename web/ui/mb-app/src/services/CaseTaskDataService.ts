import http from '@/http-common'
import ResponseData from '@/types/ResponseData'

/* eslint-disable */
class CaseTaskDataService {
  getAll(): Promise<ResponseData> {
    return http.get('/getCaseList').then(response => response.data)
  }

  runSingleCase(itfName: string, caseName: string): Promise<ResponseData> {
    let bodyFormData = new FormData();
    bodyFormData.append('itfName', itfName)
    bodyFormData.append('caseName', caseName)

    return http.post("/runSingleCase", bodyFormData).then(response => response.data)
  }
}

export default new CaseTaskDataService();