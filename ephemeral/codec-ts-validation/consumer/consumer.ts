import type { Envelope } from './generated/ts/index.js';

const valid: Envelope = {
  primary: { kind: 'created', name: 'first' },
  alternate: { '!kind': 'alt-created', name: 'alternate' },
  maybe: { kind: 'deleted', id: 'optional' },
  events: [
    { kind: 'created', name: 'slice-value' },
    { kind: 'deleted', id: 'slice-pointer' },
  ],
  string_state: 'StateDone',
  numeric_state: 2,
  optional_state: 'StateNew',
  nullable_state: 'StateDone',
  null_state: null,
  'ordinary-name': 'ordinary',
  meta: { visible: 'visible' },
};

const { maybe: _maybe, optional_state: _optionalState, ...withoutOptionalFields } = valid;
const optionalFieldsMayBeAbsent: Envelope = withoutOptionalFields;

// @ts-expect-error string-mode enums use registered Go constant names, not String() output
const wrongStringEnum: Envelope = { ...valid, string_state: 'human-state-2' };
// @ts-expect-error numeric mode retains the exact registered numeric enum values
const wrongNumericEnum: Envelope = { ...valid, numeric_state: 99 };
// @ts-expect-error Optional enum fields do not accept null
const nullOptionalEnum: Envelope = { ...valid, optional_state: null };
// @ts-expect-error each field keeps its own discriminator property and values
const wrongAlternateDiscriminator: Envelope = { ...valid, alternate: { kind: 'created', name: 'wrong' } };

void [valid, optionalFieldsMayBeAbsent, wrongStringEnum, wrongNumericEnum, nullOptionalEnum, wrongAlternateDiscriminator];
