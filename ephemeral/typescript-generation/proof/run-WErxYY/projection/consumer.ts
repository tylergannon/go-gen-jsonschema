import type { Empty, Exact, Owner, Payload } from './generated/edge/index.js';

const exact: Exact = 9007199254740992;
// @ts-expect-error the adjacent integer is not the generated exact member
const wrongExact: Exact = 9007199254740991;

const payloadWithoutTag: Payload = { value: 'shared' };
const payloadWithOtherTag: Payload = { 'kind"\\\n雪': 'other', value: 'shared' };

const event: Owner['event'] = { 'kind"\\\n雪': '', value: 'variant' };
// @ts-expect-error a field-local variant requires its singleton tag
const missingTag: Owner['event'] = { value: 'variant' };
// @ts-expect-error the field-local variant rejects another tag
const wrongTag: Owner['event'] = { 'kind"\\\n雪': 'other', value: 'variant' };

const empty: Empty = {};
// @ts-expect-error empty objects exclude strings
const stringIsNotEmpty: Empty = 'primitive';
// @ts-expect-error empty objects exclude numbers
const numberIsNotEmpty: Empty = 1;
// @ts-expect-error empty objects exclude booleans
const booleanIsNotEmpty: Empty = false;
// @ts-expect-error empty objects exclude null
const nullIsNotEmpty: Empty = null;

void [
  exact,
  wrongExact,
  payloadWithoutTag,
  payloadWithOtherTag,
  event,
  missingTag,
  wrongTag,
  empty,
  stringIsNotEmpty,
  numberIsNotEmpty,
  booleanIsNotEmpty,
  nullIsNotEmpty,
];
