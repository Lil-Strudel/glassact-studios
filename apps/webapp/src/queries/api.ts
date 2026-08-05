import axios, {
  AxiosError,
  InternalAxiosRequestConfig,
  isAxiosError,
} from "axios";
import { getValidAccessToken, refreshAccessToken } from "./access-token";

const api = axios.create({
  baseURL: "/api",
  withCredentials: true,
});

// /auth/* authenticates with the refresh cookie and is what mints the bearer in
// the first place, so attaching one would recurse.
function isAuthRoute(url: string | undefined) {
  return url?.startsWith("/auth/") ?? false;
}

type RetriableConfig = InternalAxiosRequestConfig & {
  retriedAfterRefresh?: boolean;
};

// Resolving the token per request rather than pinning it once means no timer
// can leave a stale bearer behind — a throttled or sleeping tab picks up a
// fresh one the moment it makes its next call.
api.interceptors.request.use(async (config) => {
  if (isAuthRoute(config.url)) {
    return config;
  }

  const { token } = await getValidAccessToken();
  config.headers.Authorization = `Bearer ${token}`;

  return config;
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

// A token can still be rejected despite looking valid — clock skew against the
// server, or revocation. Refresh once and replay rather than surfacing the 401.
api.interceptors.response.use(undefined, async (error: unknown) => {
  if (!isAxiosError(error)) {
    return Promise.reject(error);
  }

  const config = (error as AxiosError).config as RetriableConfig | undefined;

  if (
    error.response?.status !== 401 ||
    !config ||
    config.retriedAfterRefresh ||
    isAuthRoute(config.url)
  ) {
    return Promise.reject(error);
  }

  config.retriedAfterRefresh = true;

  try {
    await refreshAccessToken();
  } catch {
    // The original 401 is the meaningful error for the caller; the refresh
    // failure is already handled by the session-expired hook.
    return Promise.reject(error);
  }

  return api(config);
});

export default api;
