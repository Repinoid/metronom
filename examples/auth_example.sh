#!/bin/bash
# Example script demonstrating the user authentication API

# Server URL
SERVER="http://localhost:8080"

echo "=== User Authentication Example ==="
echo ""

# 1. Register a new user
echo "1. Registering a new user..."
REGISTER_RESPONSE=$(curl -s -X POST "$SERVER/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demouser",
    "password": "demopass123"
  }')
echo "Response: $REGISTER_RESPONSE"
echo ""

# 2. Login with the user
echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$SERVER/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demouser",
    "password": "demopass123"
  }')
echo "Response: $LOGIN_RESPONSE"
echo ""

# Extract token from response (requires jq)
if command -v jq &> /dev/null; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
    echo "Token: $TOKEN"
    echo ""
    
    # 3. Use the token to access protected resources (example)
    echo "3. Accessing protected resource with token..."
    echo "curl -H \"Authorization: Bearer $TOKEN\" $SERVER/protected-endpoint"
    echo ""
else
    echo "Install 'jq' to automatically extract the token"
    echo "You can manually extract the token from the JSON response above"
fi

echo "=== Example Complete ==="
echo ""
echo "To use the token in subsequent requests:"
echo "curl -H \"Authorization: Bearer YOUR_TOKEN\" $SERVER/your-endpoint"
