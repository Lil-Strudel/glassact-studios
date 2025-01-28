import axios from "redaxios";

const api = axios.create({
  baseURL: "/api",
  withCredentials: true,
});

export default api;
