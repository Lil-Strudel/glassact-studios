const PHONE_DIGIT_COUNT = 10;

// Reduces anything a user might type or paste to the digits we store. A leading
// US country code is dropped rather than truncating the subscriber number off
// the end, so pasting "+1 (555) 123-4567" keeps the right ten digits.
export function toPhoneDigits(input: string): string {
  let digits = input.replace(/\D/g, "");

  if (digits.length === PHONE_DIGIT_COUNT + 1 && digits.startsWith("1")) {
    digits = digits.slice(1);
  }

  return digits.slice(0, PHONE_DIGIT_COUNT);
}

// Formats stored digits for display. Partial input formats as far as it can so
// the field prettifies while being typed into. No punctuation is added that a
// further digit would not justify — otherwise backspacing past it would
// re-add it and trap the cursor.
export function formatPhone(digits: string): string {
  const area = digits.slice(0, 3);
  const prefix = digits.slice(3, 6);
  const line = digits.slice(6, PHONE_DIGIT_COUNT);

  if (line) return `(${area}) ${prefix}-${line}`;
  if (prefix) return `(${area}) ${prefix}`;
  return area;
}
