export type ApiError = Error & { data: any };

export function isApiError(error: Error): error is ApiError {
  return "data" in error;
}
