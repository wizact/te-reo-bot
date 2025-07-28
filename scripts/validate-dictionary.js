#!/usr/bin/env node
/**
 * Te Reo Māori Dictionary Validation Script
 *
 * This script validates the structure and content of dictionary.json
 * to ensure data integrity for the Te Reo Māori bot.
 * 
 * Validation Rules:
 * - Required fields: index (unique integer), word (string, can be duplicate), meaning (string, can be duplicate)
 * - Optional fields: link, photo, photo_attribution (all strings)
 */

const fs = require('fs');
const path = require('path');

// Colors for console output
const colors = {
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  reset: '\x1b[0m'
};

function log(color, message) {
  console.log(`${color}${message}${colors.reset}`);
}

function validateDictionaryStructure(data) {
  const errors = [];
  const warnings = [];

  // Check if root has 'dictionary' property
  if (!data.hasOwnProperty('dictionary')) {
    errors.push('Root object must have a "dictionary" property');
    return { errors, warnings };
  }

  // Check if dictionary is an array
  if (!Array.isArray(data.dictionary)) {
    errors.push('Dictionary property must be an array');
    return { errors, warnings };
  }

  if (data.dictionary.length === 0) {
    errors.push('Dictionary array cannot be empty');
    return { errors, warnings };
  }

  log(colors.blue, `Found ${data.dictionary.length} dictionary entries`);

  // Track duplicate indices (must be unique)
  const seenIndices = new Set();

  // Validate each entry
  data.dictionary.forEach((entry, index) => {
    const result = validateDictionaryEntry(entry, index, seenIndices);
    errors.push(...result.errors);
    warnings.push(...result.warnings);
  });

  return { errors, warnings };
}

function validateDictionaryEntry(entry, position, seenIndices) {
  const errors = [];
  const warnings = [];
  const entryPrefix = `Entry ${position + 1}`;

  // Check if entry is an object
  if (typeof entry !== 'object' || entry === null) {
    errors.push(`${entryPrefix}: Entry must be an object`);
    return { errors, warnings };
  }

  // Required fields validation
  const requiredFields = ['index', 'word', 'meaning'];
  for (const field of requiredFields) {
    if (!entry.hasOwnProperty(field)) {
      errors.push(`${entryPrefix}: Missing required field "${field}"`);
    } else if (entry[field] === null || entry[field] === undefined) {
      errors.push(`${entryPrefix}: Field "${field}" cannot be null or undefined`);
    }
  }

  // Type validation for required fields
  if (entry.hasOwnProperty('index')) {
    if (typeof entry.index !== 'number' || !Number.isInteger(entry.index)) {
      errors.push(`${entryPrefix}: Field "index" must be an integer (got: ${typeof entry.index})`);
    } else if (entry.index <= 0) {
      errors.push(`${entryPrefix}: Field "index" must be a positive integer (got: ${entry.index})`);
    } else {
      // Check for duplicate indices - these must be unique
      if (seenIndices.has(entry.index)) {
        errors.push(`${entryPrefix}: Duplicate index "${entry.index}" found - indices must be unique`);
      }
      seenIndices.add(entry.index);
    }
  }

  if (entry.hasOwnProperty('word')) {
    if (typeof entry.word !== 'string') {
      errors.push(`${entryPrefix}: Field "word" must be a string (got: ${typeof entry.word})`);
    } else if (entry.word.trim() === '') {
      errors.push(`${entryPrefix}: Field "word" cannot be empty`);
    }
    // Note: Words are allowed to be duplicates per requirements
  }

  if (entry.hasOwnProperty('meaning')) {
    if (typeof entry.meaning !== 'string') {
      errors.push(`${entryPrefix}: Field "meaning" must be a string (got: ${typeof entry.meaning})`);
    } else if (entry.meaning.trim() === '') {
      errors.push(`${entryPrefix}: Field "meaning" cannot be empty`);
    }
  }

  // Optional fields validation (these can be empty strings)
  const optionalFields = ['link', 'photo', 'photo_attribution'];
  for (const field of optionalFields) {
    if (entry.hasOwnProperty(field)) {
      if (typeof entry[field] !== 'string') {
        errors.push(`${entryPrefix}: Field "${field}" must be a string if present (got: ${typeof entry[field]})`);
      }
    }
  }

  // Check for unexpected fields
  const allowedFields = ['index', 'word', 'meaning', 'link', 'photo', 'photo_attribution'];
  const extraFields = Object.keys(entry).filter(key => !allowedFields.includes(key));
  if (extraFields.length > 0) {
    warnings.push(`${entryPrefix}: Unexpected fields found: ${extraFields.join(', ')}`);
  }

  return { errors, warnings };
}

function validateJSON(filePath) {
  try {
    const fileContent = fs.readFileSync(filePath, 'utf8');
    const data = JSON.parse(fileContent);
    return { success: true, data };
  } catch (error) {
    if (error instanceof SyntaxError) {
      return {
        success: false,
        error: `Invalid JSON syntax: ${error.message}`
      };
    }
    return {
      success: false,
      error: `Error reading file: ${error.message}`
    };
  }
}

function main() {
  const dictionaryPath = path.join(__dirname, '..', 'cmd', 'server', 'dictionary.json');

  log(colors.blue, '🔍 Validating Te Reo Māori Dictionary...');
  log(colors.blue, `📂 File: ${dictionaryPath}`);

  // Check if file exists
  if (!fs.existsSync(dictionaryPath)) {
    log(colors.red, `❌ Dictionary file not found: ${dictionaryPath}`);
    process.exit(1);
  }

  // Validate JSON syntax
  const jsonResult = validateJSON(dictionaryPath);
  if (!jsonResult.success) {
    log(colors.red, `❌ JSON Validation Failed:`);
    log(colors.red, jsonResult.error);
    process.exit(1);
  }

  log(colors.green, '✅ JSON syntax is valid');

  // Validate dictionary structure
  const structureResult = validateDictionaryStructure(jsonResult.data);

  if (structureResult.errors.length > 0) {
    log(colors.red, `❌ Dictionary Structure Validation Failed:`);
    structureResult.errors.forEach(error => {
      log(colors.red, `   • ${error}`);
    });

    log(colors.yellow, '\n💡 Common fixes:');
    log(colors.yellow, '   • Ensure all entries have index, word, and meaning fields');
    log(colors.yellow, '   • Check that indices are unique positive integers');
    log(colors.yellow, '   • Words and meanings can be duplicates');
    log(colors.yellow, '   • Confirm all required fields are non-empty strings (except index which is integer)');

    process.exit(1);
  }

  // Show warnings if any
  if (structureResult.warnings.length > 0) {
    log(colors.yellow, `⚠️  Found ${structureResult.warnings.length} warnings:`);
    // Only show first 10 warnings to avoid overwhelming output
    structureResult.warnings.slice(0, 10).forEach(warning => {
      log(colors.yellow, `   • ${warning}`);
    });
    if (structureResult.warnings.length > 10) {
      log(colors.yellow, `   ... and ${structureResult.warnings.length - 10} more warnings`);
    }
    log(colors.yellow, '\n💡 Note: These are warnings for existing data and do not block validation');
  }

  log(colors.green, '✅ Dictionary structure validation passed');
  log(colors.green, `🎉 Dictionary is valid! Found ${jsonResult.data.dictionary.length} entries.`);

  // Show some statistics
  const stats = generateStatistics(jsonResult.data.dictionary);
  log(colors.blue, '\n📊 Statistics:');
  Object.entries(stats).forEach(([key, value]) => {
    log(colors.blue, `   • ${key}: ${value}`);
  });
}

function generateStatistics(entries) {
  const stats = {
    'Total entries': entries.length,
    'Entries with photos': entries.filter(e => e.photo && e.photo.trim() !== '').length,
    'Entries with links': entries.filter(e => e.link && e.link.trim() !== '').length,
    'Entries with photo attribution': entries.filter(e => e.photo_attribution && e.photo_attribution.trim() !== '').length,
    'Highest index': Math.max(...entries.map(e => e.index)),
    'Average word length': Math.round(entries.reduce((sum, e) => sum + e.word.length, 0) / entries.length)
  };

  return stats;
}

// Run the validation
if (require.main === module) {
  main();
}

module.exports = {
  validateDictionaryStructure,
  validateDictionaryEntry,
  validateJSON
};
