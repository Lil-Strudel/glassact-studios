import { Tooltip, TooltipContent, TooltipTrigger } from "@glassact/ui";

export default function FooterContent() {
  return (
    <Tooltip>
      <TooltipTrigger> GlassAct Studios Inc. © 2014 - 2025 </TooltipTrigger>
      <TooltipContent>
        <p>All Images and Text © 2014 - 2025</p>
        <p>GlassAct Studios Inc.</p>
        <p>All Rights Reserved</p>
        <p>US Patent 9,470,009 B2 and other patents pending</p>
      </TooltipContent>
    </Tooltip>
  );
}
