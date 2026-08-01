import type { PaymentTiming } from "@glassact/data";

export const PAYMENT_TIMING_LABELS: Record<PaymentTiming, string> = {
  "pre-manufacturing": "Before manufacturing",
  "pre-shipping": "Before shipping",
  "post-shipping": "After shipping",
};

export interface PaymentTimingNotice {
  tone: "info" | "warning";
  message: string;
}

// `info` states the rule that applies to the project; `warning` is the same
// rule when there is actually an unpaid invoice sitting in its way. Neither
// blocks anything — the whole feature is informational.
const NOTICES: Record<
  Exclude<PaymentTiming, "post-shipping">,
  { info: (owner: string) => string; warning: string }
> = {
  "pre-manufacturing": {
    info: (owner) =>
      `${owner} pays before manufacturing. Production starts once the invoice is paid.`,
    warning: "Waiting on payment before manufacturing can begin.",
  },
  "pre-shipping": {
    info: (owner) =>
      `${owner} pays before shipping. Orders ship once the invoice is paid.`,
    warning:
      "This project is waiting to ship until the invoice is paid. Once payment is received it will be released for shipment.",
  },
};

export function paymentTimingNotice(args: {
  timing: PaymentTiming | undefined;
  awaitingPayment: boolean;
  isInternal: boolean;
}): PaymentTimingNotice | null {
  if (!args.timing || args.timing === "post-shipping") return null;

  const notice = NOTICES[args.timing];
  if (args.awaitingPayment) {
    return { tone: "warning", message: notice.warning };
  }

  return {
    tone: "info",
    message: notice.info(args.isInternal ? "This dealership" : "Your account"),
  };
}
