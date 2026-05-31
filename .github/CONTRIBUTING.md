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

### Dictionary Management

The word dictionary is stored in SQLite database (`data/words.db`) as the single source of truth.

#### Database Schema

Words are stored with the following structure:
- **id**: Auto-incrementing primary key
- **day_index**: Integer (1-366) for word-of-the-day selection
- **word**: Māori word text
- **meaning**: English translation
- **link**: Reference URL (optional)
- **photo**: Image filename (optional)
- **photo_attribution**: Photo credit (optional)
- **created_at**: Timestamp
- **updated_at**: Timestamp
- **is_active**: Boolean flag

#### Adding/Updating Words

Words should be managed directly in the SQLite database:

```bash
# Access the database
sqlite3 data/words.db

# Example: Add a new word
INSERT INTO words (day_index, word, meaning, link, photo, photo_attribution)
VALUES (367, 'aroha', 'love, compassion', 'https://maoridictionary.co.nz/...', 'aroha.jpg', 'Photo credit');

# Example: Update existing word
UPDATE words SET meaning = 'updated meaning' WHERE word = 'kia ora';
```

#### Validation

Database constraints ensure data integrity:
- **Unique day_index**: Each day (1-366) can only have one word
- **Required fields**: word and meaning are mandatory
- **Schema validation**: Auto-runs on server startup

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
4. **For adding new words or amending existing ones, please update the SQLite database (`data/words.db`). See the Dictionary Management section above for details.**
5. **Ensure all pre-commit hooks pass** before submitting your pull request.

## Dictionary Contributions

When contributing to the Te Reo Māori dictionary:

1. **Directly modify the SQLite database** (`data/words.db`)
2. **Use the next available day_index** (currently 1-366 for leap year coverage)
3. **Include all required fields**: `word`, `meaning`
4. **Māori characters** are fully supported (ā, ē, ī, ō, ū, etc.)
5. **Optional fields** can be left as empty strings if not available:
   - `link`: External reference URL
   - `photo`: Image filename
   - `photo_attribution`: Photo credit/description

**Example SQL insert:**
```sql
INSERT INTO words (day_index, word, meaning, link, photo, photo_attribution)
VALUES (367, 'whakatōhea', 'to make brave, encourage', '', '', '');
```

**Note**: The database is tracked in Git as a binary file via `.gitattributes`.