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
