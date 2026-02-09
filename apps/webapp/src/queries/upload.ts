import api from "./api";

export interface UploadResponse {
  url: string;
  filename: string;
  size: number;
  content_type: string;
  key: string;
  uploaded_at: string;
}

export async function postUpload(
  file: File,
  uploadPath: string
): Promise<UploadResponse> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("uploadPath", uploadPath);

  const res = await api.post("/upload", formData);

  return res.data;
}
