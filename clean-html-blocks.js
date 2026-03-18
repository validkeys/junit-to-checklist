#!/usr/bin/env node

/**
 * @fileoverview Standalone utility to clean HTML blocks from markdown files
 *
 * This is a utility script for post-processing markdown files that contain
 * verbose HTML DOM dumps with ANSI color codes. It's provided as a separate
 * tool for cases where you need to clean existing files that were generated
 * without the built-in cleaning functionality.
 *
 * Note: The main parse-failures.js script now includes this cleaning automatically,
 * so this utility is primarily for cleaning legacy or externally-generated files.
 *
 * @author junit-to-checklist contributors
 * @license MIT
 *
 * @example
 * // Clean the default failing-tests.md file
 * node clean-html-blocks.js
 *
 * @example
 * // Clean a custom file
 * node clean-html-blocks.js ./path/to/file.md
 */

const fs = require('fs');
const path = require('path');

const filePath = process.argv[2] || './failing-tests.md';

console.log(`Reading ${filePath}...`);
const content = fs.readFileSync(filePath, 'utf8');

// Split into lines
const lines = content.split('\n');
const cleaned = [];
let inHtmlBlock = false;
let skippedBlocks = 0;

for (let i = 0; i < lines.length; i++) {
  const line = lines[i];

  // Check if this line starts an HTML block (contains ANSI codes and <html>)
  if (line.includes('[36m<html>[39m') || line.includes('[36m<html>')) {
    inHtmlBlock = true;
    skippedBlocks++;
    continue;
  }

  // Check if this line ends an HTML block
  if (inHtmlBlock && (line.includes('[36m</html>[39m') || line.includes('</html>[39m'))) {
    inHtmlBlock = false;
    continue;
  }

  // Skip lines that are part of an HTML block
  if (inHtmlBlock) {
    continue;
  }

  // Also skip lines that contain "Ignored nodes: comments, script, style" which precede HTML blocks
  if (line.includes('Ignored nodes: comments, script, style')) {
    continue;
  }

  // Keep all other lines
  cleaned.push(line);
}

// Write back to file
const outputPath = filePath;
fs.writeFileSync(outputPath, cleaned.join('\n'), 'utf8');

console.log(`✓ Cleaned ${skippedBlocks} HTML blocks`);
console.log(`✓ Original lines: ${lines.length}`);
console.log(`✓ Cleaned lines: ${cleaned.length}`);
console.log(`✓ Removed ${lines.length - cleaned.length} lines`);
console.log(`✓ Saved to ${outputPath}`);
