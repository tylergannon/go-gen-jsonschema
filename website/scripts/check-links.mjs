import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { extname, join, relative, sep } from 'node:path';

const dist = new URL('../dist/', import.meta.url).pathname;
const htmlFiles = walk(dist).filter((path) => path.endsWith('.html'));
const failures = [];

for (const sourceFile of htmlFiles) {
  const sourceHtml = readFileSync(sourceFile, 'utf8');
  const sourceRoute = routeFor(sourceFile);
  const hrefs = [...sourceHtml.matchAll(/\shref=(?:"([^"]+)"|'([^']+)')/g)]
    .map((match) => match[1] ?? match[2]);

  for (const href of hrefs) {
    if (!href || /^(?:[a-z]+:)?\/\//i.test(href) || /^(?:mailto|tel|data):/i.test(href)) continue;

    const target = new URL(href, `https://docs.invalid${sourceRoute}`);
    const targetFile = fileFor(target.pathname);
    if (!targetFile) {
      failures.push(`${sourceRoute} -> ${href} (missing target)`);
      continue;
    }

    if (target.hash && extname(targetFile) === '.html') {
      const id = decodeURIComponent(target.hash.slice(1));
      if (!id) continue;
      const targetHtml = readFileSync(targetFile, 'utf8');
      const escaped = escapeRegExp(id);
      if (!new RegExp(`(?:id|name)=["']${escaped}["']`).test(targetHtml)) {
        failures.push(`${sourceRoute} -> ${href} (missing fragment)`);
      }
    }
  }
}

if (failures.length > 0) {
  console.error(`Broken internal links:\n${failures.map((failure) => `- ${failure}`).join('\n')}`);
  process.exit(1);
}

console.log(`Checked ${htmlFiles.length} HTML pages; all internal links resolve.`);

function walk(directory) {
  return readdirSync(directory).flatMap((name) => {
    const path = join(directory, name);
    return statSync(path).isDirectory() ? walk(path) : [path];
  });
}

function routeFor(file) {
  const path = `/${relative(dist, file).split(sep).join('/')}`;
  return path.endsWith('/index.html') ? path.slice(0, -'index.html'.length) : path;
}

function fileFor(pathname) {
  const path = decodeURIComponent(pathname).replace(/^\/+/, '');
  const exact = join(dist, path);
  const candidates = pathname.endsWith('/')
    ? [join(exact, 'index.html')]
    : [exact, `${exact}.html`, join(exact, 'index.html')];
  return candidates.find((candidate) => existsSync(candidate) && !statSync(candidate).isDirectory());
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
