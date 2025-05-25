#!/bin/bash

LOGIN_URL="http://localhost:8080/login"
USERNAME="testuser"
PASSWORD="password123"

# 1. Логинимся и сохраняем cookie
COOKIE=$(curl -s -i -X POST "$LOGIN_URL" \
  -H "Content-Type: application/json" \
  -d "{\"login\":\"$USERNAME\", \"password\":\"$PASSWORD\"}" | \
  grep -i "Set-Cookie" | head -n 1 | cut -d':' -f2 | cut -d';' -f1 | xargs)

echo "Полученные куки: $COOKIE"

# 2. Тестируем создание пользователя (POST /users)
CREATE_USER_URL="http://localhost:8080/users"
sqlmap -u "$CREATE_USER_URL" --method POST \
  --data='{
    "user": {
      "login": "newuser", 
      "password": "password123",
      "phone": "+79001234567", 
      "email": "newuser@example.com"
    },
    "profile": {
      "firstName": "Иван", 
      "lastName": "Иванов", 
      "isMale": true
    }
  }' \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"


# 3. Тестируем вход в систему (POST /users/login)
LOGIN_USER_URL="http://localhost:8080/users/login"
sqlmap -u "$LOGIN_USER_URL" --method POST \
  --data='{"login": "testuser", "password": "password123"}' \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 4. Тестируем выход из системы (POST /users/logout)
LOGOUT_USER_URL="http://localhost:8080/users/logout"
sqlmap -u "$LOGOUT_USER_URL" --method POST \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 5. Тестируем удаление пользователя (DELETE /users/{id})
DELETE_USER_URL="http://localhost:8080/users/1"
sqlmap -u "$DELETE_USER_URL" --method DELETE \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 6. Тестируем получение профиля пользователя (GET /profiles/{id})
GET_PROFILE_URL="http://localhost:8080/profiles/1"
sqlmap -u "$GET_PROFILE_URL" --method GET \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 7. Тестируем обновление профиля (POST /profiles/update)
UPDATE_PROFILE_URL="http://localhost:8080/profiles/update"
sqlmap -u "$UPDATE_PROFILE_URL" --method POST \
  --data='{
    "profileId": 6,
    "firstName": "Егор",
    "lastName": "Заславский",
    "isMale": true,
    "birthday": "1990-05-15T00:00:00Z",
    "description": "Программист",
    "location": "США",
    "interests": ["Коммунизм", "Капитан", "МГТУ"],
    "likedBy": [10, 9],
    "preferences": [
      {"preference_description": "bodyType", "preference_value": "Беларусь"},
      {"preference_description": "hairColor", "preference_value": "Беларусь"},
      {"preference_description": "eyeColor", "preference_value": "Беларусь"},
      {"preference_description": "tattoo", "preference_value": "Беларусь"}
    ],
    "photos": ["/anatolini.jpg"]
  }' \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 8. Тестируем создание чата (POST /chats/create)
CREATE_CHAT_URL="http://localhost:8080/chats/create"
sqlmap -u "$CREATE_CHAT_URL" --method POST \
  --data='{"firstID":1,"secondID":2}' \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 9. Тестируем удаление чата (DELETE /chats/delete)
DELETE_CHAT_URL="http://localhost:8080/chats/delete"
sqlmap -u "$DELETE_CHAT_URL" --method DELETE \
  --data='{"firstID":1,"secondID":2}' \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 10. Тестируем создание жалобы (POST /complaints/create)
CREATE_COMPLAINT_URL="http://localhost:8080/complaints/create"
sqlmap -u "$CREATE_COMPLAINT_URL" --method POST \
  --data='{
    "Complaint_type": "spam",
    "Complaint_text": "This is spam",
    "Complaint_on": "user123"
  }' \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"

# 11. Тестируем получение жалоб (GET /complaints/get)
GET_COMPLAINTS_URL="http://localhost:8080/complaints/get"
sqlmap -u "$GET_COMPLAINTS_URL" --method GET \
  --headers="Content-Type: application/json" \
  --cookie="$COOKIE" \
  --risk=3 --level=5 --batch --ignore-code="500,400,403"
