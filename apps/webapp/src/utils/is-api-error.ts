export type ApiError = Error & { data: { error?: string; message?: string } };

export function isApiError(error: Error): error is ApiError {
  return "data" in error;
}
