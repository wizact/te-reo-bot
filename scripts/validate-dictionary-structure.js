#!/usr/bin/env node
/**
 * Basic Dictionary Structure Validation Script
 * 
 * This script performs basic JSON structure validation for dictionary.json
 * to ensure it has the required root structure before running more detailed validation.
 */

const fs = require('fs');

function validateDictionaryStructure() {
  const dictionaryPath = 'cmd/server/dictionary.json';
  
  try {
    const data = JSON.parse(fs.readFileSync(dictionaryPath, 'utf8'));
    
    if (!data.dictionary || !Array.isArray(data.dictionary)) {
      console.error('❌ Dictionary must have a dictionary array at root');
      process.exit(1);
    }
    
    console.log('✅ Dictionary JSON structure is valid');
  } catch (error) {
    console.error('❌ Dictionary JSON is invalid:', error.message);
    process.exit(1);
  }
}

// Run validation if this script is executed directly
if (require.main === module) {
  validateDictionaryStructure();
}

module.exports = { validateDictionaryStructure };