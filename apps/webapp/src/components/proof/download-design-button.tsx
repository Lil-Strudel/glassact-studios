import { getProofDesignDownloadUrl } from "../../queries/proof";
import { PresignedDownloadButton } from "../presigned-download-button";

interface DownloadDesignButtonProps {
  proofUuid: string;
}

export function DownloadDesignButton(props: DownloadDesignButtonProps) {
  return (
    <PresignedDownloadButton
      fetchUrl={() => getProofDesignDownloadUrl(props.proofUuid)}
      label="Download Design"
      errorTitle="Failed to download design"
    />
  );
}
