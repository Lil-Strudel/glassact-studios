function validator(num: any) {
  return !isNaN(Number(num));
}
export const zodStringNumber = [
  validator,
  { message: "Must be number" },
] as const;
