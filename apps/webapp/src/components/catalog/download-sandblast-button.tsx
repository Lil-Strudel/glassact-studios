import { getBackgroundSandblastUrl } from "../../queries/background";
import { PresignedDownloadButton } from "../presigned-download-button";

interface DownloadSandblastButtonProps {
  backgroundUuid: string;
  class?: string;
}

export function DownloadSandblastButton(props: DownloadSandblastButtonProps) {
  return (
    <PresignedDownloadButton
      fetchUrl={() => getBackgroundSandblastUrl(props.backgroundUuid)}
      label="Download sandblast file"
      errorTitle="Failed to download sandblast file"
      class={props.class}
    />
  );
}
