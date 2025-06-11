#!/bin/bash

ACCESS_TOKEN=$(curl -s -X POST "https://openapi.koreainvestment.com:9443/oauth2/tokenP" \
  -H "Content-Type: application/json; charset=UTF-8" \
  -d "{\"grant_type\":\"client_credentials\",\"appkey\":\"$KIS_APP_KEY\",\"appsecret\":\"$KIS_APP_SECRET\"}" | jq -r '.access_token')

MOCK_ACCESS_TOKEN=$(curl -s -X POST "https://openapivts.koreainvestment.com:29443/oauth2/tokenP" \
  -H "Content-Type: application/json; charset=UTF-8" \
  -d "{\"grant_type\":\"client_credentials\",\"appkey\":\"$KIS_MOCK_APP_KEY\",\"appsecret\":\"$KIS_MOCK_APP_SECRET\"}" | jq -r '.access_token')

export "KIS_ACCESS_TOKEN=$ACCESS_TOKEN"
export "KIS_MOCK_ACCESS_TOKEN=$MOCK_ACCESS_TOKEN"
