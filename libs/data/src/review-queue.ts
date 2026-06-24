import type { GET } from "./helpers";
import type { Inlay, InlayCatalogInfo, InlayCustomInfo } from "./inlays";
import type { InlayProof } from "./inlay-proofs";

// The inlay as returned in the review queue: the base inlay plus its catalog or
// custom subtype info (no pricing/readiness enrichment — the queue acts on the
// pending proof directly).
export type ReviewQueueInlay = GET<Inlay> & {
  catalog_info?: GET<InlayCatalogInfo> | null;
  custom_info?: GET<InlayCustomInfo> | null;
};

export type ReviewQueueItem = {
  project_uuid: string;
  project_name: string;
  inlay: ReviewQueueInlay;
  pending_proof?: GET<InlayProof>;
};

export type ReviewQueue = {
  needs_approval: ReviewQueueItem[];
  needs_proof: ReviewQueueItem[];
};
