import axios from "axios";

// Treat a token as spent a minute before the server expiry so one can never die
// between being attached and being read. Capped at half the token's lifetime so
// a short-lived token is still reused rather than refreshed on every request.
const EXPIRY_SKEW_MS = 60_000;

// The shared `api` client attaches a bearer via getValidAccessToken(), so the
// refresh call must not go through it. This endpoint authenticates with the
// refresh cookie anyway.
const refreshClient = axios.create({ baseURL: "/api", withCredentials: true });

interface AccessToken {
  token: string;
  expiresAt: number;
  skewMs: number;
}

let currentToken: AccessToken | null = null;
let inFlightRefresh: Promise<AccessToken> | null = null;
let onSessionExpired: (() => void) | null = null;

export function setOnSessionExpired(handler: () => void) {
  onSessionExpired = handler;
}

async function requestAccessToken(): Promise<AccessToken> {
  const res = await refreshClient.post<{
    access_token: string;
    access_token_exp: string;
  }>("/auth/token/access");

  const expiresAt = new Date(res.data.access_token_exp).getTime();
  const lifetimeMs = expiresAt - Date.now();

  return {
    token: res.data.access_token,
    expiresAt,
    skewMs: Math.min(EXPIRY_SKEW_MS, Math.max(lifetimeMs / 2, 0)),
  };
}

// Callers that miss at the same moment — a page's worth of queries waking from
// sleep together — must collapse into a single request, not one each.
export function refreshAccessToken(): Promise<AccessToken> {
  if (!inFlightRefresh) {
    inFlightRefresh = requestAccessToken()
      .then((token) => {
        currentToken = token;
        return token;
      })
      .catch((error: unknown) => {
        currentToken = null;
        // A network blip must not sign anyone out — only a rejected refresh
        // cookie means the session is really gone.
        if (axios.isAxiosError(error) && error.response?.status === 401) {
          onSessionExpired?.();
        }
        throw error;
      })
      .finally(() => {
        inFlightRefresh = null;
      });
  }

  return inFlightRefresh;
}

export function getValidAccessToken(): Promise<AccessToken> {
  if (
    currentToken &&
    currentToken.expiresAt - currentToken.skewMs > Date.now()
  ) {
    return Promise.resolve(currentToken);
  }

  return refreshAccessToken();
}
