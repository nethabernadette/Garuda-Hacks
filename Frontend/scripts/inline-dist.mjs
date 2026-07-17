import { readFile, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const dist = resolve(root, 'dist');
const indexPath = resolve(dist, 'index.html');

let html = await readFile(indexPath, 'utf8');

const cssMatch = html.match(/<link rel="stylesheet" crossorigin href="\.\/([^"]+)">/);
if (cssMatch) {
  const cssPath = resolve(dist, cssMatch[1]);
  const css = await readFile(cssPath, 'utf8');
  html = html.replace(cssMatch[0], `<style>\n${css}\n</style>`);
}

const jsMatch = html.match(/<script type="module" crossorigin src="\.\/([^"]+)"><\/script>/);
if (jsMatch) {
  const jsPath = resolve(dist, jsMatch[1]);
  const js = await readFile(jsPath, 'utf8');
  html = html.replace(jsMatch[0], `<script>\n${js}\n</script>`);
}

await writeFile(indexPath, html, 'utf8');
console.log('Created single-file dist/index.html that can be opened with file://');
