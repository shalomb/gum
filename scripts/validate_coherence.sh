#!/bin/bash

# Coherence Validation Script
# Ensures perfect alignment between Solution Intent → BDD → TDD

set -e

echo "🎯 Gum Coherence Validation"
echo "=========================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS") echo -e "${GREEN}✅ $message${NC}" ;;
        "FAIL") echo -e "${RED}❌ $message${NC}" ;;
        "WARN") echo -e "${YELLOW}⚠️  $message${NC}" ;;
        "INFO") echo -e "${BLUE}ℹ️  $message${NC}" ;;
    esac
}

# Function to check if file exists
check_file() {
    local file=$1
    if [[ -f "$file" ]]; then
        print_status "PASS" "File exists: $file"
        return 0
    else
        print_status "FAIL" "File missing: $file"
        return 1
    fi
}

# Function to count BDD scenarios
count_bdd_scenarios() {
    local feature_file=$1
    local count=$(grep -c "## Scenario:" "$feature_file" 2>/dev/null || echo "0")
    echo "$count"
}

# Function to count TDD tests
count_tdd_tests() {
    local test_file=$1
    local count=$(grep -c "func Test" "$test_file" 2>/dev/null || echo "0")
    echo "$count"
}

echo "📋 Phase 1: Document Structure Validation"
echo "=========================================="

# Check that all required documents exist
print_status "INFO" "Checking document structure..."

required_docs=(
    "docs/solution-intent.md"
    "docs/test-traceability-matrix.md"
    "docs/bdd/project-discovery.feature"
    "docs/bdd/directory-management.feature"
    "docs/bdd/performance.feature"
    "docs/bdd/configuration.feature"
    "docs/bdd/concurrency.feature"
    "docs/bdd/database-migration.feature"
    "docs/bdd/github-sync.feature"
    "docs/bdd/locate-integration.feature"
)

missing_docs=0
for doc in "${required_docs[@]}"; do
    if ! check_file "$doc"; then
        ((missing_docs++))
    fi
done

if [[ $missing_docs -eq 0 ]]; then
    print_status "PASS" "All required documents present"
else
    print_status "FAIL" "$missing_docs documents missing"
fi

echo ""
echo "🧪 Phase 2: BDD Scenario Analysis"
echo "================================="

# Count BDD scenarios
total_bdd_scenarios=0
bdd_files=("docs/bdd"/*.feature)

for feature_file in "${bdd_files[@]}"; do
    if [[ -f "$feature_file" ]]; then
        count=$(count_bdd_scenarios "$feature_file")
        total_bdd_scenarios=$((total_bdd_scenarios + count))
        print_status "INFO" "$(basename "$feature_file"): $count scenarios"
    fi
done

print_status "INFO" "Total BDD scenarios: $total_bdd_scenarios"

echo ""
echo "🔬 Phase 3: TDD Test Analysis"
echo "============================="

# Count TDD tests
total_tdd_tests=0
test_files=(
    "cmd/dirs_test.go"
    "cmd/search_test.go"
    "cmd/performance_test.go"
    "cmd/frecency_test.go"
    "integration_test.go"
    "internal/database/database_test.go"
    "internal/database/concurrency_test.go"
    "internal/cache/cache_test.go"
    "internal/locate/locate_test.go"
)

for test_file in "${test_files[@]}"; do
    if [[ -f "$test_file" ]]; then
        count=$(count_tdd_tests "$test_file")
        total_tdd_tests=$((total_tdd_tests + count))
        print_status "INFO" "$(basename "$test_file"): $count tests"
    else
        print_status "WARN" "Missing test file: $test_file"
    fi
done

print_status "INFO" "Total TDD tests: $total_tdd_tests"

echo ""
echo "📊 Phase 4: Coverage Analysis"
echo "============================="

# Calculate coverage ratio
if [[ $total_bdd_scenarios -gt 0 ]]; then
    coverage_ratio=$((total_tdd_tests * 100 / total_bdd_scenarios))
    print_status "INFO" "BDD to TDD coverage ratio: $coverage_ratio%"
    
    if [[ $coverage_ratio -ge 90 ]]; then
        print_status "PASS" "Excellent coverage ratio"
    elif [[ $coverage_ratio -ge 70 ]]; then
        print_status "WARN" "Good coverage ratio, but could be improved"
    else
        print_status "FAIL" "Poor coverage ratio, needs improvement"
    fi
else
    print_status "FAIL" "No BDD scenarios found"
fi

echo ""
echo "🚀 Phase 5: Performance Validation"
echo "=================================="

# Test performance requirements
print_status "INFO" "Testing performance requirements..."

# Test project discovery speed
start_time=$(date +%s%N)
if gum projects > /dev/null 2>&1; then
    end_time=$(date +%s%N)
    duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    if [[ $duration_ms -lt 200 ]]; then
        print_status "PASS" "Project discovery: ${duration_ms}ms (< 200ms required)"
    else
        print_status "FAIL" "Project discovery: ${duration_ms}ms (>= 200ms, too slow)"
    fi
else
    print_status "FAIL" "Project discovery command failed"
fi

# Test cache response speed
start_time=$(date +%s%N)
if gum projects > /dev/null 2>&1; then
    end_time=$(date +%s%N)
    duration_ms=$(( (end_time - start_time) / 1000000 ))
    
    if [[ $duration_ms -lt 100 ]]; then
        print_status "PASS" "Cache response: ${duration_ms}ms (< 100ms required)"
    else
        print_status "WARN" "Cache response: ${duration_ms}ms (>= 100ms, could be faster)"
    fi
else
    print_status "FAIL" "Cache response command failed"
fi

echo ""
echo "🔍 Phase 6: Coherence Validation"
echo "==============================="

# Run coherence tests if they exist
if [[ -f "test_coherence.go" ]]; then
    print_status "INFO" "Running coherence validation tests..."
    
    if go test -v -run TestCoherence > /dev/null 2>&1; then
        print_status "PASS" "Coherence tests passed"
    else
        print_status "FAIL" "Coherence tests failed"
    fi
else
    print_status "WARN" "Coherence test file not found"
fi

echo ""
echo "📈 Phase 7: Summary Report"
echo "=========================="

# Generate summary report
echo "## Coherence Validation Summary"
echo ""
echo "### Document Structure"
echo "- Required documents: ${#required_docs[@]}"
echo "- Missing documents: $missing_docs"
echo ""

echo "### Test Coverage"
echo "- BDD scenarios: $total_bdd_scenarios"
echo "- TDD tests: $total_tdd_tests"
echo "- Coverage ratio: $coverage_ratio%"
echo ""

echo "### Performance"
echo "- Project discovery: ${duration_ms}ms"
echo "- Cache response: ${duration_ms}ms"
echo ""

# Overall status
if [[ $missing_docs -eq 0 && $coverage_ratio -ge 70 ]]; then
    print_status "PASS" "Overall coherence validation: PASSED"
    echo ""
    echo "🎉 Congratulations! Your solution demonstrates excellent coherence"
    echo "   between Solution Intent → BDD → TDD."
    exit 0
else
    print_status "FAIL" "Overall coherence validation: FAILED"
    echo ""
    echo "🔧 Action required: Please address the issues above to improve coherence."
    exit 1
fi