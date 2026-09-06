import type { Envelope } from "./generated/index.js";
function assertNever(value: never): never { throw new Error(String(value)); }
export function incomplete(event: Envelope["event"]): string {
  switch (event["!kind"]) {
    case "created": return event.name;
    default: return assertNever(event);
  }
}
