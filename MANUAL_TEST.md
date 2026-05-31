# Manual Test Plan - SQLite Migration

## Prerequisites

```bash
# Build the server
make build

# Verify database has 366 words
sqlite3 data/words.db "SELECT COUNT(*) FROM words WHERE day_index IS NOT NULL;"
# Expected: 366

# Check sample words exist
sqlite3 data/words.db "SELECT day_index, word FROM words WHERE day_index IN (1, 60, 366);"
```

---

## Test 1: Happy Path - Server Startup & DB Connection

**Objective**: Verify server starts successfully with DB connection pooling

```bash
# Start server (dry-run mode to avoid posting)
TEREOBOT_DRYRUN=true ./te-reo-bot start-server -address="localhost" -port="8080"
```

**Expected results:**
- ✅ Log: "Initializing database schema..."
- ✅ Log: "Server configuration loaded successfully" with `dryrun: true`
- ✅ Server starts on http://localhost:8080
- ✅ No errors about database connection
- ✅ No repeated schema initialization messages

**Edge case - Missing database:**
```bash
# Test with non-existent DB path
DB_PATH=/tmp/nonexistent.db ./te-reo-bot start-server -address="localhost" -port="8080"
```
Expected: Server should create database and initialize schema

---

## Test 2: Happy Path - Word Selection by Day

**Objective**: Verify O(1) word lookup by day index

```bash
# In another terminal, test day 1 (New Year)
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"day": 1}'
```

**Expected response:**
```json
{
  "word": "ngā mihi o te tau hou",
  "meaning": "happy new year, greetings for the new year",
  "day_index": 1,
  "link": "...",
  "photo": "...",
  "photo_attribution": "..."
}
```

**Verify in logs:**
- ✅ No "Failed to open database" errors
- ✅ No repeated DB initialization
- ✅ Request processed successfully

---

## Test 3: Edge Case - Invalid Day Index

**Objective**: Verify error handling for out-of-range days

### Test 3a: Day 0 (invalid)
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"day": 0}'
```

**Expected**: HTTP 400 or 404 with error message

### Test 3b: Day 367 (invalid)
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"day": 367}'
```

**Expected**: HTTP 400 or 404 with error message

### Test 3c: Negative day
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"day": -5}'
```

**Expected**: HTTP 400 with error message

---

## Test 4: Edge Case - Boundary Days

**Objective**: Verify first and last days work correctly

### Test 4a: Day 1 (first day)
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"day": 1}'
```

**Expected**: Word "ngā mihi o te tau hou"

### Test 4b: Day 366 (last day - leap year)
```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"day": 366}'
```

**Expected**: Word "Hōngongoi" (July)

---

## Test 5: Happy Path - Word Selection by Index

**Objective**: Verify selection by index parameter

```bash
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"index": 60}'
```

**Expected**: Word at index 60 ("Papa" - Flat)

---

## Test 6: Performance - Connection Pooling

**Objective**: Verify DB connection is reused across requests

```bash
# Send multiple rapid requests
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/messages \
    -H "Content-Type: application/json" \
    -H "X-Api-Key: YOUR_API_KEY" \
    -d "{\"day\": $i}" &
done
wait
```

**Expected behavior:**
- ✅ All 10 requests succeed
- ✅ Server logs show NO repeated "Initializing database schema"
- ✅ Server logs show NO "Failed to open database" errors
- ✅ Fast response times (< 100ms each)

**Check logs for:**
- ❌ Should NOT see: Multiple "Initializing database schema" messages
- ❌ Should NOT see: "database is locked" errors
- ✅ Should see: 10 successful "Processing message request" logs

---

## Test 7: Edge Case - Healthcheck Endpoint

**Objective**: Verify healthcheck doesn't depend on DB

```bash
curl http://localhost:8080/healthcheck
```

**Expected**: HTTP 200 (even if DB has issues)

---

## Test 8: Edge Case - Missing/Corrupt Database

**Objective**: Verify graceful handling of DB issues

### Test 8a: Database file permissions
```bash
# Make DB read-only
chmod 444 data/words.db

# Restart server
./te-reo-bot start-server -address="localhost" -port="8080"
```

**Expected**: Server starts in read-only mode (acceptable for this use case)

```bash
# Restore permissions
chmod 644 data/words.db
```

### Test 8b: Corrupt database
```bash
# Backup first!
cp data/words.db data/words.db.backup

# Corrupt the database
echo "corrupt data" > data/words.db

# Try to start server
./te-reo-bot start-server -address="localhost" -port="8080"
```

**Expected**: 
- Server logs error about corrupted database
- Server exits or returns error on startup

```bash
# Restore
mv data/words.db.backup data/words.db
```

---

## Test 9: Edge Case - Empty Response Fields

**Objective**: Verify handling of words with optional fields

```bash
# Check for words with NULL optional fields
sqlite3 data/words.db "SELECT day_index, word, link, photo FROM words WHERE link IS NULL OR photo IS NULL LIMIT 3;"

# Request one of those words
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: YOUR_API_KEY" \
  -d '{"day": <DAY_INDEX_FROM_ABOVE>}'
```

**Expected**: Response includes empty/null fields gracefully (no crash)

---

## Test 10: Regression - No Auth Header

**Objective**: Verify authentication still works

```bash
# Request without API key
curl -X POST http://localhost:8080/api/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"day": 1}'
```

**Expected**: HTTP 401 Unauthorized

---

## Success Criteria Checklist

- [ ] Server starts successfully with DB connection pooling
- [ ] Database schema auto-initializes once on startup
- [ ] Valid days (1-366) return correct words
- [ ] Invalid days (0, 367, negative) return proper errors
- [ ] Boundary days (1, 366) work correctly
- [ ] Index-based selection works
- [ ] Multiple rapid requests succeed (connection pooling)
- [ ] No per-request DB connection logs
- [ ] Response time < 100ms for word selection
- [ ] Healthcheck endpoint works independently
- [ ] Authentication still enforced
- [ ] Graceful handling of DB permission issues

---

## Cleanup

```bash
# Stop the server (Ctrl+C)

# Verify database integrity after tests
sqlite3 data/words.db "PRAGMA integrity_check;"
# Expected: ok

# Check word count unchanged
sqlite3 data/words.db "SELECT COUNT(*) FROM words WHERE day_index IS NOT NULL;"
# Expected: 366
```
