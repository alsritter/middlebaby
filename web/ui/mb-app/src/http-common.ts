import axios, { AxiosInstance } from "axios";

const apiClient: AxiosInstance = axios.create({
  baseURL: "http://localhost:6060/v1",
  headers: {
    "Content-type": "application/json",
  },
});

export default apiClient;
