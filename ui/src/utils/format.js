/**
 * Pure formatting helpers for the stats dashboard.
 * No DOM/Vue dependencies - safe to unit test in isolation.
 */

/**
 * Format an integer as a locale-grouped string, e.g. 1284 -> "1,284".
 * Values at or above 100,000 compact to a single decimal + suffix,
 * e.g. 128400 -> "128.4K", 12900000 -> "12.9M", 1284000000 -> "1.3B".
 *
 * @param {number} n
 * @returns {string}
 */
const COMPACT_UNITS = ['', 'K', 'M', 'B'];

export function compactNumber(n) {
  const value = Number(n) || 0;
  const abs = Math.abs(value);

  if (abs < 100_000) {
    return Math.round(value).toLocaleString('en-US');
  }

  let exponent = Math.min(Math.floor(Math.log10(abs) / 3), COMPACT_UNITS.length - 1);
  let scaled = trimDecimal(value / Math.pow(1000, exponent));

  // A value like 999,999 rounds to "1000K" at one decimal - bump to the next unit.
  if (Math.abs(parseFloat(scaled)) >= 1000 && exponent < COMPACT_UNITS.length - 1) {
    exponent += 1;
    scaled = trimDecimal(value / Math.pow(1000, exponent));
  }

  return `${scaled}${COMPACT_UNITS[exponent]}`;
}

/**
 * Format a byte count into human-readable units (base 1024).
 * e.g. 0 -> "0 B", 5153960755 -> "4.8 GB"
 *
 * @param {number} n
 * @returns {string}
 */
export function formatBytes(n) {
  const value = Number(n) || 0;
  if (value <= 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const exponent = Math.min(
    Math.floor(Math.log(value) / Math.log(1024)),
    units.length - 1
  );
  const scaled = value / Math.pow(1024, exponent);
  const decimals = exponent === 0 ? 0 : 1;

  return `${scaled.toFixed(decimals)} ${units[exponent]}`;
}

/**
 * Round to one decimal place, dropping a trailing ".0",
 * e.g. 12.94 -> "12.9", 250.0 -> "250".
 */
function trimDecimal(n) {
  const rounded = Math.round(n * 10) / 10;
  return Number.isInteger(rounded) ? String(rounded) : rounded.toFixed(1);
}
