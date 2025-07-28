#!/usr/bin/env node
/**
 * Test script to verify dictionary validation
 * This creates temporary test files to ensure validation works correctly
 */

const fs = require('fs');
const path = require('path');
const { validateDictionaryStructure, validateJSON } = require('./validate-dictionary');

// Create test directory
const testDir = '/tmp/dictionary-tests';
if (!fs.existsSync(testDir)) {
  fs.mkdirSync(testDir, { recursive: true });
}

function runTest(name, testData, expectSuccess = true) {
  console.log(`\n🧪 Running test: ${name}`);

  const testFile = path.join(testDir, `${name.replace(/\s+/g, '-').toLowerCase()}.json`);
  fs.writeFileSync(testFile, JSON.stringify(testData, null, 2));

  const jsonResult = validateJSON(testFile);
  if (!jsonResult.success) {
    console.log(`❌ JSON parsing failed: ${jsonResult.error}`);
    return false;
  }

  const structureResult = validateDictionaryStructure(jsonResult.data);
  const hasErrors = structureResult.errors.length > 0;

  if (expectSuccess && hasErrors) {
    console.log(`❌ Expected success but got errors:`);
    structureResult.errors.forEach(err => console.log(`   • ${err}`));
    return false;
  }

  if (!expectSuccess && !hasErrors) {
    console.log(`❌ Expected failure but validation passed`);
    return false;
  }

  if (expectSuccess) {
    console.log(`✅ Test passed - validation successful`);
    if (structureResult.warnings.length > 0) {
      console.log(`   ℹ️  ${structureResult.warnings.length} warnings (expected)`);
    }
  } else {
    console.log(`✅ Test passed - validation correctly failed`);
  }

  return true;
}

// Test cases
const tests = [
  // Valid structure
  {
    name: "Valid Dictionary",
    data: {
      dictionary: [
        {
          index: 1,
          word: "Kia ora",
          meaning: "Hello, greetings",
          link: "",
          photo: "hello.jpg",
          photo_attribution: "A greeting photo"
        }
      ]
    },
    expectSuccess: true
  },

  // Missing dictionary key
  {
    name: "Missing Dictionary Key",
    data: {
      words: [
        { index: 1, word: "Test", meaning: "Test word" }
      ]
    },
    expectSuccess: false
  },

  // Empty dictionary
  {
    name: "Empty Dictionary",
    data: {
      dictionary: []
    },
    expectSuccess: false
  },

  // Missing required fields
  {
    name: "Missing Required Fields",
    data: {
      dictionary: [
        { word: "Test", meaning: "Test word" } // Missing index
      ]
    },
    expectSuccess: false
  },

  // Invalid data types
  {
    name: "Invalid Data Types",
    data: {
      dictionary: [
        {
          index: "not-a-number",
          word: "Test",
          meaning: "Test word"
        }
      ]
    },
    expectSuccess: false
  },

  // Optional fields can be empty
  {
    name: "Empty Optional Fields",
    data: {
      dictionary: [
        {
          index: 1,
          word: "Test",
          meaning: "Test word",
          link: "",
          photo: "",
          photo_attribution: ""
        }
      ]
    },
    expectSuccess: true
  },

  // Duplicate words should be allowed
  {
    name: "Duplicate Words Allowed",
    data: {
      dictionary: [
        {
          index: 1,
          word: "Kia ora",
          meaning: "Hello, greetings"
        },
        {
          index: 2,
          word: "Kia ora", 
          meaning: "Farewell, goodbye"
        }
      ]
    },
    expectSuccess: true
  },

  // Duplicate indices should fail
  {
    name: "Duplicate Indices Not Allowed",
    data: {
      dictionary: [
        {
          index: 1,
          word: "Hello",
          meaning: "Greeting"
        },
        {
          index: 1,
          word: "Goodbye", 
          meaning: "Farewell"
        }
      ]
    },
    expectSuccess: false
  }
];

console.log('🧪 Running Dictionary Validation Tests...');

let passedTests = 0;
let totalTests = tests.length;

tests.forEach(test => {
  if (runTest(test.name, test.data, test.expectSuccess)) {
    passedTests++;
  }
});

console.log(`\n📊 Test Results: ${passedTests}/${totalTests} tests passed`);

if (passedTests === totalTests) {
  console.log('🎉 All tests passed!');
  process.exit(0);
} else {
  console.log('❌ Some tests failed');
  process.exit(1);
}

// Clean up
fs.rmSync(testDir, { recursive: true, force: true });
