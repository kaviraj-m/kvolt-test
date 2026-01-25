#!/bin/bash

BASE_URL="http://localhost:8080"
FAIL=0

echo "Starting API Verification..."

# 1. GET Users (List)
echo -n "Checking GET /users... "
HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}\n" "${BASE_URL}/users")
if [ "$HTTP_CODE" -eq 200 ]; then echo "OK ($HTTP_CODE)"; else echo "FAIL ($HTTP_CODE)"; FAIL=1; fi

# 2. POST User (Create) - Use a new ID to avoid conflict
echo -n "Checking POST /users... "
HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}\n" -d '{"name":"New User", "email":"new@test.com"}' "${BASE_URL}/users")
if [ "$HTTP_CODE" -eq 201 ]; then echo "OK ($HTTP_CODE)"; else echo "FAIL ($HTTP_CODE)"; FAIL=1; fi

# 3. GET User (Retrieve)
echo -n "Checking GET /users/1... "
HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}\n" "${BASE_URL}/users/1")
if [ "$HTTP_CODE" -eq 200 ]; then echo "OK ($HTTP_CODE)"; else echo "FAIL ($HTTP_CODE)"; FAIL=1; fi

# 4. PUT User (Update)
echo -n "Checking PUT /users/1... "
HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}\n" -X PUT -d '{"name":"Updated", "email":"up@test.com"}' "${BASE_URL}/users/1")
if [ "$HTTP_CODE" -eq 200 ]; then echo "OK ($HTTP_CODE)"; else echo "FAIL ($HTTP_CODE)"; FAIL=1; fi

# 5. POST Post (Create Resource)
echo -n "Checking POST /posts... "
HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}\n" -d '{"title":"Verification Post"}' "${BASE_URL}/posts")
if [ "$HTTP_CODE" -eq 201 ]; then echo "OK ($HTTP_CODE)"; else echo "FAIL ($HTTP_CODE)"; FAIL=1; fi

# 6. GET Posts (List Resource)
echo -n "Checking GET /posts... "
HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}\n" "${BASE_URL}/posts")
if [ "$HTTP_CODE" -eq 200 ]; then echo "OK ($HTTP_CODE)"; else echo "FAIL ($HTTP_CODE)"; FAIL=1; fi

# 7. DELETE User (Delete) - Do this last
echo -n "Checking DELETE /users/1... "
HTTP_CODE=$(curl -o /dev/null -s -w "%{http_code}\n" -X DELETE "${BASE_URL}/users/1")
if [ "$HTTP_CODE" -eq 204 ]; then echo "OK ($HTTP_CODE)"; else echo "FAIL ($HTTP_CODE)"; FAIL=1; fi

echo "--------------------------------"
if [ $FAIL -eq 0 ]; then
    echo "✅ All tests passed!"
else
    echo "❌ Some tests failed!"
fi
