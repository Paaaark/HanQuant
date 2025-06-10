#!/bin/bash

ACCESS_TOKEN=$(curl -s -X POST "https://openapivts.koreainvestment.com:29443/oauth2/tokenP" \
  -H "Content-Type: application/json; charset=UTF-8" \
  -d "{\"grant_type\":\"client_credentials\",\"appkey\":\"$TradingApiKey\",\"appsecret\":\"$TradingSecretKey\"}" | jq -r '.access_token')

export "KIS_ACCESS_TOKEN=$ACCESS_TOKEN"
