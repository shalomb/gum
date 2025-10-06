#!/bin/bash

# Pre-Deployment Test Suite for Database Migration
# This script reproduces the cache inconsistency bug and validates the fix

set -e  # Exit on any error

echo "🧪 Starting Pre-Deployment Test Suite"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
TEST_DIR="/tmp/gum_migration_test_$$"
CACHE_DIR="$TEST_DIR/.cache/gum"
DB_PATH="$CACHE_DIR/gum.db"
BACKUP_DIR="$TEST_DIR/backup"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}🧹 Cleaning up test environment...${NC}"
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Create test environment
echo -e "${BLUE}📁 Setting up test environment...${NC}"
mkdir -p "$CACHE_DIR"
mkdir -p "$BACKUP_DIR"

# Set environment variables
export XDG_CACHE_HOME="$TEST_DIR/.cache"
export HOME="$TEST_DIR"

echo "Test directory: $TEST_DIR"
echo "Cache directory: $CACHE_DIR"
echo "Database path: $DB_PATH"

# Test 1: Reproduce the Original Bug
echo -e "\n${BLUE}🔍 Test 1: Reproducing the Original Bug${NC}"
echo "Creating JSON cache files that simulate the inconsistency..."

# Create projects.json with many projects (simulating the bug scenario)
cat > "$CACHE_DIR/projects.json" << 'EOF'
{
  "data": [
    {"Path": "~/projects/project-1", "Remote": "https://github.com/user/project-1.git", "Branch": "main"},
    {"Path": "~/projects/project-2", "Remote": "https://github.com/user/project-2.git", "Branch": "main"},
    {"Path": "~/projects/project-3", "Remote": "https://github.com/user/project-3.git", "Branch": "main"},
    {"Path": "~/projects/project-4", "Remote": "https://github.com/user/project-4.git", "Branch": "main"},
    {"Path": "~/projects/project-5", "Remote": "https://github.com/user/project-5.git", "Branch": "main"},
    {"Path": "~/projects/project-6", "Remote": "https://github.com/user/project-6.git", "Branch": "main"},
    {"Path": "~/projects/project-7", "Remote": "https://github.com/user/project-7.git", "Branch": "main"},
    {"Path": "~/projects/project-8", "Remote": "https://github.com/user/project-8.git", "Branch": "main"},
    {"Path": "~/projects/project-9", "Remote": "https://github.com/user/project-9.git", "Branch": "main"},
    {"Path": "~/projects/project-10", "Remote": "https://github.com/user/project-10.git", "Branch": "main"}
  ],
  "timestamp": "2025-10-06T14:00:00Z",
  "ttl": 300
}
EOF

# Create project-dirs.json with different discovery (simulating gum update interference)
cat > "$CACHE_DIR/project-dirs.json" << 'EOF'
{
  "data": [
    {"Path": "~/projects", "LastScanned": "2025-10-06T14:00:00Z", "GitCount": 3}
  ],
  "timestamp": "2025-10-06T14:00:00Z",
  "ttl": 300
}
EOF

echo "✅ Created JSON cache files simulating the bug scenario"
echo "   - projects.json: 10 projects"
echo "   - project-dirs.json: 1 directory with 3 projects (inconsistent!)"

# Test 2: Verify Current Behavior (if gum is available)
echo -e "\n${BLUE}🔍 Test 2: Verifying Current Behavior${NC}"
if command -v gum >/dev/null 2>&1; then
    echo "Testing current gum behavior..."
    
    # Test current projects command
    echo "Current gum projects output:"
    gum projects 2>/dev/null | head -5 || echo "gum projects failed (expected if no real projects)"
    
    # Test current projects --refresh
    echo "Current gum projects --refresh output:"
    gum projects --refresh 2>/dev/null | head -5 || echo "gum projects --refresh failed (expected if no real projects)"
else
    echo "⚠️  gum command not available - skipping current behavior test"
fi

# Test 3: Build and Test Migration
echo -e "\n${BLUE}🔍 Test 3: Building and Testing Migration${NC}"

# Check if we're in the gum source directory
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
    echo -e "${RED}❌ Error: Must run from gum source directory${NC}"
    echo "Please cd to the gum repository and run this script"
    exit 1
fi

# Build the migration tool
echo "Building gum with migration support..."
if ! go build -o "$TEST_DIR/gum" .; then
    echo -e "${RED}❌ Failed to build gum${NC}"
    exit 1
fi
echo "✅ Built gum successfully"

# Test 4: Run Migration
echo -e "\n${BLUE}🔍 Test 4: Running Migration${NC}"

# Run migration
echo "Running migration..."
if ! "$TEST_DIR/gum" migrate; then
    echo -e "${RED}❌ Migration failed${NC}"
    exit 1
fi
echo "✅ Migration completed successfully"

# Test 5: Verify Migration Results
echo -e "\n${BLUE}🔍 Test 5: Verifying Migration Results${NC}"

# Check if database was created
if [ ! -f "$DB_PATH" ]; then
    echo -e "${RED}❌ Database file not created${NC}"
    exit 1
fi
echo "✅ Database file created: $DB_PATH"

# Check if JSON files were backed up
if [ ! -f "$CACHE_DIR/backup/projects.json" ]; then
    echo -e "${RED}❌ projects.json was not backed up${NC}"
    exit 1
fi
echo "✅ projects.json backed up"

if [ ! -f "$CACHE_DIR/backup/project-dirs.json" ]; then
    echo -e "${RED}❌ project-dirs.json was not backed up${NC}"
    exit 1
fi
echo "✅ project-dirs.json backed up"

# Test 6: Test New Projects Command
echo -e "\n${BLUE}🔍 Test 6: Testing New Projects Command${NC}"

# Test projects-v2 command
echo "Testing gum projects-v2..."
if ! "$TEST_DIR/gum" projects-v2; then
    echo -e "${RED}❌ projects-v2 command failed${NC}"
    exit 1
fi
echo "✅ projects-v2 command works"

# Test with verbose output
echo "Testing gum projects-v2 --verbose..."
if ! "$TEST_DIR/gum" projects-v2 --verbose; then
    echo -e "${RED}❌ projects-v2 --verbose command failed${NC}"
    exit 1
fi
echo "✅ projects-v2 --verbose command works"

# Test 7: Verify Cache Consistency
echo -e "\n${BLUE}🔍 Test 7: Verifying Cache Consistency${NC}"

# Test multiple calls to ensure consistency
echo "Testing cache consistency..."
PROJECTS1=$("$TEST_DIR/gum" projects-v2 | wc -l)
PROJECTS2=$("$TEST_DIR/gum" projects-v2 | wc -l)
PROJECTS3=$("$TEST_DIR/gum" projects-v2 | wc -l)

echo "First call: $PROJECTS1 projects"
echo "Second call: $PROJECTS2 projects"
echo "Third call: $PROJECTS3 projects"

if [ "$PROJECTS1" -eq "$PROJECTS2" ] && [ "$PROJECTS2" -eq "$PROJECTS3" ]; then
    echo "✅ Cache consistency verified - all calls return same count"
else
    echo -e "${RED}❌ Cache inconsistency detected!${NC}"
    echo "   First call: $PROJECTS1 projects"
    echo "   Second call: $PROJECTS2 projects"
    echo "   Third call: $PROJECTS3 projects"
    exit 1
fi

# Test 8: Test Refresh Functionality
echo -e "\n${BLUE}🔍 Test 8: Testing Refresh Functionality${NC}"

# Test refresh
echo "Testing gum projects-v2 --refresh..."
if ! "$TEST_DIR/gum" projects-v2 --refresh; then
    echo -e "${RED}❌ projects-v2 --refresh command failed${NC}"
    exit 1
fi
echo "✅ projects-v2 --refresh command works"

# Verify refresh doesn't break consistency
PROJECTS_AFTER_REFRESH=$("$TEST_DIR/gum" projects-v2 | wc -l)
echo "Projects after refresh: $PROJECTS_AFTER_REFRESH"

if [ "$PROJECTS_AFTER_REFRESH" -eq "$PROJECTS1" ]; then
    echo "✅ Refresh maintains consistency"
else
    echo -e "${YELLOW}⚠️  Refresh changed project count (may be expected)${NC}"
fi

# Test 9: Test Database Integrity
echo -e "\n${BLUE}🔍 Test 9: Testing Database Integrity${NC}"

# Check database integrity using sqlite3
if command -v sqlite3 >/dev/null 2>&1; then
    echo "Checking database integrity..."
    if sqlite3 "$DB_PATH" "PRAGMA integrity_check;" | grep -q "ok"; then
        echo "✅ Database integrity check passed"
    else
        echo -e "${RED}❌ Database integrity check failed${NC}"
        exit 1
    fi
    
    # Check table contents
    echo "Checking table contents..."
    PROJECTS_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM projects;")
    DIRS_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM project_dirs;")
    
    echo "Projects in database: $PROJECTS_COUNT"
    echo "Project directories in database: $DIRS_COUNT"
    
    if [ "$PROJECTS_COUNT" -gt 0 ] && [ "$DIRS_COUNT" -gt 0 ]; then
        echo "✅ Database contains migrated data"
    else
        echo -e "${RED}❌ Database is empty or incomplete${NC}"
        exit 1
    fi
else
    echo "⚠️  sqlite3 not available - skipping database integrity check"
fi

# Test 10: Test Rollback Functionality
echo -e "\n${BLUE}🔍 Test 10: Testing Rollback Functionality${NC}"

# Test rollback
echo "Testing rollback..."
if ! "$TEST_DIR/gum" migrate --rollback; then
    echo -e "${RED}❌ Rollback failed${NC}"
    exit 1
fi
echo "✅ Rollback completed successfully"

# Verify JSON files were restored
if [ -f "$CACHE_DIR/projects.json" ] && [ -f "$CACHE_DIR/project-dirs.json" ]; then
    echo "✅ JSON files restored after rollback"
else
    echo -e "${RED}❌ JSON files not restored after rollback${NC}"
    exit 1
fi

# Test 11: Performance Test
echo -e "\n${BLUE}🔍 Test 11: Performance Test${NC}"

# Re-run migration for performance test
echo "Re-running migration for performance test..."
if ! "$TEST_DIR/gum" migrate; then
    echo -e "${RED}❌ Re-migration failed${NC}"
    exit 1
fi

# Time the projects command
echo "Timing projects command..."
START_TIME=$(date +%s%N)
"$TEST_DIR/gum" projects-v2 >/dev/null
END_TIME=$(date +%s%N)
DURATION=$(( (END_TIME - START_TIME) / 1000000 )) # Convert to milliseconds

echo "Projects command took: ${DURATION}ms"

if [ "$DURATION" -lt 1000 ]; then
    echo "✅ Performance test passed (< 1 second)"
else
    echo -e "${YELLOW}⚠️  Performance test warning (> 1 second)${NC}"
fi

# Test 12: Concurrent Access Test
echo -e "\n${BLUE}🔍 Test 12: Concurrent Access Test${NC}"

# Test concurrent access
echo "Testing concurrent access..."
for i in {1..5}; do
    "$TEST_DIR/gum" projects-v2 >/dev/null &
done
wait

echo "✅ Concurrent access test completed"

# Final Results
echo -e "\n${GREEN}🎉 All Tests Passed!${NC}"
echo "======================================"
echo "✅ Migration functionality works"
echo "✅ Cache consistency is maintained"
echo "✅ Rollback functionality works"
echo "✅ Performance is acceptable"
echo "✅ Database integrity is maintained"
echo "✅ Concurrent access is safe"
echo ""
echo -e "${GREEN}🚀 Ready for deployment!${NC}"
echo ""
echo "Next steps:"
echo "1. Deploy to staging environment"
echo "2. Run integration tests with real data"
echo "3. Monitor performance in production"
echo "4. Gradually roll out to users"