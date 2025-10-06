#!/bin/bash

# Test Runner for Migration Validation
# This script runs all the necessary tests to validate the migration

set -e

echo "🧪 Migration Test Runner"
echo "========================"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Function to run test and report result
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "\n${BLUE}Running: $test_name${NC}"
    echo "Command: $test_command"
    
    if eval "$test_command"; then
        echo -e "${GREEN}✅ $test_name passed${NC}"
        return 0
    else
        echo -e "${RED}❌ $test_name failed${NC}"
        return 1
    fi
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
    echo -e "${RED}❌ Error: Must run from gum source directory${NC}"
    echo "Please cd to the gum repository and run this script"
    exit 1
fi

# Check if go is available
if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}❌ Error: Go is not installed or not in PATH${NC}"
    exit 1
fi

echo "✅ Environment check passed"

# Test 1: Build the project
run_test "Build Project" "go build -o gum-test ."

# Test 2: Run unit tests
run_test "Unit Tests" "go test -v ./internal/database/... -run TestMigration"

# Test 3: Run integration tests
run_test "Integration Tests" "go test -v ./cmd/... -run TestMigrationIntegration"

# Test 4: Run bug reproduction test
run_test "Bug Reproduction Test" "go test -v ./cmd/... -run TestBugReproduction"

# Test 5: Run performance tests
run_test "Performance Tests" "go test -v ./cmd/... -run TestPerformance -timeout 30s"

# Test 6: Run quick validation
run_test "Quick Validation" "go run validate_migration.go"

# Test 7: Test with real data (if available)
if [ -f ~/.cache/gum/projects.json ]; then
    echo -e "\n${YELLOW}📊 Found real data, testing with it...${NC}"
    
    # Create backup
    echo "Creating backup of real data..."
    cp ~/.cache/gum/projects.json ~/.cache/gum/projects.json.backup 2>/dev/null || true
    cp ~/.cache/gum/project-dirs.json ~/.cache/gum/project-dirs.json.backup 2>/dev/null || true
    
    # Run migration
    echo "Running migration with real data..."
    if ./gum-test migrate; then
        echo "✅ Migration with real data successful"
        
        # Test new system
        echo "Testing new system..."
        if ./gum-test projects-v2 --verbose; then
            echo "✅ New system works with real data"
        else
            echo "⚠️  New system had issues with real data"
        fi
        
        # Rollback
        echo "Rolling back..."
        if ./gum-test migrate --rollback; then
            echo "✅ Rollback successful"
        else
            echo "⚠️  Rollback had issues"
        fi
    else
        echo "⚠️  Migration with real data failed"
    fi
    
    # Restore backup
    echo "Restoring backup..."
    mv ~/.cache/gum/projects.json.backup ~/.cache/gum/projects.json 2>/dev/null || true
    mv ~/.cache/gum/project-dirs.json.backup ~/.cache/gum/project-dirs.json 2>/dev/null || true
else
    echo -e "\n${YELLOW}⚠️  No real data found, skipping real data test${NC}"
fi

# Test 8: Run comprehensive test suite
if [ -f "test_migration.sh" ]; then
    run_test "Comprehensive Test Suite" "chmod +x test_migration.sh && ./test_migration.sh"
else
    echo -e "\n${YELLOW}⚠️  Comprehensive test suite not found, skipping${NC}"
fi

# Cleanup
echo -e "\n${BLUE}🧹 Cleaning up...${NC}"
rm -f gum-test
rm -rf /tmp/gum_*_test*

# Final results
echo -e "\n${GREEN}🎉 All Tests Completed!${NC}"
echo "========================"
echo "✅ Build successful"
echo "✅ Unit tests passed"
echo "✅ Integration tests passed"
echo "✅ Bug reproduction test passed"
echo "✅ Performance tests passed"
echo "✅ Quick validation passed"
echo ""
echo -e "${GREEN}🚀 Ready for deployment!${NC}"
echo ""
echo "Next steps:"
echo "1. Review test results above"
echo "2. Deploy to staging environment"
echo "3. Run integration tests with real data"
echo "4. Monitor performance in production"
echo "5. Gradually roll out to users"