import { createSignal, Show } from "solid-js";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { Button, TextArea, TextFieldRoot, showToast } from "@glassact/ui";
import type { GET, InlayProof } from "@glassact/data";
import { postApproveProofOpts, postDeclineProofOpts } from "../../queries/proof";
import { isApiError } from "../../utils/is-api-error";
import type { Component } from "solid-js";

interface ProofActionsProps {
  proof: GET<InlayProof>;
  inlayUuid: string;
}

const ProofActions: Component<ProofActionsProps> = (props) => {
  const queryClient = useQueryClient();
  const approveMutation = useMutation(() => postApproveProofOpts());
  const declineMutation = useMutation(() => postDeclineProofOpts());

  const [showDeclineForm, setShowDeclineForm] = createSignal(false);
  const [declineReason, setDeclineReason] = createSignal("");

  const invalidateQueries = () => {
    queryClient.invalidateQueries({ queryKey: ["inlay", props.inlayUuid, "proofs"] });
    queryClient.invalidateQueries({ queryKey: ["inlay", props.inlayUuid, "chats"] });
    queryClient.invalidateQueries({ queryKey: ["inlay", props.inlayUuid] });
    queryClient.invalidateQueries({ queryKey: ["project"] });
  };

  const handleApprove = () => {
    approveMutation.mutate(props.proof.uuid, {
      onSuccess() {
        showToast({ title: "Proof approved", variant: "success" });
        invalidateQueries();
      },
      onError(error) {
        showToast({
          title: "Failed to approve",
          description: isApiError(error) ? error.data?.error : "Unknown error",
          variant: "error",
        });
      },
    });
  };

  const handleDecline = () => {
    const reason = declineReason().trim();
    if (!reason) {
      showToast({ title: "Please provide a reason", variant: "error" });
      return;
    }

    declineMutation.mutate(
      { proofUuid: props.proof.uuid, body: { decline_reason: reason } },
      {
        onSuccess() {
          showToast({ title: "Proof declined", variant: "success" });
          setShowDeclineForm(false);
          setDeclineReason("");
          invalidateQueries();
        },
        onError(error) {
          showToast({
            title: "Failed to decline",
            description: isApiError(error) ? error.data?.error : "Unknown error",
            variant: "error",
          });
        },
      },
    );
  };

  return (
    <Show when={props.proof.status === "pending"}>
      <div class="flex flex-col gap-2 mt-2">
        <Show
          when={!showDeclineForm()}
          fallback={
            <div class="flex flex-col gap-2">
              <TextFieldRoot>
                <TextArea
                  value={declineReason()}
                  onInput={(e: InputEvent & { currentTarget: HTMLTextAreaElement }) =>
                    setDeclineReason(e.currentTarget.value)
                  }
                  placeholder="Reason for declining..."
                />
              </TextFieldRoot>
              <div class="flex gap-2">
                <Button
                  size="sm"
                  variant="destructive"
                  disabled={declineMutation.isPending || !declineReason().trim()}
                  onClick={handleDecline}
                >
                  {declineMutation.isPending ? "Declining..." : "Confirm Decline"}
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    setShowDeclineForm(false);
                    setDeclineReason("");
                  }}
                >
                  Cancel
                </Button>
              </div>
            </div>
          }
        >
          <div class="flex gap-2">
            <Button
              size="sm"
              variant="default"
              disabled={approveMutation.isPending}
              onClick={handleApprove}
            >
              {approveMutation.isPending ? "Approving..." : "Approve"}
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setShowDeclineForm(true)}
            >
              Decline
            </Button>
          </div>
        </Show>
      </div>
    </Show>
  );
};

export default ProofActions;
