import type { StoneTextElement } from "@glassact/data";
import { newElementId } from "./stone-geometry";

interface NewTextOptions {
  text: string;
  x: number;
  y: number;
  size: number;
  font: StoneTextElement["font"];
}

export function newTextElement(options: NewTextOptions): StoneTextElement {
  return {
    id: newElementId(),
    type: "text",
    text: options.text,
    font: options.font,
    size: options.size,
    align: "center",
    finish: "engraved",
    x: options.x,
    y: options.y,
    rotation: 0,
    group_id: null,
  };
}
