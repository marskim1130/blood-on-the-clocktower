import { readFileSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath, URL as NodeURL } from 'node:url';
import { expect, it } from 'vitest';

it('keeps mobile confirmation button labels within the four-character platform limit', () => {
  const failures: string[] = [];
  function check(directory: string): void {
    for (const entry of readdirSync(directory, { withFileTypes: true })) {
      const path = join(directory, entry.name);
      if (entry.isDirectory()) check(path);
      else if (entry.name.endsWith('.tsx')) {
        for (const match of readFileSync(path, 'utf8').matchAll(/confirmText:\s*([^,\n]+)/g)) {
          for (const label of match[1]!.matchAll(/'([^']*)'/g)) {
            if ([...label[1]!].length > 4) failures.push(`${entry.name}: ${label[1]}`);
          }
        }
      }
    }
  }
  check(fileURLToPath(new NodeURL('../', import.meta.url)));
  expect(failures).toEqual([]);
});
