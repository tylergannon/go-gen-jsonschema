import type { Shipment } from './generated/ts/index.js';

function claim(condition: boolean, message: string): asserts condition {
  if (!condition) throw new Error(message);
}

export function inspectGoAndAuthorNode(text: string): string {
  const shipment = JSON.parse(text) as Shipment;
  claim(shipment.id === 'shipment-42', 'Node did not observe Go shipment id');
  claim(shipment.status === 'StatusReady', 'Node did not observe registered enum name');
  claim(shipment.event.kind === 'created', 'Node did not observe Go union discriminator');
  claim(shipment.event.actor === 'go-service', 'Node did not observe Go union payload');
  claim(shipment.event.at === '2026-09-04T18:30:00Z', 'Node did not observe Go timestamp string');
  claim(shipment.note === 'packed', 'Node did not observe present Optional value');
  claim(shipment.eta === '2026-09-05T14:00:00Z', 'Node did not observe present Nullable value');
  claim(shipment.quantity === 3, 'Node did not observe Go quantity');

  const nodeAuthored: Shipment = {
    ...shipment,
    status: 'StatusDelivered',
    event: { kind: 'dispatched', carrier: 'Parcel Post', tracking: 'TRACK-42' },
    note: '',
    eta: null,
    quantity: 4,
  };
  return JSON.stringify(nodeAuthored, null, 2) + '\n';
}

export function inspectGoReencode(text: string): void {
  const shipment = JSON.parse(text) as Shipment;
  claim(shipment.id === 'shipment-42', 'Go re-encode changed shipment id');
  claim(shipment.status === 'StatusDelivered', 'Go re-encode changed enum name');
  claim(shipment.event.kind === 'dispatched', 'Go re-encode changed union discriminator');
  claim(shipment.event.carrier === 'Parcel Post', 'Go re-encode changed union payload');
  claim(shipment.event.tracking === 'TRACK-42', 'Go re-encode changed tracking value');
  claim(shipment.note === 'accepted-by-go', 'Node did not observe Go Optional update');
  claim(shipment.eta === '2026-09-05T16:30:00Z', 'Node did not observe Go Nullable update');
  claim(shipment.quantity === 5, 'Node did not observe Go quantity update');
}
