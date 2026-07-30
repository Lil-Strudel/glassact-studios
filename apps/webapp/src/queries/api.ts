import axios from "axios";

const api = axios.create({
  baseURL: "/api",
  withCredentials: true,
});

api.interceptors.response.use((res) => {
  const contentType = res.headers["content-type"];
  if (typeof contentType === "string" && contentType.includes("text/html")) {
    return Promise.reject(
      new Error(`Expected JSON from ${res.config.url} but received HTML`),
    );
  }
  return res;
});

export default api;
