#!/bin/bash

# Test script for GitHub sync functionality

set -e

echo "🧪 Testing GitHub Sync Functionality"
echo "=================================="

# Test 1: Check if gum is built
echo "Test 1: Checking if gum is built..."
if [ ! -f "./gum" ]; then
    echo "❌ gum binary not found. Run 'make build' first."
    exit 1
fi
echo "✅ gum binary found"

# Test 2: Check GitHub authentication
echo "Test 2: Checking GitHub authentication..."
if ! gh auth status > /dev/null 2>&1; then
    echo "❌ GitHub CLI not authenticated. Run 'gh auth login' first."
    exit 1
fi
echo "✅ GitHub CLI authenticated"

# Test 3: Test dry-run sync
echo "Test 3: Testing dry-run sync..."
if ! ./gum sync --dry-run --type incremental; then
    echo "❌ Dry-run sync failed"
    exit 1
fi
echo "✅ Dry-run sync successful"

# Test 4: Check database creation
echo "Test 4: Checking database creation..."
CACHE_DIR="$HOME/.cache/gum"
if [ ! -d "$CACHE_DIR" ]; then
    echo "❌ Cache directory not created: $CACHE_DIR"
    exit 1
fi
echo "✅ Cache directory created: $CACHE_DIR"

# Test 5: Run a small sync test
echo "Test 5: Running small sync test..."
if ! ./gum sync --type incremental; then
    echo "❌ Incremental sync failed"
    exit 1
fi
echo "✅ Incremental sync successful"

# Test 6: Check database contents
echo "Test 6: Checking database contents..."
DB_PATH="$CACHE_DIR/gum.db"
if [ ! -f "$DB_PATH" ]; then
    echo "❌ Database file not created: $DB_PATH"
    exit 1
fi

# Check if tables exist
if ! sqlite3 "$DB_PATH" ".tables" | grep -q "github_metadata"; then
    echo "❌ github_metadata table not found"
    exit 1
fi
echo "✅ Database tables created"

# Test 7: Check sync status
echo "Test 7: Checking sync status..."
SYNC_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM github_sync_status;" | tr -d ' ')
if [ "$SYNC_COUNT" -eq 0 ]; then
    echo "❌ No sync status records found"
    exit 1
fi
echo "✅ Sync status records found: $SYNC_COUNT"

# Test 8: Check metadata records
echo "Test 8: Checking metadata records..."
METADATA_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM github_metadata;" | tr -d ' ')
echo "✅ Metadata records: $METADATA_COUNT"

if [ "$METADATA_COUNT" -gt 0 ]; then
    echo "Sample metadata:"
    sqlite3 "$DB_PATH" "SELECT full_name, language, star_count FROM github_metadata LIMIT 3;"
fi

# Test 9: Test crontab generation
echo "Test 9: Testing crontab generation..."
if ! ./gum --crontab | grep -q "gum sync"; then
    echo "❌ Sync command not found in crontab output"
    exit 1
fi
echo "✅ Sync command found in crontab output"

echo ""
echo "🎉 All tests passed! GitHub sync is working correctly."
echo ""
echo "Next steps:"
echo "1. Run 'gum sync --type full' to sync all repositories"
echo "2. Add 'gum --crontab' output to your crontab for daily automation"
echo "3. Check database with: sqlite3 ~/.cache/gum/gum.db 'SELECT * FROM github_metadata LIMIT 10;'"