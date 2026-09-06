import type { Composition, Detail, Empty, Envelope, Status } from "./generated/index.js";

type Equal<A, B> = (<T>() => T extends A ? 1 : 2) extends
  (<T>() => T extends B ? 1 : 2) ? true : false;
type Assert<T extends true> = T;
type IsAny<T> = 0 extends (1 & T) ? true : false;
type IsUnknown<T> = unknown extends T ? true : false;
export type Proof = [
  Assert<Equal<IsAny<Envelope["event"]>, false>>,
  Assert<Equal<IsUnknown<Envelope["event"]>, false>>,
  Assert<Equal<Envelope["label"], string | undefined>>,
  Assert<Equal<Envelope["detail"], Detail | null>>,
  Assert<Equal<Envelope["status"], "ready" | 'wait"ing' | "converted">>,
  Assert<Equal<Envelope["priority"], 0 | 1 | 8 | 4>>,
  Assert<Equal<Envelope["priority_name"], "Low" | "High" | "Urgent" | "Medium">>,
  Assert<Equal<Envelope["when"], string>>,
  Assert<Equal<Composition["a"], number[][]>>,
  Assert<Equal<Composition["b"], Detail[]>>,
  Assert<Equal<Composition["c"], Status[] | undefined>>,
  Assert<Equal<Composition["d"], Status | null>>,
  Assert<Equal<Composition["e"], string[]>>,
  Assert<Equal<Composition["f"], boolean>>,
  Assert<Equal<Composition["g"], Envelope[]>>,
  Assert<Equal<Composition["h"]["value"], string>>,
  Assert<Equal<Composition["i"]["value"], number>>,
];

const created: Envelope["event"] = { "!kind": "Created", name: "Ada" };
const deleted: Envelope["event"] = { "!kind": "Deleted", id: 1 };
const envelope: Envelope = {
  event: created, other: deleted,
  events: [created, deleted], detail: null, shared: { message: "ok" },
  status: "ready", priority: 0, priority_name: "Low", when: "2026-09-04T00:00:00Z",
  "strange-key": "ok", empty: {},
};
export { envelope };

function assertNever(value: never): never { throw new Error(String(value)); }
export function exhaustive(event: Envelope["event"]): string {
  switch (event["!kind"]) {
    case "Created": return event.name;
    case "Deleted": return String(event.id);
    default: return assertNever(event);
  }
}

// @ts-expect-error a registered branch requires its discriminator
const missingTag: Envelope["event"] = { name: "Ada" };
// @ts-expect-error an unregistered discriminator is not admitted
const wrongTag: Envelope["event"] = { "!kind": "wrong", name: "Ada" };
// @ts-expect-error branches preserve payload requirements
const wrongPayload: Envelope["event"] = { "!kind": "Created", id: 1 };
// @ts-expect-error Optional excludes explicit null
const optionalNull: Envelope = { ...envelope, label: null };
// @ts-expect-error exactOptionalPropertyTypes excludes present undefined
const optionalUndefined: Envelope = { ...envelope, label: undefined };
const { detail: _detail, ...withoutDetail } = envelope;
// @ts-expect-error Nullable is a required property
const missingNullable: Envelope = withoutDetail;
// @ts-expect-error enum membership remains literal
const wrongEnum: Envelope["status"] = "unknown";
// @ts-expect-error an unrelated untyped constant is not a member of Status
const untypedEnum: Envelope["status"] = "not-a-status";
// @ts-expect-error String() output is not the string-mode wire value
const wrongName: Envelope["priority_name"] = "not-the-wire-value";
// @ts-expect-error pointer implementation does not imply null
const nullBranch: Envelope["event"] = null;
// @ts-expect-error empty Go objects must not admit primitives
const emptyNumber: Empty = 1;
// @ts-expect-error empty Go objects must not admit null
const emptyNull: Empty = null;
void [missingTag, wrongTag, wrongPayload, optionalNull, optionalUndefined,
  missingNullable, wrongEnum, untypedEnum, wrongName, nullBranch, emptyNumber, emptyNull];
