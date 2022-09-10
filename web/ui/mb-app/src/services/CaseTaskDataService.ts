import http from '@/http-common'
import ResponseData from '@/types/ResponseData'

/* eslint-disable */
class CaseTaskDataService {
  getAll(): Promise<ResponseData> {
    return http.get('/getCaseList').then(response => response.data)
  }
}

export default new CaseTaskDataService();