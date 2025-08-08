#!/bin/bash
set -e

echo "🧪 dockrune Smoke Test Suite"
echo "============================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test function
run_test() {
    local test_name="$1"
    local test_cmd="$2"
    
    TESTS_RUN=$((TESTS_RUN + 1))
    echo -n "Testing $test_name... "
    
    if eval "$test_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo "  Command: $test_cmd"
    fi
}

echo "1️⃣  Go Compilation Tests"
echo "------------------------"
run_test "Main binary compilation" "go build -o /tmp/dockrune-test ./cmd/dockrune"
run_test "Go module verification" "go mod verify"
run_test "Go formatting check" "test -z \$(gofmt -l .)"

echo ""
echo "2️⃣  Unit Tests"
echo "--------------"
run_test "Detector tests" "go test ./internal/detector"
run_test "Webhook tests" "go test ./internal/webhook"
run_test "Storage tests" "go test ./internal/storage"
run_test "Config tests" "go test ./internal/config"
run_test "Models tests" "go test ./internal/models"

echo ""
echo "3️⃣  Integration Tests"
echo "--------------------"
run_test "All tests with coverage" "go test ./... -cover"
run_test "Race condition check" "go test -race ./internal/deployer"

echo ""
echo "4️⃣  Binary Functionality"
echo "------------------------"
if [ -f /tmp/dockrune-test ]; then
    run_test "Binary help command" "/tmp/dockrune-test --help"
    run_test "Binary version command" "/tmp/dockrune-test --version"
    rm -f /tmp/dockrune-test
fi

echo ""
echo "5️⃣  Docker Build Test"
echo "--------------------"
run_test "Dockerfile syntax" "docker build --no-cache -t dockrune-test:smoke -f Dockerfile . --target builder"

echo ""
echo "6️⃣  Project Structure"
echo "--------------------"
run_test "Required files exist" "test -f go.mod && test -f Dockerfile && test -f docker-compose.yml"
run_test "README exists" "test -f README.md || test -f README_COMPLETE.md"
run_test "Environment example" "test -f .env.example"

echo ""
echo "7️⃣  Security Checks"
echo "------------------"
# Check for hardcoded secrets
run_test "No hardcoded secrets" "! grep -r 'password\\|secret\\|token' --include='*.go' . | grep -v test | grep -v example | grep '=\"'"
run_test "No exposed ports in code" "! grep -r ':8080\\|:3000\\|:8000' --include='*.go' . | grep -v test | grep -v config"

echo ""
echo "📊 Test Results"
echo "=============="
echo -e "Tests Run:    ${YELLOW}$TESTS_RUN${NC}"
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ All smoke tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed. Please review the output above.${NC}"
    exit 1
fi