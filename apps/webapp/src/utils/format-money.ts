// Nothing in the platform prices tax or shipping yet, so every amount a
// dealership sees is the product price alone. One shared string keeps that
// caveat worded identically wherever a total is shown.
export const PRICE_CAVEAT = "Before tax and shipping.";

export function formatMoney(
  amount: number,
  options: {
    currency?: string;
    useParenthesesForNegative?: boolean;
  } = {},
): string {
  if (!isFinite(amount)) {
    return "Invalid amount";
  }

  const { currency = "$", useParenthesesForNegative = false } = options;

  const formatter = new Intl.NumberFormat("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });

  if (amount < 0 && useParenthesesForNegative) {
    return `(${currency}${formatter.format(Math.abs(amount))})`;
  }

  return `${currency}${formatter.format(amount)}`;
}
