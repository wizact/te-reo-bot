# Contributing

When contributing to this repository, please first discuss the change you wish to make via issue,
email, or any other method with the owners of this repository before making a change. 

## Development Setup

### Prerequisites

- Go 1.13 or later
- Node.js 14.0 or later
- Git

### Installing Pre-commit Hooks

This repository uses pre-commit hooks to ensure code quality and data integrity. To set up the development environment:

1. **Install pre-commit** (if not already installed):
   ```bash
   # Using pip
   pip install pre-commit
   
   # Or using homebrew (macOS)
   brew install pre-commit
   
   # Or using conda
   conda install -c conda-forge pre-commit
   ```

2. **Install Node.js dependencies**:
   ```bash
   npm install
   ```

3. **Install the git hook scripts**:
   ```bash
   pre-commit install
   ```

4. **Test the setup** (optional):
   ```bash
   pre-commit run --all-files
   ```

### Dictionary Validation

The dictionary.json file has automated validation to ensure data integrity:

#### Manual Validation

Run the validation script manually:
```bash
# Validate the dictionary structure and content
npm run validate-dictionary

# Or run directly
node scripts/validate-dictionary.js
```

#### Validation Rules

The dictionary validation checks for:

- **Structure**: Must have a "dictionary" array at the root level
- **Required fields**: Each entry must have `index`, `word`, and `meaning`
- **Data types**: 
  - `index` must be a unique positive integer
  - `word` and `meaning` must be non-empty strings (duplicates allowed)
- **Optional fields**: `link`, `photo`, and `photo_attribution` can be empty strings
- **Uniqueness**: Index values must be unique across all entries

#### Troubleshooting

**Common Issues:**

1. **"Dictionary file not found"**
   - Ensure you're running the command from the repository root
   - Check that `cmd/server/dictionary.json` exists

2. **"Invalid JSON syntax"**
   - Use a JSON validator or formatter to fix syntax errors
   - Run `npm run format-json` to auto-format the file

3. **"Missing required field"**
   - Ensure all entries have `index`, `word`, and `meaning` fields
   - Check for null or undefined values

4. **"Field must be a string/number"**
   - Verify data types match the requirements
   - Numbers should not be quoted, strings should be quoted

5. **"Duplicate index found"**
   - Each entry must have a unique index value
   - Find the highest existing index and use the next sequential number
   - Words and meanings can be duplicates

6. **Pre-commit hooks failing**
   - Run `npm run validate-dictionary` to see detailed errors
   - Fix validation issues before committing
   - Use `git commit --no-verify` only in emergencies (not recommended)

**Getting Help:**
- Check the validation output for specific error messages
- Run `node scripts/test-validation.js` to verify the validation system works
- Review existing entries in the dictionary for examples

## Pull Request Process

1. Ensure any install or build dependencies are removed before the end of the layer when doing a 
   build.
2. Update the README.md with details of changes to the interface. This includes new environment 
   variables, exposed ports, useful file locations and container parameters.
3. Increase the version numbers in any examples files and the VERSION.md to the new version that this
   Pull Request would represent. The versioning scheme we use is [SemVer](http://semver.org/).
4. **For adding new words, or amend any existing ones, please update the `dictionary.json` file. Add the words to the end of the list.**
5. **Ensure all pre-commit hooks pass** before submitting your pull request.

## Dictionary Contributions

When contributing to the Te Reo Māori dictionary:

1. **Add new entries at the end** of the dictionary array
2. **Use the next available index number**
3. **Include all required fields**: `index`, `word`, `meaning`
4. **Māori characters** are fully supported (ā, ē, ī, ō, ū, etc.)
5. **Optional fields** can be left as empty strings if not available:
   - `link`: External reference URL
   - `photo`: Image filename
   - `photo_attribution`: Photo credit/description

**Example new entry:**
```json
{
    "index": 366,
    "word": "whakatōhea",
    "meaning": "to make brave, encourage",
    "link": "",
    "photo": "",
    "photo_attribution": ""
}
```