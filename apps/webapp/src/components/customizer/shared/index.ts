// Shared customizer primitives consumed by both the admin manifest editor and
// the consumer customizer. Import from this barrel:
//
//   import {
//     CustomizerCanvas,
//     SwatchPicker,
//     buildPieceSourceMap,
//     resolvePieceHex,
//   } from "../customizer/shared";

export { CustomizerCanvas } from "./customizer-canvas";
export type { CustomizerCanvasProps } from "./customizer-canvas";

export { SwatchPicker } from "./swatch-picker";
export type { Swatch } from "./swatch-picker";

export {
  buildPieceSourceMap,
  buildGroutPieceIds,
  resolvePieceHex,
  groupGlassId,
  customPieceCount,
  totalCustomPieces,
} from "./resolution";
export type { Selection, GlassById } from "./resolution";
