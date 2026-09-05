import { readFileSync, writeFileSync } from 'node:fs';
import { inspectGoAndAuthorNode, inspectGoReencode } from './dist/transport.js';

const [mode, input, output] = process.argv.slice(2);

if (mode === 'from-go' && input && output) {
  writeFileSync(output, inspectGoAndAuthorNode(readFileSync(input, 'utf8')));
  console.log('node_from_go_ok status=StatusReady kind=created optional=packed nullable=timestamp authored_status=StatusDelivered authored_kind=dispatched');
} else if (mode === 'verify-back' && input && !output) {
  inspectGoReencode(readFileSync(input, 'utf8'));
  console.log('node_verify_back_ok status=StatusDelivered kind=dispatched note=accepted-by-go eta=timestamp quantity=5');
} else {
  console.error('usage: runtime.mjs from-go INPUT OUTPUT | runtime.mjs verify-back INPUT');
  process.exit(2);
}
